package agents

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/antigravity"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/claude"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/codex"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/cursor"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/gemini"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/hermes"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/kilocode"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/kimi"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/kiro"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/openclaw"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/opencode"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/pi"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/qwen"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/trae"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/vscode"
	"github.com/gentleman-programming/gentle-ai/v2/internal/agents/windsurf"
	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
)

func TestSkillDiscoveryRootsFallback(t *testing.T) {
	tests := []struct {
		name    string
		adapter Adapter
		want    []SkillDiscoveryRoot
	}{
		{
			name:    "adapter without capability uses native skills directory",
			adapter: skillDiscoveryFallbackAdapter{supportsSkills: true, skillsDir: "/home/test/.native/skills"},
			want: []SkillDiscoveryRoot{{
				Scope: SkillDiscoveryNativeGlobal,
				Path:  "/home/test/.native/skills",
			}},
		},
		{
			name:    "adapter without generic skills support has no roots",
			adapter: skillDiscoveryFallbackAdapter{supportsSkills: false, skillsDir: "/home/test/.native/skills"},
		},
		{
			name: "disabled generic skills reject optional provider roots",
			adapter: skillDiscoveryUnsupportedProviderAdapter{
				Adapter: pi.NewAdapter(),
				roots: SkillDiscoveryCapabilities{Roots: []SkillDiscoveryRoot{{
					Scope: SkillDiscoverySharedGlobal,
					Path:  "/home/test/.agents/skills",
				}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SkillDiscoveryRoots(tt.adapter, "/home/test"); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("SkillDiscoveryRoots() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestSkillDiscoveryAdapterMatrix(t *testing.T) {
	home := t.TempDir()
	shared := filepath.Join(home, ".agents", "skills")
	native := func(path ...string) string { return filepath.Join(append([]string{home}, path...)...) }

	tests := []struct {
		name         string
		adapter      Adapter
		wantAgent    model.AgentID
		wantProvider bool
		wantRoots    []SkillDiscoveryRoot
		wantSkills   bool
	}{
		{
			name:       "Codex uses native fallback only",
			adapter:    codex.NewAdapter(),
			wantAgent:  model.AgentCodex,
			wantSkills: true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".codex", "skills")},
			},
		},
		{
			name:         "OpenCode declares native then shared roots",
			adapter:      opencode.NewAdapter(),
			wantAgent:    model.AgentOpenCode,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".config", "opencode", "skills")},
				{Scope: SkillDiscoverySharedGlobal, Path: shared},
			},
		},
		{
			name:         "Gemini CLI declares shared then native roots",
			adapter:      gemini.NewAdapter(),
			wantAgent:    model.AgentGeminiCLI,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoverySharedGlobal, Path: shared},
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".gemini", "skills")},
			},
		},
		{
			name:       "Cursor uses native fallback only",
			adapter:    cursor.NewAdapter(),
			wantAgent:  model.AgentCursor,
			wantSkills: true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".cursor", "skills")},
			},
		},
		{
			name:       "VS Code Copilot uses native fallback only",
			adapter:    vscode.NewAdapter(),
			wantAgent:  model.AgentVSCodeCopilot,
			wantSkills: true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".copilot", "skills")},
			},
		},
		{
			name:       "Windsurf uses native fallback only",
			adapter:    windsurf.NewAdapter(),
			wantAgent:  model.AgentWindsurf,
			wantSkills: true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".codeium", "windsurf", "skills")},
			},
		},
		{
			name:         "Kimi declares native then shared roots",
			adapter:      kimi.NewAdapter(),
			wantAgent:    model.AgentKimi,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".kimi", "skills")},
				{Scope: SkillDiscoverySharedGlobal, Path: native(".config", "agents", "skills")},
				{Scope: SkillDiscoverySharedGlobal, Path: shared},
			},
		},
		{
			name:         "OpenClaw declares shared then native roots",
			adapter:      openclaw.NewAdapter(),
			wantAgent:    model.AgentOpenClaw,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoverySharedGlobal, Path: shared},
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".openclaw", "skills")},
			},
		},
		{
			name:       "Claude Code uses native fallback only",
			adapter:    claude.NewAdapter(),
			wantAgent:  model.AgentClaudeCode,
			wantSkills: true,
			wantRoots:  []SkillDiscoveryRoot{{Scope: SkillDiscoveryNativeGlobal, Path: native(".claude", "skills")}},
		},
		{
			name:       "Kiro IDE uses native fallback only",
			adapter:    kiro.NewAdapter(),
			wantAgent:  model.AgentKiroIDE,
			wantSkills: true,
			wantRoots:  []SkillDiscoveryRoot{{Scope: SkillDiscoveryNativeGlobal, Path: native(".kiro", "skills")}},
		},
		{
			name:       "Qwen Code uses native fallback only",
			adapter:    qwen.NewAdapter(),
			wantAgent:  model.AgentQwenCode,
			wantSkills: true,
			wantRoots:  []SkillDiscoveryRoot{{Scope: SkillDiscoveryNativeGlobal, Path: native(".qwen", "skills")}},
		},
		{
			name:       "Antigravity uses native fallback only",
			adapter:    antigravity.NewAdapter(),
			wantAgent:  model.AgentAntigravity,
			wantSkills: true,
			wantRoots:  []SkillDiscoveryRoot{{Scope: SkillDiscoveryNativeGlobal, Path: native(".gemini", "antigravity-cli", "skills")}},
		},
		{
			name:       "Hermes uses native fallback only",
			adapter:    hermes.NewAdapter(),
			wantAgent:  model.AgentHermes,
			wantSkills: true,
			wantRoots:  []SkillDiscoveryRoot{{Scope: SkillDiscoveryNativeGlobal, Path: native(".hermes", "skills")}},
		},
		{
			name:       "Kilo Code uses native fallback only",
			adapter:    kilocode.NewAdapter(),
			wantAgent:  model.AgentKilocode,
			wantSkills: true,
			wantRoots:  []SkillDiscoveryRoot{{Scope: SkillDiscoveryNativeGlobal, Path: native(".config", "kilo", "skills")}},
		},
		{
			name:       "Trae IDE uses native fallback only",
			adapter:    trae.NewAdapter(),
			wantAgent:  model.AgentTrae,
			wantSkills: true,
			wantRoots:  []SkillDiscoveryRoot{{Scope: SkillDiscoveryNativeGlobal, Path: native(".trae", "skills")}},
		},
		{
			name:       "Pi has no generic route",
			adapter:    pi.NewAdapter(),
			wantAgent:  model.AgentPi,
			wantSkills: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.adapter.Agent(); got != tt.wantAgent {
				t.Fatalf("Agent() = %q, want %q", got, tt.wantAgent)
			}
			if got := tt.adapter.SupportsSkills(); got != tt.wantSkills {
				t.Fatalf("SupportsSkills() = %t, want %t", got, tt.wantSkills)
			}
			_, gotProvider := tt.adapter.(SkillDiscoveryProvider)
			if gotProvider != tt.wantProvider {
				t.Fatalf("implements SkillDiscoveryProvider = %t, want %t", gotProvider, tt.wantProvider)
			}
			if got := SkillDiscoveryRoots(tt.adapter, home); !reflect.DeepEqual(got, tt.wantRoots) {
				t.Fatalf("SkillDiscoveryRoots() = %#v, want %#v", got, tt.wantRoots)
			}
		})
	}
}

type skillDiscoveryFallbackAdapter struct {
	Adapter
	supportsSkills bool
	skillsDir      string
}

func (a skillDiscoveryFallbackAdapter) SupportsSkills() bool { return a.supportsSkills }
func (a skillDiscoveryFallbackAdapter) SkillsDir(string) string {
	return a.skillsDir
}

type skillDiscoveryUnsupportedProviderAdapter struct {
	Adapter
	roots SkillDiscoveryCapabilities
}

func (a skillDiscoveryUnsupportedProviderAdapter) SupportsSkills() bool { return false }

func (a skillDiscoveryUnsupportedProviderAdapter) SkillDiscovery(string) SkillDiscoveryCapabilities {
	return a.roots
}
