package planner

import (
	"path/filepath"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents"
	"github.com/gentleman-programming/gentle-ai/v2/internal/catalog"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

type SkillRouteStatus string

const (
	SkillRouteReady                 SkillRouteStatus = "ready"
	SkillRouteNoReliableDestination SkillRouteStatus = "no-reliable-destination"
)

// SkillRoute is a pure placement decision. It does not claim ownership or
// observe filesystem state; later routing stages can supply narrower roots
// before calling this resolver.
type SkillRoute struct {
	Skill       model.SkillID
	Placement   catalog.SkillPlacement
	Destination string
	Status      SkillRouteStatus
}

// ResolveSkillRoute combines catalog placement with adapter-declared discovery
// roots. The result is only a destination decision; it performs no I/O.
func ResolveSkillRoute(skill catalog.Skill, roots []agents.SkillDiscoveryRoot) SkillRoute {
	route := SkillRoute{
		Skill:     skill.ID,
		Placement: skill.Placement,
		Status:    SkillRouteNoReliableDestination,
	}

	root, ok := skillRouteRoot(skill.Placement, roots)
	if !ok {
		return route
	}

	route.Destination = filepath.Join(root.Path, string(skill.ID))
	route.Status = SkillRouteReady
	return route
}

func skillRouteRoot(placement catalog.SkillPlacement, roots []agents.SkillDiscoveryRoot) (agents.SkillDiscoveryRoot, bool) {
	switch placement {
	case catalog.SkillPlacementAgentSkillsShared:
		if root, ok := firstSkillDiscoveryRoot(roots, agents.SkillDiscoverySharedGlobal); ok {
			return root, true
		}
		return firstSkillDiscoveryRoot(roots, agents.SkillDiscoveryNativeGlobal)
	case catalog.SkillPlacementAgentSkillsRequired:
		for _, root := range roots {
			if root.Path != "" {
				return root, true
			}
		}
	}
	return agents.SkillDiscoveryRoot{}, false
}

func firstSkillDiscoveryRoot(roots []agents.SkillDiscoveryRoot, scope agents.SkillDiscoveryScope) (agents.SkillDiscoveryRoot, bool) {
	for _, root := range roots {
		if root.Scope == scope && root.Path != "" {
			return root, true
		}
	}
	return agents.SkillDiscoveryRoot{}, false
}
