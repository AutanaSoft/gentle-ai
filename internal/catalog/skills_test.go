package catalog

import (
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/components/skills"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

// TestMVPSkillsCoverAllPresetSkills ensures every skill that presets.go would
// install is also registered in the catalog's mvpSkills allowlist. This
// prevents a future addition to sddSkills or foundationSkills from being
// silently rejected by normalizeSkills in cli/validate.go.
func TestMVPSkillsCoverAllPresetSkills(t *testing.T) {
	catalogSet := make(map[model.SkillID]bool)
	for _, s := range MVPSkills() {
		catalogSet[s.ID] = true
	}

	presetSkills := skills.AllSkillIDs()
	for _, id := range presetSkills {
		if !catalogSet[id] {
			t.Errorf("skill %q is in presets but missing from catalog mvpSkills", id)
		}
	}
}

// TestMVPSkillsNoDuplicates ensures no skill is listed twice in mvpSkills.
func TestMVPSkillsNoDuplicates(t *testing.T) {
	seen := make(map[model.SkillID]bool)
	for _, s := range MVPSkills() {
		if seen[s.ID] {
			t.Errorf("duplicate skill %q in mvpSkills", s.ID)
		}
		seen[s.ID] = true
	}
}

func TestMVPSkillsHaveCompletePlacementInventory(t *testing.T) {
	want := map[model.SkillID]SkillPlacement{
		model.SkillSDDInit:             SkillPlacementAgentSkillsRequired,
		model.SkillSDDExplore:          SkillPlacementAgentSkillsRequired,
		model.SkillSDDPropose:          SkillPlacementAgentSkillsRequired,
		model.SkillSDDSpec:             SkillPlacementAgentSkillsRequired,
		model.SkillSDDDesign:           SkillPlacementAgentSkillsRequired,
		model.SkillSDDTasks:            SkillPlacementAgentSkillsRequired,
		model.SkillSDDApply:            SkillPlacementAgentSkillsRequired,
		model.SkillSDDVerify:           SkillPlacementAgentSkillsRequired,
		model.SkillSDDArchive:          SkillPlacementAgentSkillsRequired,
		model.SkillSDDOnboard:          SkillPlacementAgentSkillsRequired,
		model.SkillJudgmentDay:         SkillPlacementAgentSkillsRequired,
		model.SkillSkillRegistry:       SkillPlacementAgentSkillsRequired,
		model.SkillChainedPR:           SkillPlacementAgentSkillsRequired,
		model.SkillGoTesting:           SkillPlacementAgentSkillsShared,
		model.SkillGentleAIBench:       SkillPlacementAgentSkillsShared,
		model.SkillCreator:             SkillPlacementAgentSkillsShared,
		model.SkillImprover:            SkillPlacementAgentSkillsShared,
		model.SkillBranchPR:            SkillPlacementAgentSkillsShared,
		model.SkillIssueCreation:       SkillPlacementAgentSkillsShared,
		model.SkillCognitiveDoc:        SkillPlacementAgentSkillsShared,
		model.SkillCommentWriter:       SkillPlacementAgentSkillsShared,
		model.SkillWorkUnitCommits:     SkillPlacementAgentSkillsShared,
		model.SkillRDDDefectWorkflow:   SkillPlacementAgentSkillsShared,
		model.SkillSystemicIssueTriage: SkillPlacementAgentSkillsShared,
	}

	skills := MVPSkills()
	if len(skills) != len(want) {
		t.Fatalf("MVPSkills() has %d entries, want %d", len(skills), len(want))
	}

	for _, skill := range skills {
		if !skill.Placement.Valid() {
			t.Errorf("skill %q has invalid or zero placement %q", skill.ID, skill.Placement)
		}
		if got, ok := want[skill.ID]; !ok {
			t.Errorf("unexpected catalog skill %q", skill.ID)
		} else if skill.Placement != got {
			t.Errorf("skill %q placement = %q, want %q", skill.ID, skill.Placement, got)
		}
	}
}

func TestMVPSkillsIncludeRequestedBundledSkillsWithCanonicalNames(t *testing.T) {
	required := map[model.SkillID]string{
		model.SkillCreator:             "skill-creator",
		model.SkillSkillRegistry:       "skill-registry",
		model.SkillCognitiveDoc:        "cognitive-doc-design",
		model.SkillCommentWriter:       "comment-writer",
		model.SkillJudgmentDay:         "judgment-day",
		model.SkillSDDInit:             "sdd-init",
		model.SkillImprover:            "skill-improver",
		model.SkillRDDDefectWorkflow:   "rdd-defect-workflow",
		model.SkillSystemicIssueTriage: "systemic-issue-triage",
		model.SkillGentleAIBench:       "gentle-ai-bench",
	}

	found := make(map[model.SkillID]string)
	for _, skill := range MVPSkills() {
		found[skill.ID] = skill.Name
		if skill.Name == "judgement-day" {
			t.Fatalf("catalog uses non-canonical spelling %q; want judgment-day", skill.Name)
		}
	}

	for id, wantName := range required {
		name, ok := found[id]
		if !ok {
			t.Fatalf("MVPSkills() missing requested bundled skill %q", id)
		}
		if name != wantName {
			t.Fatalf("MVPSkills() name for %q = %q, want %q", id, name, wantName)
		}
	}
}
