package v2

import (
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

// SessionOption is a function that configures a V2SessionOptions.
type SessionOption func(*V2SessionOptions)

// PromptOption is a function that configures a one-shot prompt operation.
type PromptOption func(*PromptOptions)

// =============================================================================
// DRY Option Generators
// These functions generate both SessionOption and PromptOption from a single
// BaseOptionFunc, eliminating duplicate implementations.
// =============================================================================

// sessionAndPromptOption generates both option types from a shared.BaseOptionFunc.
// This is the core DRY mechanism for option functions that apply to BaseOptions.
func sessionAndPromptOption(baseFn shared.BaseOptionFunc) (SessionOption, PromptOption) {
	return func(opts *V2SessionOptions) { baseFn(&opts.BaseOptions) },
		func(opts *PromptOptions) { baseFn(&opts.BaseOptions) }
}


// PromptOptions contains configuration for one-shot prompt operations.
// This mirrors the options parameter in unstable_v2_prompt().
// Embeds shared.BaseOptions for common fields (DRY).
type PromptOptions struct {
	shared.BaseOptions

	// Timeout specifies how long to wait for the response.
	Timeout time.Duration

	// clientFactory provides a factory for creating clients (DIP compliance).
	// If nil, the default factory is used.
	clientFactory ClientFactory

	// cliChecker provides a CLI availability checker (DIP compliance for testability).
	// If nil, the default cli.IsCLIAvailable is used.
	cliChecker shared.CLIChecker
}

// DefaultPromptOptions returns the default options for a prompt operation.
func DefaultPromptOptions() *PromptOptions {
	return &PromptOptions{
		BaseOptions: shared.DefaultBaseOptions(),
		Timeout:     60 * time.Second,
	}
}

// Validate validates the prompt options.
func (o *PromptOptions) Validate() error {
	if o.Model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// Resolve short model names before validation
	o.Model = shared.ResolveModelName(o.Model)

	if o.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// WithModel sets the Claude model to use for sessions.
func WithModel(model string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseModel(model))
	return s
}

// WithPromptModel sets the model for a one-shot prompt.
func WithPromptModel(model string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseModel(model))
	return p
}

// WithTimeout sets the timeout for session operations.
func WithTimeout(timeout time.Duration) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.Timeout = timeout
	}
}

// WithPromptTimeout sets the timeout for a one-shot prompt.
func WithPromptTimeout(timeout time.Duration) PromptOption {
	return func(opts *PromptOptions) {
		opts.Timeout = timeout
	}
}

// WithSystemPrompt sets the system prompt for sessions.
// This replaces any existing system prompt.
func WithSystemPrompt(prompt string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseSystemPrompt(prompt))
	return s
}

// WithPromptSystemPrompt sets the system prompt for a one-shot prompt.
func WithPromptSystemPrompt(prompt string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseSystemPrompt(prompt))
	return p
}

// WithAppendSystemPrompt appends to the system prompt.
// Useful for adding domain-specific instructions to the base prompt.
// Order matters: use WithSystemPrompt first, then WithAppendSystemPrompt.
//
// Example:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithSystemPrompt("You are a Go expert."),
//	    v2.WithAppendSystemPrompt("Always use idiomatic Go patterns."),
//	)
func WithAppendSystemPrompt(prompt string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseAppendSystemPrompt(prompt))
	return s
}

// WithPromptAppendSystemPrompt appends to the system prompt for a one-shot prompt.
func WithPromptAppendSystemPrompt(prompt string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseAppendSystemPrompt(prompt))
	return p
}

// WithAllowedTools restricts which tools Claude can use in sessions.
func WithAllowedTools(tools ...string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseAllowedTools(tools...))
	return s
}

// WithPromptAllowedTools restricts tools for a one-shot prompt.
func WithPromptAllowedTools(tools ...string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseAllowedTools(tools...))
	return p
}

// WithPermissionMode sets the permission mode for session file operations.
func WithPermissionMode(mode string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBasePermissionMode(mode))
	return s
}

// WithPromptPermissionMode sets the permission mode for a one-shot prompt.
func WithPromptPermissionMode(mode string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBasePermissionMode(mode))
	return p
}

// WithContextFiles adds files to the session context.
func WithContextFiles(files ...string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseContextFiles(files...))
	return s
}

// WithPromptContextFiles adds files to the context for a one-shot prompt.
func WithPromptContextFiles(files ...string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseContextFiles(files...))
	return p
}

// WithCustomArgs adds custom CLI arguments for sessions.
func WithCustomArgs(args ...string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseCustomArgs(args...))
	return s
}

// WithPromptCustomArgs adds custom CLI arguments for a one-shot prompt.
func WithPromptCustomArgs(args ...string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseCustomArgs(args...))
	return p
}

// WithEnv sets environment variables for session subprocess.
func WithEnv(env map[string]string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseEnv(env))
	return s
}

// WithPromptEnv sets environment variables for a one-shot prompt.
func WithPromptEnv(env map[string]string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseEnv(env))
	return p
}

// WithEnablePartialMessages enables streaming of partial messages.
func WithEnablePartialMessages(enable bool) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.EnablePartialMessages = enable
	}
}

// WithClientFactory sets a custom client factory for session creation (DIP compliance).
// This allows injecting mock clients for testing or custom client implementations.
//
// Example:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithClientFactory(myMockFactory))
func WithClientFactory(factory ClientFactory) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.clientFactory = factory
	}
}

// WithPromptClientFactory sets a custom client factory for one-shot prompts (DIP compliance).
// This allows injecting mock clients for testing or custom client implementations.
func WithPromptClientFactory(factory ClientFactory) PromptOption {
	return func(opts *PromptOptions) {
		opts.clientFactory = factory
	}
}

// WithPromptEnablePartialMessages enables partial messages for a one-shot prompt.
func WithPromptEnablePartialMessages(enable bool) PromptOption {
	return func(opts *PromptOptions) {
		// Partial messages are not applicable for one-shot prompts
		// This option is ignored but kept for API compatibility
	}
}

// WithCLIChecker sets a custom CLI checker for testability (DIP compliance).
// This allows tests to inject a mock CLI checker to control availability checks.
//
// Example:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithCLIChecker(shared.AlwaysAvailableCLIChecker{}))
func WithCLIChecker(checker shared.CLIChecker) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.cliChecker = checker
	}
}

// WithPromptCLIChecker sets a custom CLI checker for one-shot prompts.
func WithPromptCLIChecker(checker shared.CLIChecker) PromptOption {
	return func(opts *PromptOptions) {
		opts.cliChecker = checker
	}
}

// DefaultSessionOptions returns the default options for a V2 session.
func DefaultSessionOptions() *V2SessionOptions {
	return &V2SessionOptions{
		BaseOptions:           shared.DefaultBaseOptions(),
		Timeout:               60 * time.Second,
		EnablePartialMessages: true,
	}
}

// Validate validates the session options.
func (o *V2SessionOptions) Validate() error {
	if o.Model == "" {
		return fmt.Errorf("model cannot be empty")
	}

	// Resolve short model names (claude aliases only)
	o.Model = shared.ResolveModelName(o.Model)

	if o.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// Validate permission mode if specified
	if o.PermissionMode != "" {
		validModes := []string{"default", "auto", "grant", "accept_edits", "plan", "bypass_permissions"}
		if !slices.Contains(validModes, o.PermissionMode) {
			return fmt.Errorf("invalid permission mode: %s", o.PermissionMode)
		}
	}

	return nil
}

// WithContinue continues the most recent conversation.
func WithContinue(cont bool) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseContinue(cont))
	return s
}

// WithPromptContinue continues the most recent conversation.
func WithPromptContinue(cont bool) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseContinue(cont))
	return p
}

// WithResume sets the session ID to resume.
func WithResume(sessionID string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseResume(sessionID))
	return s
}

// WithPromptResume sets the session ID to resume.
func WithPromptResume(sessionID string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseResume(sessionID))
	return p
}

// WithResumeSessionAt resumes at a specific message UUID.
func WithResumeSessionAt(messageUUID string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseResumeSessionAt(messageUUID))
	return s
}

// WithForkSession forks instead of continuing on resume.
func WithForkSession(fork bool) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseForkSession(fork))
	return s
}

// WithPersistSession saves sessions to disk.
func WithPersistSession(persist bool) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBasePersistSession(persist))
	return s
}

// WithDisallowedTools sets tools explicitly disallowed.
func WithDisallowedTools(tools ...string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseDisallowedTools(tools...))
	return s
}

// WithPromptDisallowedTools sets tools explicitly disallowed.
func WithPromptDisallowedTools(tools ...string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseDisallowedTools(tools...))
	return p
}

// WithMaxThinkingTokens limits thinking tokens.
func WithMaxThinkingTokens(tokens int) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseMaxThinkingTokens(tokens))
	return s
}

// WithPromptMaxThinkingTokens limits thinking tokens.
func WithPromptMaxThinkingTokens(tokens int) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseMaxThinkingTokens(tokens))
	return p
}

// WithMaxTurns limits conversation turns.
func WithMaxTurns(turns int) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseMaxTurns(turns))
	return s
}

// WithPromptMaxTurns limits conversation turns.
func WithPromptMaxTurns(turns int) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseMaxTurns(turns))
	return p
}

// WithMaxBudgetUSD sets the USD budget limit.
func WithMaxBudgetUSD(budget float64) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseMaxBudgetUSD(budget))
	return s
}

// WithPromptMaxBudgetUSD sets the USD budget limit.
func WithPromptMaxBudgetUSD(budget float64) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseMaxBudgetUSD(budget))
	return p
}

// WithFallbackModel sets the model used if primary fails.
func WithFallbackModel(model string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseFallbackModel(model))
	return s
}

// WithPromptFallbackModel sets the model used if primary fails.
func WithPromptFallbackModel(model string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseFallbackModel(model))
	return p
}

// WithAdditionalDirectories sets extra accessible directories.
func WithAdditionalDirectories(dirs ...string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseAdditionalDirectories(dirs...))
	return s
}

// WithPromptAdditionalDirectories sets extra accessible directories.
func WithPromptAdditionalDirectories(dirs ...string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseAdditionalDirectories(dirs...))
	return p
}

// WithAgent sets the main thread agent name.
func WithAgent(agent string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseAgent(agent))
	return s
}

// WithPromptAgent sets the main thread agent name.
func WithPromptAgent(agent string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseAgent(agent))
	return p
}

// WithAgents sets custom subagent definitions.
func WithAgents(agents map[string]shared.AgentDefinition) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseAgents(agents))
	return s
}

// WithPromptAgents sets custom subagent definitions.
func WithPromptAgents(agents map[string]shared.AgentDefinition) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseAgents(agents))
	return p
}

// WithBetas enables beta features.
func WithBetas(betas ...string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseBetas(betas...))
	return s
}

// WithPromptBetas enables beta features.
func WithPromptBetas(betas ...string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseBetas(betas...))
	return p
}

// WithEnableFileCheckpointing tracks file changes.
func WithEnableFileCheckpointing(enable bool) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseEnableFileCheckpointing(enable))
	return s
}

// WithOutputFormat sets structured output configuration.
func WithOutputFormat(format *shared.OutputFormat) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseOutputFormat(format))
	return s
}

// WithPromptOutputFormat sets structured output configuration.
func WithPromptOutputFormat(format *shared.OutputFormat) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseOutputFormat(format))
	return p
}

// WithPlugins sets plugin configurations.
func WithPlugins(plugins ...shared.PluginConfig) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBasePlugins(plugins...))
	return s
}

// WithSettingSources controls which settings to load.
func WithSettingSources(sources ...shared.SettingSource) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseSettingSources(sources...))
	return s
}

// WithSandbox sets sandbox configuration.
func WithSandbox(sandbox *shared.SandboxSettings) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseSandbox(sandbox))
	return s
}

// WithStrictMcpConfig enables strict MCP validation.
func WithStrictMcpConfig(strict bool) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseStrictMcpConfig(strict))
	return s
}

// WithAllowDangerouslySkipPermissions enables bypass mode (requires flag).
func WithAllowDangerouslySkipPermissions(allow bool) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseAllowDangerouslySkipPermissions(allow))
	return s
}

// WithPermissionPromptToolName sets the MCP tool for permission prompts.
func WithPermissionPromptToolName(toolName string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBasePermissionPromptToolName(toolName))
	return s
}

// WithMcpServers sets MCP server configurations.
func WithMcpServers(servers map[string]shared.McpServerConfig) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseMcpServers(servers))
	return s
}

// WithPromptMcpServers sets MCP server configurations.
func WithPromptMcpServers(servers map[string]shared.McpServerConfig) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseMcpServers(servers))
	return p
}

// WithExtraArgs sets additional CLI arguments.
func WithExtraArgs(args map[string]string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseExtraArgs(args))
	return s
}

// WithHooks registers hook handlers for session lifecycle events.
// Hooks are executed in-process when the corresponding events occur.
//
// Example:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithModel("claude-sonnet-4-20250514"),
//	    v2.WithHooks(
//	        shared.HookConfig{
//	            Event:   shared.HookEventPreToolUse,
//	            Matcher: "Write|Edit",
//	            Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
//	                // Validate file path before write
//	                preToolInput := input.(*shared.PreToolUseHookInput)
//	                filePath := preToolInput.ToolInput["file_path"].(string)
//	                if !strings.HasPrefix(filePath, "/allowed/dir") {
//	                    return &shared.SyncHookOutput{
//	                        Decision: "block",
//	                        Reason:   "Cannot write outside allowed directory",
//	                    }, nil
//	                }
//	                return &shared.SyncHookOutput{Continue: true}, nil
//	            },
//	        },
//	    ),
//	)
func WithHooks(hooks ...shared.HookConfig) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.Hooks = append(opts.Hooks, hooks...)
	}
}

// WithCwd sets the working directory for the Claude CLI subprocess.
// File operations will resolve relative to this directory.
//
// Example:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithCwd("/path/to/project"),
//	)
func WithCwd(cwd string) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseCwd(cwd))
	return s
}

// WithPromptCwd sets the working directory for a one-shot prompt.
func WithPromptCwd(cwd string) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseCwd(cwd))
	return p
}

// WithTools sets the tools configuration (preset or explicit list).
//
// Example using preset:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithTools(shared.ToolsPreset("claude_code")),
//	)
//
// Example using explicit list:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithTools(shared.ToolsExplicit("Read", "Write", "Bash")),
//	)
func WithTools(tools *shared.ToolsConfig) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseTools(tools))
	return s
}

// WithPromptTools sets the tools configuration for a one-shot prompt.
func WithPromptTools(tools *shared.ToolsConfig) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseTools(tools))
	return p
}

// WithStderr sets a callback for subprocess stderr output.
// Each line from stderr is passed to the callback for logging or processing.
//
// Example:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithStderr(func(line string) {
//	        log.Printf("[Claude CLI] %s", line)
//	    }),
//	)
func WithStderr(callback func(line string)) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseStderr(callback))
	return s
}

// WithPromptStderr sets a callback for stderr output for a one-shot prompt.
func WithPromptStderr(callback func(line string)) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseStderr(callback))
	return p
}

// WithCanUseTool sets a permission callback for runtime tool approval.
// The callback is invoked before each tool execution to allow/deny usage.
//
// Example:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithCanUseTool(func(ctx context.Context, toolName string, toolInput map[string]any, opts shared.CanUseToolOptions) (shared.PermissionResult, error) {
//	        // Allow all reads, but require confirmation for writes
//	        if toolName == "Read" {
//	            return shared.PermissionResult{Behavior: shared.PermissionBehaviorAllow}, nil
//	        }
//	        if toolName == "Write" {
//	            // Check file path
//	            path := toolInput["file_path"].(string)
//	            if !strings.HasPrefix(path, "/allowed/") {
//	                return shared.PermissionResult{
//	                    Behavior: shared.PermissionBehaviorDeny,
//	                    Message:  "Cannot write outside /allowed/ directory",
//	                }, nil
//	            }
//	        }
//	        return shared.PermissionResult{Behavior: shared.PermissionBehaviorAllow}, nil
//	    }),
//	)
func WithCanUseTool(callback shared.CanUseToolCallback) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseCanUseTool(callback))
	return s
}

// WithPromptCanUseTool sets a permission callback for a one-shot prompt.
func WithPromptCanUseTool(callback shared.CanUseToolCallback) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseCanUseTool(callback))
	return p
}

// =============================================================================
// Debug Writer Options (P2)
// =============================================================================

// WithDebugWriter sets the writer for CLI debug output.
// If not set, stderr is isolated to prevent deadlocks (default behavior).
// Common values: os.Stderr, io.Discard, or a custom io.Writer like bytes.Buffer.
//
// Example:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithDebugWriter(os.Stderr), // See debug output in terminal
//	)
func WithDebugWriter(w io.Writer) SessionOption {
	s, _ := sessionAndPromptOption(shared.WithBaseDebugWriter(w))
	return s
}

// WithPromptDebugWriter sets the writer for CLI debug output in one-shot prompts.
func WithPromptDebugWriter(w io.Writer) PromptOption {
	_, p := sessionAndPromptOption(shared.WithBaseDebugWriter(w))
	return p
}

// WithDebugStderr redirects CLI debug output to os.Stderr.
// Useful for seeing debug output in real-time during development.
func WithDebugStderr() SessionOption {
	return WithDebugWriter(os.Stderr)
}

// WithDebugDisabled discards all CLI debug output.
// This is more explicit than nil but has the same effect.
func WithDebugDisabled() SessionOption {
	return WithDebugWriter(io.Discard)
}

// WithPromptDebugStderr redirects CLI debug output to os.Stderr for one-shot prompts.
func WithPromptDebugStderr() PromptOption {
	return WithPromptDebugWriter(os.Stderr)
}

// WithPromptDebugDisabled discards CLI debug output for one-shot prompts.
func WithPromptDebugDisabled() PromptOption {
	return WithPromptDebugWriter(io.Discard)
}
