package v2

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"agent-sdk-go/internal/claude/shared"
)

// SessionOption is a function that configures a V2SessionOptions.
type SessionOption func(*V2SessionOptions)

// PromptOption is a function that configures a one-shot prompt operation.
type PromptOption func(*PromptOptions)

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
	return func(opts *V2SessionOptions) {
		opts.Model = model
	}
}

// WithPromptModel sets the model for a one-shot prompt.
func WithPromptModel(model string) PromptOption {
	return func(opts *PromptOptions) {
		opts.Model = model
	}
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
func WithSystemPrompt(prompt string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.SystemPrompt = prompt
	}
}

// WithPromptSystemPrompt sets the system prompt for a one-shot prompt.
func WithPromptSystemPrompt(prompt string) PromptOption {
	return func(opts *PromptOptions) {
		opts.SystemPrompt = prompt
	}
}

// WithAllowedTools restricts which tools Claude can use in sessions.
func WithAllowedTools(tools ...string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.AllowedTools = tools
	}
}

// WithPromptAllowedTools restricts tools for a one-shot prompt.
func WithPromptAllowedTools(tools ...string) PromptOption {
	return func(opts *PromptOptions) {
		opts.AllowedTools = tools
	}
}

// WithPermissionMode sets the permission mode for session file operations.
func WithPermissionMode(mode string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.PermissionMode = mode
	}
}

// WithPromptPermissionMode sets the permission mode for a one-shot prompt.
func WithPromptPermissionMode(mode string) PromptOption {
	return func(opts *PromptOptions) {
		opts.PermissionMode = mode
	}
}

// WithContextFiles adds files to the session context.
func WithContextFiles(files ...string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.ContextFiles = append(opts.ContextFiles, files...)
	}
}

// WithPromptContextFiles adds files to the context for a one-shot prompt.
func WithPromptContextFiles(files ...string) PromptOption {
	return func(opts *PromptOptions) {
		opts.ContextFiles = append(opts.ContextFiles, files...)
	}
}

// WithCustomArgs adds custom CLI arguments for sessions.
func WithCustomArgs(args ...string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.CustomArgs = args
	}
}

// WithPromptCustomArgs adds custom CLI arguments for a one-shot prompt.
func WithPromptCustomArgs(args ...string) PromptOption {
	return func(opts *PromptOptions) {
		opts.CustomArgs = args
	}
}

// WithEnv sets environment variables for session subprocess.
func WithEnv(env map[string]string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.Env = env
	}
}

// WithPromptEnv sets environment variables for a one-shot prompt.
func WithPromptEnv(env map[string]string) PromptOption {
	return func(opts *PromptOptions) {
		opts.Env = env
	}
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

	// Resolve short model names before validation
	o.Model = shared.ResolveModelName(o.Model)

	// Validate model name format (basic check)
	if !strings.HasPrefix(o.Model, "claude-") {
		return fmt.Errorf("invalid model name: %s", o.Model)
	}

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
	return func(opts *V2SessionOptions) {
		opts.Continue = cont
	}
}

// WithResume sets the session ID to resume.
func WithResume(sessionID string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.Resume = sessionID
	}
}

// WithResumeSessionAt resumes at a specific message UUID.
func WithResumeSessionAt(messageUUID string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.ResumeSessionAt = messageUUID
	}
}

// WithForkSession forks instead of continuing on resume.
func WithForkSession(fork bool) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.ForkSession = fork
	}
}

// WithPersistSession saves sessions to disk.
func WithPersistSession(persist bool) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.PersistSession = &persist
	}
}

// WithDisallowedTools sets tools explicitly disallowed.
func WithDisallowedTools(tools ...string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.DisallowedTools = tools
	}
}

// WithMaxThinkingTokens limits thinking tokens.
func WithMaxThinkingTokens(tokens int) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.MaxThinkingTokens = &tokens
	}
}

// WithMaxTurns limits conversation turns.
func WithMaxTurns(turns int) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.MaxTurns = &turns
	}
}

// WithMaxBudgetUSD sets the USD budget limit.
func WithMaxBudgetUSD(budget float64) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.MaxBudgetUSD = &budget
	}
}

// WithFallbackModel sets the model used if primary fails.
func WithFallbackModel(model string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.FallbackModel = model
	}
}

// WithAdditionalDirectories sets extra accessible directories.
func WithAdditionalDirectories(dirs ...string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.AdditionalDirectories = dirs
	}
}

// WithAgent sets the main thread agent name.
func WithAgent(agent string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.Agent = agent
	}
}

// WithAgents sets custom subagent definitions.
func WithAgents(agents map[string]shared.AgentDefinition) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.Agents = agents
	}
}

// WithBetas enables beta features.
func WithBetas(betas ...string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.Betas = betas
	}
}

// WithEnableFileCheckpointing tracks file changes.
func WithEnableFileCheckpointing(enable bool) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.EnableFileCheckpointing = enable
	}
}

// WithOutputFormat sets structured output configuration.
func WithOutputFormat(format *shared.OutputFormat) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.OutputFormat = format
	}
}

// WithPlugins sets plugin configurations.
func WithPlugins(plugins ...shared.PluginConfig) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.Plugins = plugins
	}
}

// WithSettingSources controls which settings to load.
func WithSettingSources(sources ...shared.SettingSource) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.SettingSources = sources
	}
}

// WithSandbox sets sandbox configuration.
func WithSandbox(sandbox *shared.SandboxSettings) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.Sandbox = sandbox
	}
}

// WithStrictMcpConfig enables strict MCP validation.
func WithStrictMcpConfig(strict bool) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.StrictMcpConfig = strict
	}
}

// WithAllowDangerouslySkipPermissions enables bypass mode (requires flag).
func WithAllowDangerouslySkipPermissions(allow bool) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.AllowDangerouslySkipPermissions = allow
	}
}

// WithPermissionPromptToolName sets the MCP tool for permission prompts.
func WithPermissionPromptToolName(toolName string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.PermissionPromptToolName = toolName
	}
}

// WithMcpServers sets MCP server configurations.
func WithMcpServers(servers map[string]shared.McpServerConfig) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.McpServers = servers
	}
}

// WithExtraArgs sets additional CLI arguments.
func WithExtraArgs(args map[string]string) SessionOption {
	return func(opts *V2SessionOptions) {
		opts.ExtraArgs = args
	}
}

// Prompt option functions for the new BaseOptions fields.

// WithPromptContinue continues the most recent conversation.
func WithPromptContinue(cont bool) PromptOption {
	return func(opts *PromptOptions) {
		opts.Continue = cont
	}
}

// WithPromptResume sets the session ID to resume.
func WithPromptResume(sessionID string) PromptOption {
	return func(opts *PromptOptions) {
		opts.Resume = sessionID
	}
}

// WithPromptDisallowedTools sets tools explicitly disallowed.
func WithPromptDisallowedTools(tools ...string) PromptOption {
	return func(opts *PromptOptions) {
		opts.DisallowedTools = tools
	}
}

// WithPromptMaxThinkingTokens limits thinking tokens.
func WithPromptMaxThinkingTokens(tokens int) PromptOption {
	return func(opts *PromptOptions) {
		opts.MaxThinkingTokens = &tokens
	}
}

// WithPromptMaxTurns limits conversation turns.
func WithPromptMaxTurns(turns int) PromptOption {
	return func(opts *PromptOptions) {
		opts.MaxTurns = &turns
	}
}

// WithPromptMaxBudgetUSD sets the USD budget limit.
func WithPromptMaxBudgetUSD(budget float64) PromptOption {
	return func(opts *PromptOptions) {
		opts.MaxBudgetUSD = &budget
	}
}

// WithPromptFallbackModel sets the model used if primary fails.
func WithPromptFallbackModel(model string) PromptOption {
	return func(opts *PromptOptions) {
		opts.FallbackModel = model
	}
}

// WithPromptAdditionalDirectories sets extra accessible directories.
func WithPromptAdditionalDirectories(dirs ...string) PromptOption {
	return func(opts *PromptOptions) {
		opts.AdditionalDirectories = dirs
	}
}

// WithPromptAgent sets the main thread agent name.
func WithPromptAgent(agent string) PromptOption {
	return func(opts *PromptOptions) {
		opts.Agent = agent
	}
}

// WithPromptAgents sets custom subagent definitions.
func WithPromptAgents(agents map[string]shared.AgentDefinition) PromptOption {
	return func(opts *PromptOptions) {
		opts.Agents = agents
	}
}

// WithPromptBetas enables beta features.
func WithPromptBetas(betas ...string) PromptOption {
	return func(opts *PromptOptions) {
		opts.Betas = betas
	}
}

// WithPromptOutputFormat sets structured output configuration.
func WithPromptOutputFormat(format *shared.OutputFormat) PromptOption {
	return func(opts *PromptOptions) {
		opts.OutputFormat = format
	}
}

// WithPromptMcpServers sets MCP server configurations.
func WithPromptMcpServers(servers map[string]shared.McpServerConfig) PromptOption {
	return func(opts *PromptOptions) {
		opts.McpServers = servers
	}
}
