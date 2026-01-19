package shared

import (
	"fmt"
	"maps"
	"slices"
	"time"
)

// Options provides configuration for the Claude client.
type Options struct {
	// Model specifies the Claude model to use.
	Model string

	// CLIPath specifies the path to the Claude CLI executable.
	// If empty, the CLI will be discovered in PATH.
	CLIPath string

	// CLICommand specifies the command to run the Claude CLI.
	// If empty, defaults to "claude".
	CLICommand string

	// PermissionMode sets the permission mode for the session.
	PermissionMode string

	// ContextFiles specifies files to include in the context.
	ContextFiles []string

	// IncludePartialMessages enables streaming of partial messages during response generation.
	IncludePartialMessages bool

	// EnableStructuredOutput enables structured output mode.
	EnableStructuredOutput bool

	// Timeout specifies the timeout for operations.
	Timeout string

	// CustomArgs provides additional arguments for the Claude CLI.
	CustomArgs []string

	// Environment variables to set for the subprocess.
	Env map[string]string

	// Maximum number of messages to process before returning an error.
	MaxMessages int

	// Buffer size for message channels.
	BufferSize int

	// Logger interface for logging operations.
	Logger Logger

	// Enable performance metrics collection.
	EnableMetrics bool

	// Trace enables detailed tracing of protocol messages.
	Trace bool

	// DisableCache disables message caching.
	DisableCache bool

	// CacheTTL sets the cache expiration time.
	CacheTTL string
}

// DefaultOptions returns default options for the Claude client.
func DefaultOptions() *Options {
	return &Options{
		Model:                  "claude-3-5-sonnet-20241022",
		CLICommand:             "claude",
		PermissionMode:         "auto",
		Timeout:                "30s",
		IncludePartialMessages: false,
		EnableStructuredOutput: false,
		MaxMessages:            1000,
		BufferSize:             100,
		EnableMetrics:          false,
		Trace:                  false,
		DisableCache:           false,
		CacheTTL:               "1h",
		Env:                    make(map[string]string),
	}
}

// Validate validates the options and returns an error if invalid.
func (o *Options) Validate() error {
	if o.Model == "" {
		return NewConfigurationError("Model", "", "model is required")
	}

	// Resolve short model names (claude aliases only)
	o.Model = ResolveModelName(o.Model)

	// Validate permission mode
	validModes := []string{"auto", "read", "write", "restricted"}
	if !slices.Contains(validModes, o.PermissionMode) {
		return NewConfigurationError("PermissionMode", o.PermissionMode,
			fmt.Sprintf("invalid permission mode, must be one of: %v", validModes))
	}

	// Validate timeout format
	if o.Timeout != "" {
		if _, err := time.ParseDuration(o.Timeout); err != nil {
			return NewConfigurationError("Timeout", o.Timeout, "timeout must be a valid duration like '30s', '5m', '1h'")
		}
	}

	// Validate context files - reject empty paths
	if slices.Contains(o.ContextFiles, "") {
		return NewConfigurationError("ContextFiles", "", "context file path cannot be empty")
	}

	return nil
}

// WithModel sets the model option.
func WithModel(model string) func(*Options) {
	return func(o *Options) {
		o.Model = model
	}
}

// WithCLIPath sets the CLI path option.
func WithCLIPath(path string) func(*Options) {
	return func(o *Options) {
		o.CLIPath = path
	}
}

// WithCLICommand sets the CLI command option.
func WithCLICommand(command string) func(*Options) {
	return func(o *Options) {
		o.CLICommand = command
	}
}

// WithPermissionMode sets the permission mode option.
func WithPermissionMode(mode string) func(*Options) {
	return func(o *Options) {
		o.PermissionMode = mode
	}
}

// WithContextFiles sets the context files option.
func WithContextFiles(files ...string) func(*Options) {
	return func(o *Options) {
		o.ContextFiles = files
	}
}

// WithIncludePartialMessages enables partial messages.
func WithIncludePartialMessages(include bool) func(*Options) {
	return func(o *Options) {
		o.IncludePartialMessages = include
	}
}

// WithEnableStructuredOutput enables structured output.
func WithEnableStructuredOutput(enable bool) func(*Options) {
	return func(o *Options) {
		o.EnableStructuredOutput = enable
	}
}

// WithTimeout sets the timeout option.
func WithTimeout(timeout string) func(*Options) {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithCustomArgs sets custom arguments.
func WithCustomArgs(args ...string) func(*Options) {
	return func(o *Options) {
		o.CustomArgs = args
	}
}

// WithEnv sets environment variables.
func WithEnv(env map[string]string) func(*Options) {
	return func(o *Options) {
		maps.Copy(o.Env, env)
	}
}

// WithMaxMessages sets the maximum number of messages.
func WithMaxMessages(max int) func(*Options) {
	return func(o *Options) {
		o.MaxMessages = max
	}
}

// WithBufferSize sets the buffer size for message channels.
func WithBufferSize(size int) func(*Options) {
	return func(o *Options) {
		o.BufferSize = size
	}
}

// WithLogger sets the logger interface.
func WithLogger(logger Logger) func(*Options) {
	return func(o *Options) {
		o.Logger = logger
	}
}

// WithEnableMetrics enables performance metrics.
func WithEnableMetrics(enable bool) func(*Options) {
	return func(o *Options) {
		o.EnableMetrics = enable
	}
}

// WithTrace enables detailed tracing.
func WithTrace(trace bool) func(*Options) {
	return func(o *Options) {
		o.Trace = trace
	}
}

// WithDisableCache disables caching.
func WithDisableCache(disable bool) func(*Options) {
	return func(o *Options) {
		o.DisableCache = disable
	}
}

// WithCacheTTL sets cache expiration time.
func WithCacheTTL(ttl string) func(*Options) {
	return func(o *Options) {
		o.CacheTTL = ttl
	}
}

// =============================================================================
// Focused Option Structs (SRP)
// These are composed into larger option types for single responsibility.
// =============================================================================

// ConnectionOptions contains options for CLI connection.
// Single Responsibility: manages how we connect to the CLI.
type ConnectionOptions struct {
	// CLIPath specifies the path to the Claude CLI executable.
	// If empty, the CLI will be discovered in PATH.
	CLIPath string

	// CLICommand specifies the command to run the Claude CLI.
	// If empty, defaults to "claude".
	CLICommand string

	// Timeout specifies the timeout for operations.
	Timeout string

	// Env contains environment variables to set for the subprocess.
	Env map[string]string
}

// DefaultConnectionOptions returns default connection options.
func DefaultConnectionOptions() ConnectionOptions {
	return ConnectionOptions{
		CLICommand: "claude",
		Timeout:    "30s",
		Env:        make(map[string]string),
	}
}

// BufferOptions contains options for message buffering.
// Single Responsibility: manages buffer sizes and limits.
type BufferOptions struct {
	// BufferSize is the buffer size for message channels.
	BufferSize int

	// MaxMessages is the maximum number of messages to process.
	MaxMessages int
}

// DefaultBufferOptions returns default buffer options.
func DefaultBufferOptions() BufferOptions {
	return BufferOptions{
		BufferSize:  100,
		MaxMessages: 1000,
	}
}

// ModelOptions contains options for Claude model configuration.
// Single Responsibility: manages model and context settings.
type ModelOptions struct {
	// Model specifies the Claude model to use.
	Model string

	// PermissionMode sets the permission mode for file operations.
	PermissionMode string

	// ContextFiles specifies files to include in the context.
	ContextFiles []string

	// CustomArgs provides additional arguments for the Claude CLI.
	CustomArgs []string
}

// DefaultModelOptions returns default model options.
func DefaultModelOptions() ModelOptions {
	return ModelOptions{
		Model:          "claude-sonnet-4-5-20250929",
		PermissionMode: "auto",
	}
}

// DebugOptions contains options for debugging and tracing.
// Single Responsibility: manages logging and diagnostic settings.
type DebugOptions struct {
	// Trace enables detailed tracing of protocol messages.
	Trace bool

	// DisableCache disables message caching.
	DisableCache bool

	// CacheTTL sets the cache expiration time.
	CacheTTL string

	// Logger interface for logging operations.
	Logger Logger

	// EnableMetrics enables performance metrics collection.
	EnableMetrics bool
}

// DefaultDebugOptions returns default debug options.
func DefaultDebugOptions() DebugOptions {
	return DebugOptions{
		Trace:         false,
		DisableCache:  false,
		CacheTTL:      "1h",
		EnableMetrics: false,
	}
}

// =============================================================================
// Focused Struct Option Functions
// These operate on the focused option structs (ConnectionOptions, BufferOptions,
// ModelOptions, DebugOptions) for use by packages that embed these structs.
// =============================================================================

// ConnectionOptionFunc is a function that configures ConnectionOptions.
type ConnectionOptionFunc func(*ConnectionOptions)

// WithConnCLIPath sets CLIPath on ConnectionOptions.
func WithConnCLIPath(path string) ConnectionOptionFunc {
	return func(o *ConnectionOptions) { o.CLIPath = path }
}

// WithConnCLICommand sets CLICommand on ConnectionOptions.
func WithConnCLICommand(command string) ConnectionOptionFunc {
	return func(o *ConnectionOptions) { o.CLICommand = command }
}

// WithConnTimeout sets Timeout on ConnectionOptions.
func WithConnTimeout(timeout string) ConnectionOptionFunc {
	return func(o *ConnectionOptions) { o.Timeout = timeout }
}

// WithConnEnv sets Env on ConnectionOptions (merges with existing).
func WithConnEnv(env map[string]string) ConnectionOptionFunc {
	return func(o *ConnectionOptions) {
		if o.Env == nil {
			o.Env = make(map[string]string)
		}
		maps.Copy(o.Env, env)
	}
}

// BufferOptionFunc is a function that configures BufferOptions.
type BufferOptionFunc func(*BufferOptions)

// WithBufBufferSize sets BufferSize on BufferOptions.
func WithBufBufferSize(size int) BufferOptionFunc {
	return func(o *BufferOptions) { o.BufferSize = size }
}

// WithBufMaxMessages sets MaxMessages on BufferOptions.
func WithBufMaxMessages(max int) BufferOptionFunc {
	return func(o *BufferOptions) { o.MaxMessages = max }
}

// ModelOptionFunc is a function that configures ModelOptions.
type ModelOptionFunc func(*ModelOptions)

// WithModelModel sets Model on ModelOptions.
func WithModelModel(model string) ModelOptionFunc {
	return func(o *ModelOptions) { o.Model = model }
}

// WithModelPermissionMode sets PermissionMode on ModelOptions.
func WithModelPermissionMode(mode string) ModelOptionFunc {
	return func(o *ModelOptions) { o.PermissionMode = mode }
}

// WithModelContextFiles sets ContextFiles on ModelOptions.
func WithModelContextFiles(files ...string) ModelOptionFunc {
	return func(o *ModelOptions) { o.ContextFiles = files }
}

// WithModelCustomArgs sets CustomArgs on ModelOptions.
func WithModelCustomArgs(args ...string) ModelOptionFunc {
	return func(o *ModelOptions) { o.CustomArgs = args }
}

// DebugOptionFunc is a function that configures DebugOptions.
type DebugOptionFunc func(*DebugOptions)

// WithDebugTrace sets Trace on DebugOptions.
func WithDebugTrace(trace bool) DebugOptionFunc {
	return func(o *DebugOptions) { o.Trace = trace }
}

// WithDebugDisableCache sets DisableCache on DebugOptions.
func WithDebugDisableCache(disable bool) DebugOptionFunc {
	return func(o *DebugOptions) { o.DisableCache = disable }
}

// WithDebugCacheTTL sets CacheTTL on DebugOptions.
func WithDebugCacheTTL(ttl string) DebugOptionFunc {
	return func(o *DebugOptions) { o.CacheTTL = ttl }
}

// WithDebugLogger sets Logger on DebugOptions.
func WithDebugLogger(logger Logger) DebugOptionFunc {
	return func(o *DebugOptions) { o.Logger = logger }
}

// WithDebugEnableMetrics sets EnableMetrics on DebugOptions.
func WithDebugEnableMetrics(enable bool) DebugOptionFunc {
	return func(o *DebugOptions) { o.EnableMetrics = enable }
}

// =============================================================================
// Small Config Types
// =============================================================================

// OutputFormat represents structured output configuration.
type OutputFormat struct {
	Type   string         `json:"type"` // "json_schema"
	Schema map[string]any `json:"schema"`
}

// PluginConfig represents a plugin configuration.
type PluginConfig struct {
	Type string `json:"type"` // "local"
	Path string `json:"path"`
}

// SdkPluginConfig represents full SDK plugin configuration.
// This provides complete control over plugin behavior including
// timeouts, concurrency limits, and custom configuration.
type SdkPluginConfig struct {
	// Enabled controls whether the plugin is active.
	Enabled bool `json:"enabled"`

	// PluginPath is the filesystem path to the plugin.
	PluginPath string `json:"pluginPath"`

	// Config contains plugin-specific configuration.
	Config map[string]any `json:"config,omitempty"`

	// Timeout is the maximum time allowed for each plugin call.
	// Zero means no timeout.
	Timeout time.Duration `json:"timeout,omitempty"`

	// MaxConcurrent is the maximum number of concurrent plugin calls.
	// Zero means unlimited.
	MaxConcurrent int `json:"maxConcurrent,omitempty"`
}

// SettingSource specifies which settings to load.
type SettingSource string

const (
	SettingSourceUser    SettingSource = "user"
	SettingSourceProject SettingSource = "project"
	SettingSourceLocal   SettingSource = "local"
)

// SdkBeta represents beta features.
type SdkBeta string

const (
	SdkBetaContext1M SdkBeta = "context-1m-2025-08-07"
)

// ExitReason represents session exit reasons.
type ExitReason string

const (
	ExitReasonClear                     ExitReason = "clear"
	ExitReasonLogout                    ExitReason = "logout"
	ExitReasonPromptInputExit           ExitReason = "prompt_input_exit"
	ExitReasonOther                     ExitReason = "other"
	ExitReasonBypassPermissionsDisabled ExitReason = "bypass_permissions_disabled"
)

// SandboxSettings represents sandbox configuration.
type SandboxSettings struct {
	Enabled    bool              `json:"enabled"`
	Type       string            `json:"type,omitempty"` // "docker", "nsjail", etc.
	Image      string            `json:"image,omitempty"`
	Options    map[string]string `json:"options,omitempty"`
	WorkingDir string            `json:"workingDir,omitempty"`
}

// =============================================================================
// Logger Interface
// =============================================================================

// Logger defines the interface for logging operations.
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// NopLogger is a logger that does nothing.
type NopLogger struct{}

func (l *NopLogger) Debugf(format string, args ...any) {}
func (l *NopLogger) Infof(format string, args ...any)  {}
func (l *NopLogger) Warnf(format string, args ...any)  {}
func (l *NopLogger) Errorf(format string, args ...any) {}
