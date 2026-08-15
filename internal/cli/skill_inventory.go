package cli

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents"
	"github.com/gentleman-programming/gentle-ai/v2/internal/catalog"
	"github.com/gentleman-programming/gentle-ai/v2/internal/components/sdd"
	"github.com/gentleman-programming/gentle-ai/v2/internal/components/skills"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/planner"
)

// routedSkillInventory is the single pure-derived authority for global skill
// materialization. It is built before mutation and then shared by install,
// sync, backup, and verification.
//
// Workspace scope deliberately does not build one: adapter discovery roots are
// globally verified, not templates that may be rebased onto a project. OpenClaw
// is also left on its established workspace-first path until it has an explicit
// workspace discovery contract.
type routedSkillInventory struct {
	byAgent map[model.AgentID]routedAgentSkillInventory
}

type routedAgentSkillInventory struct {
	ordinary []routedSkillGroup
	sdd      []routedSkillGroup
}

type routedSkillGroup struct {
	root  string
	ids   []model.SkillID
	paths []string
}

type routedSkillInventoryBuilder struct {
	ordinary map[model.AgentID]map[string][]model.SkillID
	sdd      map[model.AgentID]map[string][]model.SkillID
}

type routedSkillInjectionResult struct {
	Changed bool
	Files   []string
}

func buildRoutedSkillInventory(homeDir string, scope InstallScope, selection model.Selection, activeComponents []model.ComponentID, adapters []agents.Adapter) (*routedSkillInventory, error) {
	if scope != ScopeGlobal || !selectionHasManagedSkills(activeComponents) {
		return nil, nil
	}
	if err := validateSelectedSkillIDs(selection); err != nil {
		return nil, err
	}
	ordinaryIDs := ordinarySkillIDs(selection, activeComponents)
	sddIDs := sddSkillIDs(activeComponents)
	for _, adapter := range adapters {
		if adapter.Agent() == model.AgentOpenClaw {
			continue
		}
		if !adapter.SupportsSkills() && (len(ordinaryIDs) > 0 || len(sddIDs) > 0) {
			return nil, fmt.Errorf("managed skills for agent %q have no reliable destination", adapter.Agent())
		}
	}

	builder := routedSkillInventoryBuilder{
		ordinary: make(map[model.AgentID]map[string][]model.SkillID),
		sdd:      make(map[model.AgentID]map[string][]model.SkillID),
	}
	hasRoutableAgent := false
	for _, adapter := range adapters {
		// OpenClaw resolves all component files under its configured workspace,
		// irrespective of InstallScope. Passing that directory to a global
		// SkillDiscoveryProvider would fabricate project roots from global
		// declarations, so retain its established workspace-first writer only.
		if adapter.Agent() == model.AgentOpenClaw {
			continue
		}
		if !adapter.SupportsSkills() {
			continue
		}
		hasRoutableAgent = true
		roots := agents.SkillDiscoveryRoots(adapter, homeDir)
		if err := builder.add(adapter.Agent(), ordinaryIDs, roots, false); err != nil {
			return nil, err
		}
		if err := builder.add(adapter.Agent(), sddIDs, roots, true); err != nil {
			return nil, err
		}
	}
	if !hasRoutableAgent {
		return nil, nil
	}

	inventory := &routedSkillInventory{byAgent: make(map[model.AgentID]routedAgentSkillInventory)}
	for _, adapter := range adapters {
		agent := adapter.Agent()
		ordinary, err := finalizeRoutedSkillGroups(builder.ordinary[agent], false)
		if err != nil {
			return nil, fmt.Errorf("build ordinary skill inventory for %q: %w", agent, err)
		}
		sddGroups, err := finalizeRoutedSkillGroups(builder.sdd[agent], true)
		if err != nil {
			return nil, fmt.Errorf("build SDD skill inventory for %q: %w", agent, err)
		}
		if len(ordinary) == 0 && len(sddGroups) == 0 {
			continue
		}
		inventory.byAgent[agent] = routedAgentSkillInventory{ordinary: ordinary, sdd: sddGroups}
	}
	if len(inventory.byAgent) == 0 {
		return nil, nil
	}
	return inventory, nil
}

func selectionHasManagedSkills(activeComponents []model.ComponentID) bool {
	return hasComponent(activeComponents, model.ComponentSkills) || hasComponent(activeComponents, model.ComponentSDD)
}

func validateSelectedSkillIDs(selection model.Selection) error {
	for _, id := range selectedSkillIDs(selection) {
		if _, ok := catalog.SkillByID(id); !ok {
			return fmt.Errorf("managed skill %q is absent from the canonical catalog", id)
		}
	}
	return nil
}

func ordinarySkillIDs(selection model.Selection, activeComponents []model.ComponentID) []model.SkillID {
	if !hasComponent(activeComponents, model.ComponentSkills) {
		return nil
	}
	ids := make([]model.SkillID, 0, len(selectedSkillIDs(selection)))
	seen := make(map[model.SkillID]struct{})
	for _, id := range selectedSkillIDs(selection) {
		if skills.IsSDDSkill(id) {
			continue
		}
		if hasComponent(activeComponents, model.ComponentSDD) && skills.IsSDDManagedSkill(id) {
			continue
		}
		if _, duplicate := seen[id]; duplicate {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

func sddSkillIDs(activeComponents []model.ComponentID) []model.SkillID {
	if !hasComponent(activeComponents, model.ComponentSDD) {
		return nil
	}
	return skills.SDDSkillIDs()
}

func (b routedSkillInventoryBuilder) add(agent model.AgentID, ids []model.SkillID, roots []agents.SkillDiscoveryRoot, sddOwned bool) error {
	var destinations map[model.AgentID]map[string][]model.SkillID
	if sddOwned {
		destinations = b.sdd
	} else {
		destinations = b.ordinary
	}
	if destinations[agent] == nil {
		destinations[agent] = make(map[string][]model.SkillID)
	}

	for _, id := range ids {
		skill, ok := catalog.SkillByID(id)
		if !ok {
			return fmt.Errorf("managed skill %q for agent %q is absent from the canonical catalog", id, agent)
		}
		route := planner.ResolveSkillRoute(skill, roots)
		if route.Status != planner.SkillRouteReady {
			return fmt.Errorf("managed skill %q for agent %q has no reliable destination", id, agent)
		}
		root := filepath.Dir(route.Destination)
		destinations[agent][root] = append(destinations[agent][root], id)
	}
	return nil
}

func finalizeRoutedSkillGroups(byRoot map[string][]model.SkillID, sddOwned bool) ([]routedSkillGroup, error) {
	roots := make([]string, 0, len(byRoot))
	for root := range byRoot {
		roots = append(roots, root)
	}
	sort.Strings(roots)
	groups := make([]routedSkillGroup, 0, len(roots))
	for _, root := range roots {
		ids := append([]model.SkillID(nil), byRoot[root]...)
		var (
			paths []string
			err   error
		)
		if sddOwned {
			paths, err = sdd.SkillDirectoryPathsForSkills(root, ids, "", true)
		} else {
			paths, err = skills.DirectoryPaths(root, ids, "")
		}
		if err != nil {
			return nil, err
		}
		groups = append(groups, routedSkillGroup{root: root, ids: ids, paths: paths})
	}
	return groups, nil
}

func (i *routedSkillInventory) hasAgent(agent model.AgentID) bool {
	if i == nil {
		return false
	}
	_, ok := i.byAgent[agent]
	return ok
}

func (i *routedSkillInventory) bypassesCompatibilityRefresh() bool {
	return i != nil && len(i.byAgent) > 0
}

func (i *routedSkillInventory) paths(agent model.AgentID, component model.ComponentID) []string {
	if i == nil {
		return nil
	}
	entry, ok := i.byAgent[agent]
	if !ok {
		return nil
	}
	var groups []routedSkillGroup
	switch component {
	case model.ComponentSkills:
		groups = entry.ordinary
	case model.ComponentSDD:
		groups = entry.sdd
	default:
		return nil
	}
	var paths []string
	for _, group := range groups {
		paths = append(paths, group.paths...)
	}
	return paths
}

func (i *routedSkillInventory) injectOrdinary(agent model.AgentID) (routedSkillInjectionResult, error) {
	if i == nil {
		return routedSkillInjectionResult{}, nil
	}
	entry, ok := i.byAgent[agent]
	if !ok {
		return routedSkillInjectionResult{}, nil
	}
	return injectRoutedOrdinaryGroups(entry.ordinary)
}

func (i *routedSkillInventory) injectSDD(agent model.AgentID, capability string) (routedSkillInjectionResult, error) {
	if i == nil {
		return routedSkillInjectionResult{}, nil
	}
	entry, ok := i.byAgent[agent]
	if !ok {
		return routedSkillInjectionResult{}, nil
	}
	return injectRoutedSDDGroups(entry.sdd, capability)
}

func injectRoutedOrdinaryGroups(groups []routedSkillGroup) (routedSkillInjectionResult, error) {
	result := routedSkillInjectionResult{}
	for _, group := range groups {
		injected, err := skills.InjectDirectoryWithCapability(group.root, group.ids, "")
		if err != nil {
			return routedSkillInjectionResult{}, err
		}
		result.Changed = result.Changed || injected.Changed
		result.Files = append(result.Files, injected.Files...)
	}
	return result, nil
}

func injectRoutedSDDGroups(groups []routedSkillGroup, capability string) (routedSkillInjectionResult, error) {
	result := routedSkillInjectionResult{}
	for _, group := range groups {
		injected, err := sdd.InjectSkillDirectoryForSkills(group.root, group.ids, capability, true)
		if err != nil {
			return routedSkillInjectionResult{}, err
		}
		result.Changed = result.Changed || injected.Changed
		result.Files = append(result.Files, injected.Files...)
	}
	return result, nil
}
