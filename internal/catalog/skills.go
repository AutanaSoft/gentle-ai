package catalog

import "github.com/gentleman-programming/gentle-ai/v2/internal/model"

// SkillPlacement identifies the canonical runtime placement semantics of a
// managed skill. It deliberately describes policy rather than a filesystem
// destination; adapters provide the runtime roots separately.
type SkillPlacement string

const (
	SkillPlacementAgentSkillsShared   SkillPlacement = "agent-skills-shared"
	SkillPlacementAgentSkillsRequired SkillPlacement = "agent-skills-required"
)

func (p SkillPlacement) Valid() bool {
	return p == SkillPlacementAgentSkillsShared || p == SkillPlacementAgentSkillsRequired
}

type Skill struct {
	ID        model.SkillID
	Name      string
	Category  string
	Priority  string
	Placement SkillPlacement
}

var mvpSkills = []Skill{
	// SDD skills
	{ID: model.SkillSDDInit, Name: "sdd-init", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},

	{ID: model.SkillSDDApply, Name: "sdd-apply", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillSDDVerify, Name: "sdd-verify", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillSDDExplore, Name: "sdd-explore", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillSDDPropose, Name: "sdd-propose", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillSDDSpec, Name: "sdd-spec", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillSDDDesign, Name: "sdd-design", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillSDDTasks, Name: "sdd-tasks", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillSDDArchive, Name: "sdd-archive", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillSDDOnboard, Name: "sdd-onboard", Category: "sdd", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	// Foundation skills
	{ID: model.SkillGoTesting, Name: "go-testing", Category: "testing", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillGentleAIBench, Name: "gentle-ai-bench", Category: "testing", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillCreator, Name: "skill-creator", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillImprover, Name: "skill-improver", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillJudgmentDay, Name: "judgment-day", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillBranchPR, Name: "branch-pr", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillIssueCreation, Name: "issue-creation", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillSkillRegistry, Name: "skill-registry", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	// Sustainable review skills
	{ID: model.SkillChainedPR, Name: "chained-pr", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsRequired},
	{ID: model.SkillCognitiveDoc, Name: "cognitive-doc-design", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillCommentWriter, Name: "comment-writer", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillWorkUnitCommits, Name: "work-unit-commits", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillRDDDefectWorkflow, Name: "rdd-defect-workflow", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
	{ID: model.SkillSystemicIssueTriage, Name: "systemic-issue-triage", Category: "workflow", Priority: "p0", Placement: SkillPlacementAgentSkillsShared},
}

func MVPSkills() []Skill {
	skills := make([]Skill, len(mvpSkills))
	copy(skills, mvpSkills)
	return skills
}

// SkillByID returns the canonical catalog record for id. Callers that need a
// skill's placement must resolve it here rather than maintaining a parallel
// SkillID-to-placement map.
func SkillByID(id model.SkillID) (Skill, bool) {
	for _, skill := range mvpSkills {
		if skill.ID == id {
			return skill, true
		}
	}
	return Skill{}, false
}
