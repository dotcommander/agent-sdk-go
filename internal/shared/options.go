package shared

import (
	"maps"
	"time"
)

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
