package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/pipeline"
	"github.com/gentleman-programming/gentle-ai/v2/internal/planner"
	"github.com/gentleman-programming/gentle-ai/v2/internal/system"
	"github.com/gentleman-programming/gentle-ai/v2/internal/verify"
)

func TestRoutedSkillInventoryRoutesGlobalSkillsAndAlignsLifecycle(t *testing.T) {
	home := t.TempDir()
	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentOpenCode},
		Components: []model.ComponentID{model.ComponentSDD, model.ComponentSkills},
		Skills:     []model.SkillID{model.SkillGoTesting},
	}
	resolved := planner.ResolvedPlan{Agents: selection.Agents, OrderedComponents: selection.Components}
	adapters := resolveAdapters(selection.Agents)
	inventory, err := buildRoutedSkillInventory(home, ScopeGlobal, selection, adapters)
	if err != nil {
		t.Fatal(err)
	}
	if inventory == nil || !inventory.hasAgent(model.AgentOpenCode) {
		t.Fatal("global OpenCode inventory was not routed")
	}

	sharedSkill := filepath.Join(home, ".agents", "skills", "go-testing", "SKILL.md")
	nativeSkill := filepath.Join(home, ".config", "opencode", "skills", "sdd-init", "SKILL.md")
	if !containsPath(inventory.paths(model.AgentOpenCode, model.ComponentSkills), sharedSkill) {
		t.Fatalf("ordinary paths do not contain shared destination %q", sharedSkill)
	}
	if !containsPath(inventory.paths(model.AgentOpenCode, model.ComponentSDD), nativeSkill) {
		t.Fatalf("SDD paths do not contain native destination %q", nativeSkill)
	}

	for _, component := range selection.Components {
		step := componentApplyStep{
			component:      component,
			homeDir:        home,
			workspaceDir:   home,
			scope:          ScopeGlobal,
			agents:         selection.Agents,
			selection:      selection,
			skillInventory: inventory,
		}
		if err := step.Run(); err != nil {
			t.Fatalf("apply %q: %v", component, err)
		}
	}

	for _, path := range []string{sharedSkill, nativeSkill, filepath.Join(home, ".config", "opencode", "skills", "_shared", "sdd-phase-common.md")} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("routed materialization missing %q: %v", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "opencode", "skills", "go-testing", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("native duplicate for shared skill exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".agents", "skills", "sdd-init", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("shared duplicate for required skill exists: %v", err)
	}

	targets, err := backupTargetsWithSkillInventory(home, home, ScopeGlobal, selection, resolved, inventory)
	if err != nil {
		t.Fatal(err)
	}
	if !containsPath(targets, sharedSkill) || !containsPath(targets, nativeSkill) {
		t.Fatalf("routed backup targets missing effective destinations: %v", targets)
	}
	if containsPath(targets, filepath.Join(home, ".config", "opencode", "skills", "go-testing", "SKILL.md")) {
		t.Fatal("routed backup targets include legacy native duplicate for shared skill")
	}

	report := runPostApplyVerification(postApplyVerificationInput{
		HomeDir:        home,
		WorkspaceDir:   home,
		Scope:          ScopeGlobal,
		Selection:      selection,
		Resolved:       resolved,
		SkillInventory: inventory,
	})
	assertVerificationPathStatus(t, report, "verify:file:"+sharedSkill, verify.CheckStatusPassed)
	assertVerificationPathAbsent(t, report, "verify:file:"+filepath.Join(home, ".config", "opencode", "skills", "go-testing", "SKILL.md"))
	if err := os.Remove(sharedSkill); err != nil {
		t.Fatal(err)
	}
	report = runPostApplyVerification(postApplyVerificationInput{
		HomeDir:        home,
		WorkspaceDir:   home,
		Scope:          ScopeGlobal,
		Selection:      selection,
		Resolved:       resolved,
		SkillInventory: inventory,
	})
	assertVerificationPathStatus(t, report, "verify:file:"+sharedSkill, verify.CheckStatusFailed)
}

func TestRoutedSkillInventoryDrivesSyncAndSkipsCompatibilityRefresh(t *testing.T) {
	home := t.TempDir()
	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentOpenCode},
		Components: []model.ComponentID{model.ComponentSDD, model.ComponentSkills},
		Skills:     []model.SkillID{model.SkillGoTesting},
	}
	inventory, err := buildRoutedSkillInventory(home, ScopeGlobal, selection, resolveAdapters(selection.Agents))
	if err != nil {
		t.Fatal(err)
	}
	var changed []string
	for _, component := range selection.Components {
		step := componentSyncStep{
			component:      component,
			homeDir:        home,
			workspaceDir:   home,
			agents:         selection.Agents,
			selection:      selection,
			changedFiles:   &changed,
			skillInventory: inventory,
		}
		if err := step.Run(); err != nil {
			t.Fatalf("sync %q: %v", component, err)
		}
	}
	sharedSkill := filepath.Join(home, ".agents", "skills", "go-testing", "SKILL.md")
	nativeSkill := filepath.Join(home, ".config", "opencode", "skills", "sdd-init", "SKILL.md")
	if !containsPath(changed, sharedSkill) || !containsPath(changed, nativeSkill) {
		t.Fatalf("sync changed files do not contain routed destinations: %v", changed)
	}
	if _, err := os.Stat(filepath.Join(home, ".config", "opencode", "skills", "go-testing", "SKILL.md")); !os.IsNotExist(err) {
		t.Fatalf("sync created native duplicate for shared skill: %v", err)
	}
	report := runPostSyncVerificationWithSkillInventory(home, home, selection, inventory)
	assertVerificationPathStatus(t, report, "verify:sync:file:"+sharedSkill, verify.CheckStatusPassed)
	assertVerificationPathAbsent(t, report, "verify:sync:file:"+filepath.Join(home, ".config", "opencode", "skills", "go-testing", "SKILL.md"))

	changed = nil
	for _, component := range selection.Components {
		step := componentSyncStep{
			component:      component,
			homeDir:        home,
			workspaceDir:   home,
			agents:         selection.Agents,
			selection:      selection,
			changedFiles:   &changed,
			skillInventory: inventory,
		}
		if err := step.Run(); err != nil {
			t.Fatalf("repeat sync %q: %v", component, err)
		}
	}
	if len(changed) != 0 {
		t.Fatalf("repeat routed sync reported changed files: %v", changed)
	}

	resolved := planner.ResolvedPlan{Agents: selection.Agents, OrderedComponents: selection.Components}
	installRuntime, err := newInstallRuntime(home, ScopeGlobal, ChannelStable, selection, resolved, system.PlatformProfile{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(installRuntime.state.cleanupCompatibilityTransaction)
	if stagePlanHasStep(installRuntime.stagePlan(), "component:compatibility-skills-refresh") {
		t.Fatal("global routed install still schedules compatibility refresh")
	}
	syncRuntime, err := newSyncRuntime(home, selection)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(syncRuntime.state.cleanupCompatibilityTransaction)
	if stagePlanHasStep(syncRuntime.stagePlan(), "sync:compatibility-skills-refresh") {
		t.Fatal("global routed sync still schedules compatibility refresh")
	}
}

func TestRoutedSkillInventoryFailsBeforeInstallMutation(t *testing.T) {
	home := t.TempDir()
	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentPi},
		Components: []model.ComponentID{model.ComponentSDD},
	}
	resolved := planner.ResolvedPlan{Agents: selection.Agents, OrderedComponents: selection.Components}
	if _, err := newInstallRuntime(home, ScopeGlobal, ChannelStable, selection, resolved, system.PlatformProfile{}); err == nil || !strings.Contains(err.Error(), "no reliable destination") {
		t.Fatalf("Pi install preflight error = %v, want no reliable destination", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".gentle-ai")); !os.IsNotExist(err) {
		t.Fatalf("failed Pi preflight mutated backup root: %v", err)
	}
}

func stagePlanHasStep(plan pipeline.StagePlan, id string) bool {
	for _, step := range plan.Apply {
		if step.ID() == id {
			return true
		}
	}
	return false
}

func assertVerificationPathStatus(t *testing.T, report verify.Report, id string, want verify.CheckStatus) {
	t.Helper()
	for _, result := range report.Checks {
		if result.ID == id {
			if result.Status != want {
				t.Fatalf("verification %q = %q (%s), want %q", id, result.Status, result.Error, want)
			}
			return
		}
	}
	t.Fatalf("verification does not contain %q: %#v", id, report.Checks)
}

func assertVerificationPathAbsent(t *testing.T, report verify.Report, id string) {
	t.Helper()
	for _, result := range report.Checks {
		if result.ID == id {
			t.Fatalf("verification unexpectedly contains legacy path %q", id)
		}
	}
}
