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
			name:         "Codex declares native then shared roots",
			adapter:      codex.NewAdapter(),
			wantAgent:    model.AgentCodex,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".codex", "skills")},
				{Scope: SkillDiscoverySharedGlobal, Path: shared},
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
			name:         "Gemini CLI declares native then shared roots",
			adapter:      gemini.NewAdapter(),
			wantAgent:    model.AgentGeminiCLI,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".gemini", "skills")},
				{Scope: SkillDiscoverySharedGlobal, Path: shared},
			},
		},
		{
			name:         "Cursor declares native then shared roots",
			adapter:      cursor.NewAdapter(),
			wantAgent:    model.AgentCursor,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".cursor", "skills")},
				{Scope: SkillDiscoverySharedGlobal, Path: shared},
			},
		},
		{
			name:         "VS Code Copilot declares native then shared roots",
			adapter:      vscode.NewAdapter(),
			wantAgent:    model.AgentVSCodeCopilot,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".copilot", "skills")},
				{Scope: SkillDiscoverySharedGlobal, Path: shared},
			},
		},
		{
			name:         "Windsurf declares native then shared roots",
			adapter:      windsurf.NewAdapter(),
			wantAgent:    model.AgentWindsurf,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".codeium", "windsurf", "skills")},
				{Scope: SkillDiscoverySharedGlobal, Path: shared},
			},
		},
		{
			name:         "Kimi declares configured then shared roots",
			adapter:      kimi.NewAdapter(),
			wantAgent:    model.AgentKimi,
			wantProvider: true,
			wantSkills:   true,
			wantRoots: []SkillDiscoveryRoot{
				{Scope: SkillDiscoveryNativeGlobal, Path: native(".config", "agents", "skills")},
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
			name:       "Kilo Code has no shared declaration",
			adapter:    kilocode.NewAdapter(),
			wantAgent:  model.AgentKilocode,
			wantSkills: true,
			wantRoots:  []SkillDiscoveryRoot{{Scope: SkillDiscoveryNativeGlobal, Path: native(".config", "kilo", "skills")}},
		},
		{
			name:       "Trae IDE has no shared declaration",
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
