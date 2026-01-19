package claude

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
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
// Performs comprehensive validation including conflict detection.
func (o *ClientOptions) Validate() error {
	// Delegate core validation to shared
	sharedOpts := o.toSharedOptions()
	if err := sharedOpts.Validate(); err != nil {
		return err
	}

	// Validate OutputFormat if set
	if o.OutputFormat != nil {
		if o.OutputFormat.Type == "" {
			return shared.NewConfigurationError("OutputFormat.Type", "", "output format type is required")
		}
		if o.OutputFormat.Type == "json_schema" && o.OutputFormat.Schema == nil {
			return shared.NewConfigurationError("OutputFormat.Schema", "nil", "JSON schema is required when type is json_schema")
		}
	}

	// Validate Sandbox settings if set
	if o.Sandbox != nil {
		if o.Sandbox.Type != "" && o.Sandbox.Type != "docker" && o.Sandbox.Type != "nsjail" {
			return shared.NewConfigurationError("Sandbox.Type", o.Sandbox.Type, "sandbox type must be 'docker' or 'nsjail'")
		}
	}

	// Validate agents if set
	for name, agent := range o.Agents {
		if name == "" {
			return shared.NewConfigurationError("Agents", "", "agent name cannot be empty")
		}
		if agent.Model != "" && agent.Model != shared.AgentModelSonnet &&
			agent.Model != shared.AgentModelOpus && agent.Model != shared.AgentModelHaiku &&
			agent.Model != shared.AgentModelInherit {
			return shared.NewConfigurationError("Agents["+name+"].Model", string(agent.Model),
				"agent model must be 'sonnet', 'opus', 'haiku', or 'inherit'")
		}
	}

	// Detect conflicting allowed/disallowed tools
	if err := o.validateToolConflicts(); err != nil {
		return err
	}

	return nil
}

// validateToolConflicts checks for tools that appear in both allowed and disallowed lists.
func (o *ClientOptions) validateToolConflicts() error {
	var allowedTools, disallowedTools []string

	// Parse CustomArgs to find --allowed-tools and --disallowed-tools
	for i := 0; i < len(o.CustomArgs); i++ {
		if o.CustomArgs[i] == "--allowed-tools" && i+1 < len(o.CustomArgs) {
			allowedTools = append(allowedTools, splitTools(o.CustomArgs[i+1])...)
		}
		if o.CustomArgs[i] == "--disallowed-tools" && i+1 < len(o.CustomArgs) {
			disallowedTools = append(disallowedTools, o.CustomArgs[i+1])
		}
	}

	// Check for conflicts
	for _, allowed := range allowedTools {
		for _, disallowed := range disallowedTools {
			if allowed == disallowed {
				return shared.NewConfigurationError("Tools", allowed,
					fmt.Sprintf("tool '%s' appears in both allowed and disallowed lists", allowed))
			}
		}
	}

	return nil
}

// splitTools splits a comma-separated tool list.
func splitTools(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, t := range strings.Split(s, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
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
	return func(o *ClientOptions) { shared.WithModelModel(model)(&o.ModelOptions) }
}

// WithCLIPath sets the CLI path option.
func WithCLIPath(path string) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithConnCLIPath(path)(&o.ConnectionOptions) }
}

// WithCLICommand sets the CLI command option.
func WithCLICommand(command string) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithConnCLICommand(command)(&o.ConnectionOptions) }
}

// WithPermissionMode sets the permission mode option.
// Accepts either a string or shared.PermissionMode constant.
//
// Example:
//
//	client, _ := claude.NewClient(claude.WithPermissionMode("default"))
func WithPermissionMode(mode string) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithModelPermissionMode(mode)(&o.ModelOptions) }
}

// WithTypedPermissionMode sets the permission mode using typed constants.
// This provides compile-time safety for permission mode values.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithTypedPermissionMode(shared.PermissionModeAcceptEdits),
//	)
func WithTypedPermissionMode(mode shared.PermissionMode) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithModelPermissionMode(string(mode))(&o.ModelOptions) }
}

// WithContextFiles sets the context files option.
func WithContextFiles(files ...string) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithModelContextFiles(files...)(&o.ModelOptions) }
}

// WithIncludePartialMessages enables partial messages.
// Note: This field is on ClientOptions directly, not on an embedded struct.
func WithIncludePartialMessages(include bool) func(*ClientOptions) {
	return func(o *ClientOptions) { o.IncludePartialMessages = include }
}

// WithEnableStructuredOutput enables structured output.
// Note: This field is on ClientOptions directly, not on an embedded struct.
func WithEnableStructuredOutput(enable bool) func(*ClientOptions) {
	return func(o *ClientOptions) { o.EnableStructuredOutput = enable }
}

// WithTimeout sets the timeout option.
func WithTimeout(timeout string) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithConnTimeout(timeout)(&o.ConnectionOptions) }
}

// WithCustomArgs sets custom arguments.
func WithCustomArgs(args ...string) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithModelCustomArgs(args...)(&o.ModelOptions) }
}

// WithEnv sets environment variables.
func WithEnv(env map[string]string) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithConnEnv(env)(&o.ConnectionOptions) }
}

// WithMaxMessages sets the maximum number of messages.
func WithMaxMessages(max int) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithBufMaxMessages(max)(&o.BufferOptions) }
}

// WithBufferSize sets the buffer size for message channels.
func WithBufferSize(size int) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithBufBufferSize(size)(&o.BufferOptions) }
}

// WithTrace enables detailed tracing.
func WithTrace(trace bool) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithDebugTrace(trace)(&o.DebugOptions) }
}

// WithDisableCache disables caching.
func WithDisableCache(disable bool) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithDebugDisableCache(disable)(&o.DebugOptions) }
}

// WithCacheTTL sets cache expiration time.
func WithCacheTTL(ttl string) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithDebugCacheTTL(ttl)(&o.DebugOptions) }
}

// WithLogger sets the logger interface.
func WithLogger(logger shared.Logger) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithDebugLogger(logger)(&o.DebugOptions) }
}

// WithEnableMetrics enables performance metrics.
func WithEnableMetrics(enable bool) func(*ClientOptions) {
	return func(o *ClientOptions) { shared.WithDebugEnableMetrics(enable)(&o.DebugOptions) }
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

// =============================================================================
// P0 Parity Options - Critical for Python SDK feature parity
// =============================================================================

// Note: WithAllowedTools is defined in mcp.go for backward compatibility.
// It restricts Claude to only use the specified tools.

// WithBetas enables beta features for the session.
// Multiple beta features can be enabled simultaneously.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithBetas("context-1m-2025-08-07"),
//	)
func WithBetas(betas ...string) ClientOption {
	return func(o *ClientOptions) {
		for _, beta := range betas {
			o.CustomArgs = append(o.CustomArgs, "--beta", beta)
		}
	}
}

// WithCanUseTool sets a permission callback that is invoked before each tool use.
// The callback can approve, deny, or modify tool inputs.
// Errors in the callback default to deny for safety.
//
// This enables the control protocol for bidirectional communication.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithCanUseTool(func(ctx context.Context, toolName string, toolInput map[string]any, opts shared.CanUseToolOptions) (shared.PermissionResult, error) {
//	        if toolName == "Bash" && strings.Contains(toolInput["command"].(string), "rm -rf") {
//	            return shared.NewPermissionResultDeny("Dangerous command blocked"), nil
//	        }
//	        return shared.NewPermissionResultAllow(), nil
//	    }),
//	)
func WithCanUseTool(callback shared.CanUseToolCallback) ClientOption {
	return func(o *ClientOptions) {
		o.CanUseTool = callback
	}
}

// WithHooks registers multiple hook callbacks at once.
// The map key is the hook event type (e.g., "PreToolUse", "PostToolUse").
// This enables the control protocol for bidirectional communication.
//
// For type-safe hook registration, prefer WithPreToolUseHook, WithPostToolUseHook, etc.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithHooks(map[shared.HookEvent][]shared.HookConfig{
//	        shared.HookEventPreToolUse: {
//	            {
//	                Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
//	                    // Handle pre-tool-use event
//	                    return &shared.SyncHookOutput{Continue: true}, nil
//	                },
//	            },
//	        },
//	    }),
//	)
func WithHooks(hooks map[shared.HookEvent][]shared.HookConfig) ClientOption {
	return func(o *ClientOptions) {
		o.Hooks = hooks
	}
}

// WithHook registers a single hook callback for the specified event.
// Can be called multiple times to register hooks for different events.
// Last registration wins for duplicate events.
//
// This enables the control protocol for bidirectional communication.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithHook(shared.HookEventPreToolUse, shared.HookConfig{
//	        Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
//	            return &shared.SyncHookOutput{Continue: true}, nil
//	        },
//	    }),
//	)
func WithHook(event shared.HookEvent, config shared.HookConfig) ClientOption {
	return func(o *ClientOptions) {
		if o.Hooks == nil {
			o.Hooks = make(map[shared.HookEvent][]shared.HookConfig)
		}
		config.Event = event // Ensure event is set
		o.Hooks[event] = append(o.Hooks[event], config)
	}
}

// withTypedHook is a generic helper that creates a typed hook wrapper.
// It eliminates duplication across WithPreToolUseHook, WithPostToolUseHook, etc.
// The fallbackOnMismatch parameter controls fail-open (continue) vs fail-closed (block) behavior.
func withTypedHook[T any](event shared.HookEvent, fn func(ctx context.Context, input *T) (*shared.SyncHookOutput, error), failClosed bool) ClientOption {
	return WithHook(event, shared.HookConfig{
		Event: event,
		Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
			if typedInput, ok := input.(*T); ok {
				return fn(ctx, typedInput)
			}
			// Type mismatch: fail-closed (block) for pre-hooks, fail-open (continue) for others
			if failClosed {
				return &shared.SyncHookOutput{Decision: "block", Reason: "invalid input type"}, nil
			}
			return &shared.SyncHookOutput{Continue: true}, nil
		},
	})
}

// WithPreToolUseHook is a convenience option for registering a PreToolUse hook.
// The callback is invoked before each tool use and can approve or block execution.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithPreToolUseHook(func(ctx context.Context, input *shared.PreToolUseHookInput) (*shared.SyncHookOutput, error) {
//	        if input.ToolName == "Bash" {
//	            return &shared.SyncHookOutput{Decision: "block", Reason: "Bash not allowed"}, nil
//	        }
//	        return &shared.SyncHookOutput{Continue: true}, nil
//	    }),
//	)
func WithPreToolUseHook(fn func(ctx context.Context, input *shared.PreToolUseHookInput) (*shared.SyncHookOutput, error)) ClientOption {
	return withTypedHook(shared.HookEventPreToolUse, fn, true) // fail-closed for pre-hooks
}

// WithPostToolUseHook is a convenience option for registering a PostToolUse hook.
// The callback is invoked after each successful tool use.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithPostToolUseHook(func(ctx context.Context, input *shared.PostToolUseHookInput) (*shared.SyncHookOutput, error) {
//	        log.Printf("Tool %s completed with result: %v", input.ToolName, input.ToolResponse)
//	        return &shared.SyncHookOutput{Continue: true}, nil
//	    }),
//	)
func WithPostToolUseHook(fn func(ctx context.Context, input *shared.PostToolUseHookInput) (*shared.SyncHookOutput, error)) ClientOption {
	return withTypedHook(shared.HookEventPostToolUse, fn, false) // fail-open for post-hooks
}

// WithSessionStartHook is a convenience option for registering a SessionStart hook.
// The callback is invoked when a session starts or resumes.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithSessionStartHook(func(ctx context.Context, input *shared.SessionStartHookInput) (*shared.SyncHookOutput, error) {
//	        log.Printf("Session started: %s", input.SessionID)
//	        return &shared.SyncHookOutput{Continue: true}, nil
//	    }),
//	)
func WithSessionStartHook(fn func(ctx context.Context, input *shared.SessionStartHookInput) (*shared.SyncHookOutput, error)) ClientOption {
	return withTypedHook(shared.HookEventSessionStart, fn, false) // fail-open for session hooks
}

// WithSessionEndHook is a convenience option for registering a SessionEnd hook.
// The callback is invoked when a session ends.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithSessionEndHook(func(ctx context.Context, input *shared.SessionEndHookInput) (*shared.SyncHookOutput, error) {
//	        log.Printf("Session ended: %s, reason: %s", input.SessionID, input.Reason)
//	        return &shared.SyncHookOutput{Continue: true}, nil
//	    }),
//	)
func WithSessionEndHook(fn func(ctx context.Context, input *shared.SessionEndHookInput) (*shared.SyncHookOutput, error)) ClientOption {
	return withTypedHook(shared.HookEventSessionEnd, fn, false) // fail-open for session hooks
}

// =============================================================================
// P0 Remaining: Structured Output Options
// =============================================================================

// WithJSONSchema sets a JSON schema for Claude's structured output.
// When set, Claude's response will conform to the provided schema.
//
// Example:
//
//	schema := map[string]any{
//	    "type": "object",
//	    "properties": map[string]any{
//	        "name": map[string]any{"type": "string"},
//	        "age":  map[string]any{"type": "number"},
//	    },
//	    "required": []string{"name", "age"},
//	}
//	client, _ := claude.NewClient(
//	    claude.WithJSONSchema(schema),
//	)
func WithJSONSchema(schema map[string]any) ClientOption {
	return func(o *ClientOptions) {
		o.OutputFormat = &shared.OutputFormat{
			Type:   "json_schema",
			Schema: schema,
		}
	}
}

// WithOutputFormat sets the output format configuration directly.
// For JSON schema output, prefer WithJSONSchema() for convenience.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithOutputFormat(&shared.OutputFormat{
//	        Type:   "json_schema",
//	        Schema: mySchema,
//	    }),
//	)
func WithOutputFormat(format *shared.OutputFormat) ClientOption {
	return func(o *ClientOptions) {
		o.OutputFormat = format
	}
}

// =============================================================================
// P1 Options: Aliases and Additional Features
// =============================================================================

// WithCwd is an alias for WithWorkingDirectory for severity1 SDK compatibility.
// Sets the working directory for the session.
//
// Example:
//
//	client, _ := claude.NewClient(claude.WithCwd("/tmp"))
func WithCwd(path string) ClientOption {
	return WithWorkingDirectory(path)
}

// WithAddDirs is an alias for WithAdditionalDirectories for severity1 SDK compatibility.
// Adds directories that Claude can access.
//
// Example:
//
//	client, _ := claude.NewClient(claude.WithAddDirs("/data", "/logs"))
func WithAddDirs(dirs ...string) ClientOption {
	return WithAdditionalDirectories(dirs...)
}

// WithForkSession creates a new session forked from an existing session.
// The fork includes conversation history up to the fork point.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithForkSession("session-uuid-here"),
//	)
func WithForkSession(sessionID string) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--resume", sessionID, "--fork")
	}
}

// WithFallbackModel sets a fallback model to use if the primary model is unavailable.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithModel("claude-opus-4"),
//	    claude.WithFallbackModel("claude-sonnet-4"),
//	)
func WithFallbackModel(model string) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--fallback-model", model)
	}
}

// WithUser sets a user identifier for usage tracking in multi-tenant scenarios.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithUser("user-12345"),
//	)
func WithUser(userID string) ClientOption {
	return func(o *ClientOptions) {
		o.CustomArgs = append(o.CustomArgs, "--user", userID)
	}
}

// WithDebugWriter sets a writer for debug output from the CLI subprocess.
// If nil (default), stderr is suppressed or isolated.
//
// Example:
//
//	var debugBuf bytes.Buffer
//	client, _ := claude.NewClient(
//	    claude.WithDebugWriter(&debugBuf),
//	)
func WithDebugWriter(w io.Writer) ClientOption {
	return func(o *ClientOptions) {
		o.DebugWriter = w
	}
}

// WithStderrCallback sets a callback for stderr output from the CLI subprocess.
// The callback is invoked for each line of stderr output.
// Takes precedence over DebugWriter if both are set.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithStderrCallback(func(line string) {
//	        log.Printf("Claude stderr: %s", line)
//	    }),
//	)
func WithStderrCallback(fn func(string)) ClientOption {
	return func(o *ClientOptions) {
		o.StderrCallback = fn
	}
}

// WithAgents registers multiple agent definitions at once.
// Agent definitions specify custom subagents with their own model and tool configurations.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithAgents(map[string]shared.AgentDefinition{
//	        "coder":    {Description: "Code writer", Model: shared.AgentModelSonnet},
//	        "reviewer": {Description: "Code reviewer", Model: shared.AgentModelOpus},
//	    }),
//	)
func WithAgents(agents map[string]shared.AgentDefinition) ClientOption {
	return func(o *ClientOptions) {
		o.Agents = agents
	}
}

// WithSettingSources controls where Claude loads settings from.
// Sources are applied in order (precedence).
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithSettingSources(shared.SettingSourceUser, shared.SettingSourceProject),
//	)
func WithSettingSources(sources ...shared.SettingSource) ClientOption {
	return func(o *ClientOptions) {
		o.SettingSources = sources
	}
}

// WithSandboxSettings configures sandbox behavior for bash command execution.
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithSandboxSettings(&shared.SandboxSettings{
//	        Enabled:              true,
//	        AutoAllowBashIfSandboxed: true,
//	        ExcludedCommands:     []string{"ls", "cat"},
//	    }),
//	)
func WithSandboxSettings(settings *shared.SandboxSettings) ClientOption {
	return func(o *ClientOptions) {
		o.Sandbox = settings
	}
}

// WithPluginConfig sets the full SDK plugin configuration.
// This provides complete control over plugin behavior including
// timeouts, concurrency limits, and custom configuration.
//
// Example:
//
//	client, err := claude.NewClient(
//	    claude.WithPluginConfig(&shared.SdkPluginConfig{
//	        Enabled:       true,
//	        PluginPath:    "/usr/local/lib/claude-plugins/analyzer.so",
//	        Config:        map[string]any{"log_level": "debug"},
//	        Timeout:       30 * time.Second,
//	        MaxConcurrent: 4,
//	    }),
//	)
func WithPluginConfig(config *shared.SdkPluginConfig) ClientOption {
	return func(o *ClientOptions) {
		o.SdkPluginConfig = config
	}
}
