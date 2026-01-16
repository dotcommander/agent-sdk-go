package shared

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
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

	// Validate model name format
	if !strings.HasPrefix(o.Model, "claude-") {
		return NewConfigurationError("Model", o.Model, "model must start with 'claude-'")
	}

	// Validate permission mode
	validModes := []string{"auto", "read", "write", "restricted"}
	if !slices.Contains(validModes, o.PermissionMode) {
		return NewConfigurationError("PermissionMode", o.PermissionMode,
			fmt.Sprintf("invalid permission mode, must be one of: %v", validModes))
	}

	// Validate timeout format
	if o.Timeout != "" {
		if !strings.HasSuffix(o.Timeout, "s") && !strings.HasSuffix(o.Timeout, "m") && !strings.HasSuffix(o.Timeout, "h") {
			return NewConfigurationError("Timeout", o.Timeout, "timeout must be in format like '30s', '5m', '1h'")
		}
	}

	// Validate context files
	for _, file := range o.ContextFiles {
		if file == "" {
			return NewConfigurationError("ContextFiles", "", "context file path cannot be empty")
		}
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

// BaseOptions contains common options shared across different option types.
// This can be embedded by package-specific option types for composition.
type BaseOptions struct {
	// Model specifies the Claude model to use.
	Model string

	// SystemPrompt sets the system prompt.
	SystemPrompt string

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
}

// DefaultBaseOptions returns default base options.
func DefaultBaseOptions() BaseOptions {
	return BaseOptions{
		Model: "claude-sonnet-4-5-20250929",
	}
}

// Logger defines the interface for logging operations.
type Logger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// CLIChecker defines the interface for checking CLI availability.
// This enables dependency injection for testability - tests can inject
// a mock that always returns true/false without requiring the actual CLI.
type CLIChecker interface {
	IsCLIAvailable() bool
}

// CLICheckerFunc is a function type that implements CLIChecker.
// This allows for easy inline checker creation in tests.
type CLICheckerFunc func() bool

// IsCLIAvailable implements CLIChecker.
func (f CLICheckerFunc) IsCLIAvailable() bool {
	return f()
}

// AlwaysAvailableCLIChecker is a CLIChecker that always returns true.
// Useful for testing without the actual CLI.
type AlwaysAvailableCLIChecker struct{}

// IsCLIAvailable implements CLIChecker.
func (AlwaysAvailableCLIChecker) IsCLIAvailable() bool {
	return true
}

// NeverAvailableCLIChecker is a CLIChecker that always returns false.
// Useful for testing CLI unavailability scenarios.
type NeverAvailableCLIChecker struct{}

// IsCLIAvailable implements CLIChecker.
func (NeverAvailableCLIChecker) IsCLIAvailable() bool {
	return false
}

// NopLogger is a logger that does nothing.
type NopLogger struct{}

func (l *NopLogger) Debugf(format string, args ...any) {}
func (l *NopLogger) Infof(format string, args ...any)  {}
func (l *NopLogger) Warnf(format string, args ...any)  {}
func (l *NopLogger) Errorf(format string, args ...any) {}

// GetDefaultCommand returns the default command name for the current platform.
func GetDefaultCommand() string {
	switch runtime.GOOS {
	case "windows":
		return "claude.exe"
	default:
		return "claude"
	}
}

// GetDefaultPath returns the default installation path for the current platform.
func GetDefaultPath() string {
	switch runtime.GOOS {
	case "darwin":
		// Check Homebrew prefix
		if brewPrefix, err := exec.LookPath("brew"); err == nil {
			output, err := exec.Command(brewPrefix, "--prefix").Output()
			if err == nil {
				prefix := strings.TrimSpace(string(output))
				return filepath.Join(prefix, "bin", GetDefaultCommand())
			}
		}
		// Default to /usr/local/bin
		return filepath.Join("/usr/local/bin", GetDefaultCommand())

	case "linux":
		// Common locations for Linux
		locations := []string{
			filepath.Join("/usr/local/bin", GetDefaultCommand()),
			filepath.Join("/usr/bin", GetDefaultCommand()),
			filepath.Join("/opt/claude/bin", GetDefaultCommand()),
		}

		for _, loc := range locations {
			if _, err := os.Stat(loc); err == nil {
				return loc
			}
		}

	case "windows":
		// Default to Program Files
		return filepath.Join(os.Getenv("ProgramFiles"), "Claude", GetDefaultCommand())
	}

	// Default fallback - just the command name
	return GetDefaultCommand()
}

