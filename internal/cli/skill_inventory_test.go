package cli

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

func TestRoutedSkillInventoryPreservesWorkspaceAndOpenClawLegacyPaths(t *testing.T) {
	home := t.TempDir()
	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentOpenCode},
		Components: []model.ComponentID{model.ComponentSkills},
		Skills:     []model.SkillID{model.SkillGoTesting},
	}
	inventory, err := buildRoutedSkillInventory(home, ScopeWorkspace, selection, selection.Components, resolveAdapters(selection.Agents))
	if err != nil {
		t.Fatal(err)
	}
	if inventory != nil {
		t.Fatal("workspace scope manufactured a routed global inventory")
	}

	selection.Agents = []model.AgentID{model.AgentOpenClaw}
	inventory, err = buildRoutedSkillInventory(home, ScopeGlobal, selection, selection.Components, resolveAdapters(selection.Agents))
	if err != nil {
		t.Fatal(err)
	}
	if inventory != nil {
		t.Fatal("workspace-first OpenClaw received a global routed inventory")
	}
}

func TestRoutedSkillInventoryUsesCanonicalLookupAndKimiDeclaredRoot(t *testing.T) {
	home := t.TempDir()
	unknown := model.Selection{
		Agents:     []model.AgentID{model.AgentOpenCode},
		Components: []model.ComponentID{model.ComponentSkills},
		Skills:     []model.SkillID{"unknown-skill"},
	}
	if _, err := buildRoutedSkillInventory(home, ScopeGlobal, unknown, unknown.Components, resolveAdapters(unknown.Agents)); err == nil || !strings.Contains(err.Error(), "canonical catalog") {
		t.Fatalf("unknown skill error = %v, want canonical catalog failure", err)
	}

	kimi := model.Selection{
		Agents:     []model.AgentID{model.AgentKimi},
		Components: []model.ComponentID{model.ComponentSkills},
		Skills:     []model.SkillID{model.SkillGoTesting},
	}
	inventory, err := buildRoutedSkillInventory(home, ScopeGlobal, kimi, kimi.Components, resolveAdapters(kimi.Agents))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(home, ".config", "agents", "skills", "go-testing", "SKILL.md")
	if !containsPath(inventory.paths(model.AgentKimi, model.ComponentSkills), path) {
		t.Fatalf("Kimi shared skill did not use first declared shared root: %v", inventory.paths(model.AgentKimi, model.ComponentSkills))
	}
}

func TestRoutedSkillInventoryFailsClosedWhenPiHasNoReliableDestination(t *testing.T) {
	home := t.TempDir()
	selection := model.Selection{
		Agents:     []model.AgentID{model.AgentPi},
		Components: []model.ComponentID{model.ComponentSDD},
	}
	if _, err := buildRoutedSkillInventory(home, ScopeGlobal, selection, selection.Components, resolveAdapters(selection.Agents)); err == nil || !strings.Contains(err.Error(), "no reliable destination") {
		t.Fatalf("Pi routed inventory error = %v, want no reliable destination", err)
	}
}
