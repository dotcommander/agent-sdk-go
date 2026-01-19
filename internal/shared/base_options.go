package shared

import "io"

// BaseOptions contains common options shared across different option types.
// This can be embedded by package-specific option types for composition.
type BaseOptions struct {
	// Model specifies the Claude model to use.
	Model string

	// SystemPrompt sets the system prompt. Replaces any existing system prompt.
	SystemPrompt string

	// AppendSystemPrompt appends to the system prompt.
	// Useful for adding domain-specific instructions to the base prompt.
	AppendSystemPrompt string

	// AllowedTools restricts which tools Claude can use.
	AllowedTools []string

	// PermissionMode sets the permission mode for file operations.
	PermissionMode string

	// ContextFiles provides files to include in the context.
	ContextFiles []string

	// CustomArgs provides additional CLI arguments.
	CustomArgs []string

	// Env contains environment variables to set for the subprocess.
	Env map[string]string

	// Cwd sets the working directory for the subprocess.
	// If empty, inherits parent process working directory.
	Cwd string

	// Tools configures tool availability.
	// Can be a preset ("claude_code") or explicit tool list.
	Tools *ToolsConfig

	// Stderr is a callback invoked for each stderr line from subprocess.
	// If nil, stderr goes to parent process stderr.
	Stderr func(line string)

	// CanUseTool is a callback for runtime permission checks.
	// Invoked before each tool use when permission mode requires it.
	CanUseTool CanUseToolCallback

	// Continue continues the most recent conversation.
	Continue bool

	// Resume is the session ID to resume.
	Resume string

	// ResumeSessionAt resumes at a specific message UUID.
	ResumeSessionAt string

	// ForkSession forks instead of continuing on resume.
	ForkSession bool

	// PersistSession saves sessions to disk (default: true).
	// Pointer to distinguish unset from false.
	PersistSession *bool

	// DisallowedTools are tools explicitly disallowed.
	DisallowedTools []string

	// MaxThinkingTokens limits thinking tokens.
	MaxThinkingTokens *int

	// MaxTurns limits conversation turns.
	MaxTurns *int

	// MaxBudgetUSD is the USD budget limit.
	MaxBudgetUSD *float64

	// FallbackModel is used if primary fails.
	FallbackModel string

	// AdditionalDirectories are extra accessible directories.
	AdditionalDirectories []string

	// Agent is the main thread agent name.
	Agent string

	// Agents are custom subagent definitions.
	Agents map[string]AgentDefinition

	// Betas are beta features to enable.
	Betas []string

	// EnableFileCheckpointing tracks file changes.
	EnableFileCheckpointing bool

	// OutputFormat is for structured responses.
	OutputFormat *OutputFormat

	// Plugins are plugin configurations.
	Plugins []PluginConfig

	// SdkPluginConfig is the full SDK plugin configuration.
	// Provides complete control over plugin behavior including
	// timeouts, concurrency limits, and custom configuration.
	SdkPluginConfig *SdkPluginConfig

	// SettingSources controls which settings to load.
	SettingSources []SettingSource

	// Sandbox is sandbox configuration.
	Sandbox *SandboxSettings

	// StrictMcpConfig enables strict MCP validation.
	StrictMcpConfig bool

	// AllowDangerouslySkipPermissions is required for bypass mode.
	AllowDangerouslySkipPermissions bool

	// PermissionPromptToolName is the MCP tool for permission prompts.
	PermissionPromptToolName string

	// McpServers are MCP server configurations.
	McpServers map[string]McpServerConfig

	// ExtraArgs are additional CLI arguments.
	ExtraArgs map[string]string

	// DebugWriter specifies where to write debug output from the CLI subprocess.
	// If nil (default), stderr is suppressed or isolated.
	// Common values: os.Stderr, io.Discard, or a custom io.Writer.
	DebugWriter io.Writer
}

// ToolsConfig is a discriminated union for tool configuration.
type ToolsConfig struct {
	Type   string   // "preset" | "explicit"
	Preset string   // "claude_code" (when Type == "preset")
	Tools  []string // Tool names (when Type == "explicit")
}

// ToolsPreset creates a tools config for a preset.
func ToolsPreset(preset string) *ToolsConfig {
	return &ToolsConfig{Type: "preset", Preset: preset}
}

// ToolsExplicit creates a tools config for explicit tool list.
func ToolsExplicit(tools ...string) *ToolsConfig {
	return &ToolsConfig{Type: "explicit", Tools: tools}
}

// DefaultBaseOptions returns default base options.
func DefaultBaseOptions() BaseOptions {
	return BaseOptions{
		Model: "claude-sonnet-4-5-20250929",
	}
}

// BaseOptionFunc is a function that configures BaseOptions.
type BaseOptionFunc func(*BaseOptions)

// WithBaseModel sets Model on BaseOptions.
func WithBaseModel(model string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Model = model
	}
}

// WithBaseSystemPrompt sets SystemPrompt on BaseOptions.
func WithBaseSystemPrompt(prompt string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.SystemPrompt = prompt
	}
}

// WithBaseAppendSystemPrompt sets AppendSystemPrompt on BaseOptions.
func WithBaseAppendSystemPrompt(prompt string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.AppendSystemPrompt = prompt
	}
}

// WithBaseAllowedTools sets AllowedTools on BaseOptions.
func WithBaseAllowedTools(tools ...string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.AllowedTools = tools
	}
}

// WithBasePermissionMode sets PermissionMode on BaseOptions.
func WithBasePermissionMode(mode string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.PermissionMode = mode
	}
}

// WithBaseContextFiles sets ContextFiles on BaseOptions.
func WithBaseContextFiles(files ...string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.ContextFiles = append(o.ContextFiles, files...)
	}
}

// WithBaseCustomArgs sets CustomArgs on BaseOptions.
func WithBaseCustomArgs(args ...string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.CustomArgs = args
	}
}

// WithBaseEnv sets Env on BaseOptions.
func WithBaseEnv(env map[string]string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Env = env
	}
}

// WithBaseDebugWriter sets DebugWriter on BaseOptions.
func WithBaseDebugWriter(w io.Writer) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.DebugWriter = w
	}
}

// WithBaseContinue sets Continue on BaseOptions.
func WithBaseContinue(cont bool) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Continue = cont
	}
}

// WithBaseResume sets Resume session ID on BaseOptions.
func WithBaseResume(sessionID string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Resume = sessionID
	}
}

// WithBaseResumeSessionAt sets ResumeSessionAt on BaseOptions.
func WithBaseResumeSessionAt(messageUUID string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.ResumeSessionAt = messageUUID
	}
}

// WithBaseForkSession sets ForkSession on BaseOptions.
func WithBaseForkSession(fork bool) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.ForkSession = fork
	}
}

// WithBasePersistSession sets PersistSession on BaseOptions.
func WithBasePersistSession(persist bool) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.PersistSession = &persist
	}
}

// WithBaseDisallowedTools sets DisallowedTools on BaseOptions.
func WithBaseDisallowedTools(tools ...string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.DisallowedTools = tools
	}
}

// WithBaseMaxThinkingTokens sets MaxThinkingTokens on BaseOptions.
func WithBaseMaxThinkingTokens(tokens int) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.MaxThinkingTokens = &tokens
	}
}

// WithBaseMaxTurns sets MaxTurns on BaseOptions.
func WithBaseMaxTurns(turns int) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.MaxTurns = &turns
	}
}

// WithBaseMaxBudgetUSD sets MaxBudgetUSD on BaseOptions.
func WithBaseMaxBudgetUSD(budget float64) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.MaxBudgetUSD = &budget
	}
}

// WithBaseFallbackModel sets FallbackModel on BaseOptions.
func WithBaseFallbackModel(model string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.FallbackModel = model
	}
}

// WithBaseAdditionalDirectories sets AdditionalDirectories on BaseOptions.
func WithBaseAdditionalDirectories(dirs ...string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.AdditionalDirectories = dirs
	}
}

// WithBaseAgent sets Agent on BaseOptions.
func WithBaseAgent(agent string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Agent = agent
	}
}

// WithBaseAgents sets Agents on BaseOptions.
func WithBaseAgents(agents map[string]AgentDefinition) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Agents = agents
	}
}

// WithBaseBetas sets Betas on BaseOptions.
func WithBaseBetas(betas ...string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Betas = betas
	}
}

// WithBaseEnableFileCheckpointing sets EnableFileCheckpointing on BaseOptions.
func WithBaseEnableFileCheckpointing(enable bool) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.EnableFileCheckpointing = enable
	}
}

// WithBaseOutputFormat sets OutputFormat on BaseOptions.
func WithBaseOutputFormat(format *OutputFormat) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.OutputFormat = format
	}
}

// WithBasePlugins sets Plugins on BaseOptions.
func WithBasePlugins(plugins ...PluginConfig) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Plugins = plugins
	}
}

// WithBaseSdkPluginConfig sets the full SDK plugin configuration on BaseOptions.
// This provides complete control over plugin behavior including timeouts,
// concurrency limits, and custom configuration.
func WithBaseSdkPluginConfig(config *SdkPluginConfig) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.SdkPluginConfig = config
	}
}

// WithBaseSettingSources sets SettingSources on BaseOptions.
func WithBaseSettingSources(sources ...SettingSource) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.SettingSources = sources
	}
}

// WithBaseSandbox sets Sandbox on BaseOptions.
func WithBaseSandbox(sandbox *SandboxSettings) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Sandbox = sandbox
	}
}

// WithBaseStrictMcpConfig sets StrictMcpConfig on BaseOptions.
func WithBaseStrictMcpConfig(strict bool) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.StrictMcpConfig = strict
	}
}

// WithBaseAllowDangerouslySkipPermissions sets AllowDangerouslySkipPermissions on BaseOptions.
func WithBaseAllowDangerouslySkipPermissions(allow bool) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.AllowDangerouslySkipPermissions = allow
	}
}

// WithBasePermissionPromptToolName sets PermissionPromptToolName on BaseOptions.
func WithBasePermissionPromptToolName(toolName string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.PermissionPromptToolName = toolName
	}
}

// WithBaseMcpServers sets McpServers on BaseOptions.
func WithBaseMcpServers(servers map[string]McpServerConfig) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.McpServers = servers
	}
}

// WithBaseExtraArgs sets ExtraArgs on BaseOptions.
func WithBaseExtraArgs(args map[string]string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.ExtraArgs = args
	}
}

// WithBaseCwd sets the working directory for the subprocess.
func WithBaseCwd(cwd string) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Cwd = cwd
	}
}

// WithBaseTools sets the tools configuration.
func WithBaseTools(tools *ToolsConfig) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Tools = tools
	}
}

// WithBaseStderr sets the stderr callback.
func WithBaseStderr(callback func(line string)) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.Stderr = callback
	}
}

// WithBaseCanUseTool sets the permission callback.
func WithBaseCanUseTool(callback CanUseToolCallback) BaseOptionFunc {
	return func(o *BaseOptions) {
		o.CanUseTool = callback
	}
}
