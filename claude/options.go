package claude

import (
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

// DefaultClientOptions returns default options for the Claude client.
// Composes defaults from all focused option structs.
func DefaultClientOptions() *ClientOptions {
	return &ClientOptions{
		ConnectionOptions:      shared.DefaultConnectionOptions(),
		BufferOptions:          shared.DefaultBufferOptions(),
		ModelOptions:           shared.DefaultModelOptions(),
		DebugOptions:           shared.DefaultDebugOptions(),
		IncludePartialMessages: false,
		EnableStructuredOutput: false,
		McpServers:             nil,
	}
}

// Validate validates the options and returns an error if invalid.
// Delegates to the shared Options.Validate for canonical validation.
func (o *ClientOptions) Validate() error {
	sharedOpts := o.toSharedOptions()
	return sharedOpts.Validate()
}

// toSharedOptions converts ClientOptions to shared.Options for validation.
func (o *ClientOptions) toSharedOptions() *shared.Options {
	return &shared.Options{
		Model:                  o.Model,
		CLIPath:                o.CLIPath,
		CLICommand:             o.CLICommand,
		PermissionMode:         o.PermissionMode,
		ContextFiles:           o.ContextFiles,
		IncludePartialMessages: o.IncludePartialMessages,
		EnableStructuredOutput: o.EnableStructuredOutput,
		Timeout:                o.Timeout,
		CustomArgs:             o.CustomArgs,
		Env:                    o.Env,
		MaxMessages:            o.MaxMessages,
		BufferSize:             o.BufferSize,
		Trace:                  o.Trace,
		DisableCache:           o.DisableCache,
		CacheTTL:               o.CacheTTL,
	}
}

// WithModel sets the model option.
func WithModel(model string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.Model = model
	}
}

// WithCLIPath sets the CLI path option.
func WithCLIPath(path string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.CLIPath = path
	}
}

// WithCLICommand sets the CLI command option.
func WithCLICommand(command string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.CLICommand = command
	}
}

// WithPermissionMode sets the permission mode option.
func WithPermissionMode(mode string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.PermissionMode = mode
	}
}

// WithContextFiles sets the context files option.
func WithContextFiles(files ...string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.ContextFiles = files
	}
}

// WithIncludePartialMessages enables partial messages.
func WithIncludePartialMessages(include bool) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.IncludePartialMessages = include
	}
}

// WithEnableStructuredOutput enables structured output.
func WithEnableStructuredOutput(enable bool) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.EnableStructuredOutput = enable
	}
}

// WithTimeout sets the timeout option.
func WithTimeout(timeout string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.Timeout = timeout
	}
}

// WithCustomArgs sets custom arguments.
func WithCustomArgs(args ...string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.CustomArgs = args
	}
}

// WithEnv sets environment variables.
func WithEnv(env map[string]string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		if o.Env == nil {
			o.Env = make(map[string]string)
		}
		for k, v := range env {
			o.Env[k] = v
		}
	}
}

// WithMaxMessages sets the maximum number of messages.
func WithMaxMessages(max int) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.MaxMessages = max
	}
}

// WithBufferSize sets the buffer size for message channels.
func WithBufferSize(size int) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.BufferSize = size
	}
}

// WithTrace enables detailed tracing.
func WithTrace(trace bool) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.Trace = trace
	}
}

// WithDisableCache disables caching.
func WithDisableCache(disable bool) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.DisableCache = disable
	}
}

// WithCacheTTL sets cache expiration time.
func WithCacheTTL(ttl string) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.CacheTTL = ttl
	}
}

// WithLogger sets the logger interface.
func WithLogger(logger shared.Logger) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.Logger = logger
	}
}

// WithEnableMetrics enables performance metrics.
func WithEnableMetrics(enable bool) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.EnableMetrics = enable
	}
}

// Focused option group setters for composition

// WithConnectionOptions sets all connection-related options at once.
func WithConnectionOptions(conn shared.ConnectionOptions) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.ConnectionOptions = conn
	}
}

// WithBufferOptions sets all buffer-related options at once.
func WithBufferOptions(buf shared.BufferOptions) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.BufferOptions = buf
	}
}

// WithModelOptions sets all model-related options at once.
func WithModelOptions(model shared.ModelOptions) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.ModelOptions = model
	}
}

// WithDebugOptions sets all debug-related options at once.
func WithDebugOptions(debug shared.DebugOptions) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.DebugOptions = debug
	}
}

// WithMcpServers sets MCP server configurations.
func WithMcpServers(servers map[string]shared.McpServerConfig) func(*ClientOptions) {
	return func(o *ClientOptions) {
		o.McpServers = servers
	}
}

// =============================================================================
// Tools Configuration Options
// =============================================================================

// WithToolsPreset sets tools configuration to use a preset.
// Available presets: "claude_code" (standard Claude Code tools)
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithToolsPreset("claude_code"),
//	)
func WithToolsPreset(preset string) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--tools", "preset:"+preset)
	}
}

// WithClaudeCodeTools is a convenience option for enabling standard Claude Code tools.
// Equivalent to WithToolsPreset("claude_code").
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithClaudeCodeTools(),
//	)
func WithClaudeCodeTools() ClientOption {
	return WithToolsPreset("claude_code")
}

// WithDisallowedTools sets tools that Claude should not use.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithDisallowedTools("Write", "Edit"),
//	)
func WithDisallowedTools(tools ...string) ClientOption {
	return func(o *ClientOptions) {
		for _, tool := range tools {
			o.CustomArgs = append(o.CustomArgs, "--disallowed-tools", tool)
		}
	}
}

// =============================================================================
// Session Options
// =============================================================================

// WithContinue continues the most recent conversation.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithContinue(),
//	)
func WithContinue() ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--continue")
	}
}

// WithResume resumes a specific session by ID.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithResume("session-uuid-here"),
//	)
func WithResume(sessionID string) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--resume", sessionID)
	}
}

// WithSystemPrompt sets the system prompt for the session.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithSystemPrompt("You are a helpful coding assistant."),
//	)
func WithSystemPrompt(prompt string) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--system-prompt", prompt)
	}
}

// WithAppendSystemPrompt appends to the system prompt.
// This adds to the base prompt rather than replacing it.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithAppendSystemPrompt("Focus on Go development."),
//	)
func WithAppendSystemPrompt(prompt string) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--append-system-prompt", prompt)
	}
}

// =============================================================================
// Limit Options
// =============================================================================

// WithMaxTurns limits the number of conversation turns.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithMaxTurns(10),
//	)
func WithMaxTurns(turns int) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--max-turns", fmt.Sprintf("%d", turns))
	}
}

// WithMaxBudgetUSD sets a spending limit in USD.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithMaxBudgetUSD(5.00),
//	)
func WithMaxBudgetUSD(budget float64) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--max-budget-usd", fmt.Sprintf("%.2f", budget))
	}
}

// WithMaxThinkingTokens limits the number of thinking tokens.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithMaxThinkingTokens(4096),
//	)
func WithMaxThinkingTokens(tokens int) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--max-thinking-tokens", fmt.Sprintf("%d", tokens))
	}
}

// =============================================================================
// Directory Options
// =============================================================================

// WithWorkingDirectory sets the working directory for the session.
// The directory must be an absolute path.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithWorkingDirectory("/path/to/project"),
//	)
func WithWorkingDirectory(dir string) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--cwd", dir)
	}
}

// WithAdditionalDirectories adds directories that Claude can access.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithAdditionalDirectories("/path/to/lib", "/path/to/data"),
//	)
func WithAdditionalDirectories(dirs ...string) ClientOption {
	return func(o *ClientOptions) {
		for _, dir := range dirs {
			o.CustomArgs = append(o.CustomArgs, "--add-dir", dir)
		}
	}
}

// =============================================================================
// Agent Options
// =============================================================================

// WithAgent sets the agent to use for the session.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithAgent("code-review"),
//	)
func WithAgent(agent string) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--agent", agent)
	}
}

// =============================================================================
// File Checkpointing
// =============================================================================

// WithFileCheckpointing enables file checkpointing for RewindFiles support.
// When enabled, file changes are tracked and can be reverted using RewindFiles().
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithFileCheckpointing(),
//	)
func WithFileCheckpointing() ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--enable-file-checkpointing")
	}
}
