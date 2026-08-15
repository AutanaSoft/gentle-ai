package planner

import (
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents"
	"github.com/gentleman-programming/gentle-ai/v2/internal/catalog"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

func TestResolveSkillRoute(t *testing.T) {
	native := filepath.Join("/home/test", ".native", "skills")
	shared := filepath.Join("/home/test", ".agents", "skills")

	tests := []struct {
		name      string
		placement catalog.SkillPlacement
		roots     []agents.SkillDiscoveryRoot
		want      SkillRoute
	}{
		{
			name:      "shared prefers shared root even when native precedes it",
			placement: catalog.SkillPlacementAgentSkillsShared,
			roots: []agents.SkillDiscoveryRoot{
				{Scope: agents.SkillDiscoveryNativeGlobal, Path: native},
				{Scope: agents.SkillDiscoverySharedGlobal, Path: shared},
			},
			want: SkillRoute{Skill: model.SkillGoTesting, Placement: catalog.SkillPlacementAgentSkillsShared, Destination: filepath.Join(shared, string(model.SkillGoTesting)), Status: SkillRouteReady},
		},
		{
			name:      "shared falls back to native root",
			placement: catalog.SkillPlacementAgentSkillsShared,
			roots: []agents.SkillDiscoveryRoot{
				{Scope: agents.SkillDiscoveryNativeGlobal, Path: native},
			},
			want: SkillRoute{Skill: model.SkillGoTesting, Placement: catalog.SkillPlacementAgentSkillsShared, Destination: filepath.Join(native, string(model.SkillGoTesting)), Status: SkillRouteReady},
		},
		{
			name:      "required selects native first",
			placement: catalog.SkillPlacementAgentSkillsRequired,
			roots: []agents.SkillDiscoveryRoot{
				{Scope: agents.SkillDiscoveryNativeGlobal, Path: native},
				{Scope: agents.SkillDiscoverySharedGlobal, Path: shared},
			},
			want: SkillRoute{Skill: model.SkillSDDInit, Placement: catalog.SkillPlacementAgentSkillsRequired, Destination: filepath.Join(native, string(model.SkillSDDInit)), Status: SkillRouteReady},
		},
		{
			name:      "required selects shared first",
			placement: catalog.SkillPlacementAgentSkillsRequired,
			roots: []agents.SkillDiscoveryRoot{
				{Scope: agents.SkillDiscoverySharedGlobal, Path: shared},
				{Scope: agents.SkillDiscoveryNativeGlobal, Path: native},
			},
			want: SkillRoute{Skill: model.SkillSDDInit, Placement: catalog.SkillPlacementAgentSkillsRequired, Destination: filepath.Join(shared, string(model.SkillSDDInit)), Status: SkillRouteReady},
		},
		{
			name:      "no roots has no reliable destination",
			placement: catalog.SkillPlacementAgentSkillsShared,
			want:      SkillRoute{Skill: model.SkillGoTesting, Placement: catalog.SkillPlacementAgentSkillsShared, Status: SkillRouteNoReliableDestination},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			skill := catalog.Skill{ID: tt.want.Skill, Placement: tt.placement}
			if got := ResolveSkillRoute(skill, tt.roots); got != tt.want {
				t.Fatalf("ResolveSkillRoute() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
