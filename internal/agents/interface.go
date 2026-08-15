package agents

import (
	"context"

	"github.com/gentleman-programming/gentle-ai/v2/internal/model"
	"github.com/gentleman-programming/gentle-ai/v2/internal/system"
)

// Capability tags for adapter feature checks.
type Capability string

// Adapter is the core abstraction for AI agent integration. Components use
// adapter methods instead of switch statements on AgentID, making it trivial
// to add new agents without modifying component code.
type Adapter interface {
	// Identity
	Agent() model.AgentID
	Tier() model.SupportTier

	// Detection
	Detect(ctx context.Context, homeDir string) (installed bool, binaryPath string, configPath string, configFound bool, err error)

	// Installation
	InstallCommand(profile system.PlatformProfile) ([][]string, error)

	// Config paths — components use these instead of hardcoding paths per agent.
	GlobalConfigDir(homeDir string) string
	SystemPromptDir(homeDir string) string
	SystemPromptFile(homeDir string) string
	SkillsDir(homeDir string) string
	SettingsPath(homeDir string) string

	// Config strategies — HOW to inject content, not WHERE (that's paths above).
	SystemPromptStrategy() model.SystemPromptStrategy
	MCPStrategy() model.MCPStrategy

	// MCP path resolution
	MCPConfigPath(homeDir string, serverName string) string

	// Optional capabilities — compatibility projections of the adapter's
	// canonical AgentCapabilityManifest.
	SupportsOutputStyles() bool
	OutputStyleDir(homeDir string) string

	SupportsSlashCommands() bool
	CommandsDir(homeDir string) string

	SupportsSubAgents() bool
	SubAgentsDir(homeDir string) string
	EmbeddedSubAgentsDir() string

	SupportsSkills() bool
	SupportsSystemPrompt() bool
	SupportsMCP() bool
}

// EffectiveCodeGraphWiringDetector is an optional adapter capability for agents
// whose configuration format requires semantic validation beyond marker checks.
type EffectiveCodeGraphWiringDetector interface {
	EffectiveCodeGraphWiring(homeDir string) (path string, configured bool)
}

// Skill discovery contract aliases are defined here so consumers depend on the
// adapter boundary. Their model storage keeps individual adapter packages from
// importing this parent package and creating an import cycle.
type SkillDiscoveryScope = model.SkillDiscoveryScope

const (
	SkillDiscoverySharedGlobal = model.SkillDiscoverySharedGlobal
	SkillDiscoveryNativeGlobal = model.SkillDiscoveryNativeGlobal
)

type SkillDiscoveryRoot = model.SkillDiscoveryRoot
type SkillDiscoveryCapabilities = model.SkillDiscoveryCapabilities

// SkillDiscoveryProvider is an optional adapter capability for runtimes with
// verified global roots beyond their native SkillsDir. Roots must be ordered by
// runtime precedence.
type SkillDiscoveryProvider interface {
	SkillDiscovery(homeDir string) SkillDiscoveryCapabilities
}

// SkillDiscoveryRoots returns declared roots when an adapter provides them.
// Adapters without the optional capability fall back to their single native
// skills directory only when generic skills are supported. This function does
// not inspect or create filesystem paths.
func SkillDiscoveryRoots(adapter Adapter, homeDir string) []SkillDiscoveryRoot {
	if provider, ok := adapter.(SkillDiscoveryProvider); ok {
		roots := provider.SkillDiscovery(homeDir).Roots
		return append([]SkillDiscoveryRoot(nil), roots...)
	}

	if !adapter.SupportsSkills() {
		return nil
	}
	if skillsDir := adapter.SkillsDir(homeDir); skillsDir != "" {
		return []SkillDiscoveryRoot{{
			Scope: SkillDiscoveryNativeGlobal,
			Path:  skillsDir,
		}}
	}
	return nil
}
