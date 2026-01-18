package shared

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
	"os/exec"
	"runtime"
	"strings"
)

// BaseError provides common error functionality that can be embedded
// in domain-specific error types. It handles reason and inner error
// formatting consistently.
type BaseError struct {
	Reason string
	Inner  error
}

// FormatReason appends the reason to a message builder if present.
func (e *BaseError) FormatReason(b *strings.Builder) {
	if e.Reason != "" {
		b.WriteString(": ")
		b.WriteString(e.Reason)
	}
}

// FormatInner appends the inner error to a message builder if present.
func (e *BaseError) FormatInner(b *strings.Builder) {
	if e.Inner != nil {
		b.WriteString(": ")
		b.WriteString(e.Inner.Error())
	}
}

// Unwrap returns the inner error for error chaining (errors.Is/As support).
func (e *BaseError) Unwrap() error {
	return e.Inner
}

// ErrorBuilder provides a fluent interface for building error messages.
// This reduces boilerplate in Error() implementations.
type ErrorBuilder struct {
	b strings.Builder
}

// NewErrorBuilder creates a new error builder starting with the given prefix.
func NewErrorBuilder(prefix string) *ErrorBuilder {
	eb := &ErrorBuilder{}
	eb.b.WriteString(prefix)
	return eb
}

// Field adds a named field to the error message.
// Format: (name=value) or (name="value") if quoted is true.
func (eb *ErrorBuilder) Field(name string, value string, quoted bool) *ErrorBuilder {
	if value != "" {
		eb.b.WriteString(" (")
		eb.b.WriteString(name)
		eb.b.WriteString("=")
		if quoted {
			eb.b.WriteString(`"`)
		}
		eb.b.WriteString(value)
		if quoted {
			eb.b.WriteString(`"`)
		}
		eb.b.WriteString(")")
	}
	return eb
}

// IntField adds a named integer field to the error message.
func (eb *ErrorBuilder) IntField(name string, value int) *ErrorBuilder {
	if value != 0 {
		eb.b.WriteString(" (")
		eb.b.WriteString(name)
		eb.b.WriteString("=")
		eb.b.WriteString(fmt.Sprintf("%d", value))
		eb.b.WriteString(")")
	}
	return eb
}

// InField adds a field in the format " in value" (e.g., " in connect").
// This is for the "client error in {operation}" pattern.
func (eb *ErrorBuilder) InField(prefix string, value string) *ErrorBuilder {
	if value != "" {
		eb.b.WriteString(" ")
		eb.b.WriteString(prefix)
		eb.b.WriteString(" ")
		eb.b.WriteString(value)
	}
	return eb
}

// Reason appends the reason from BaseError if present.
func (eb *ErrorBuilder) Reason(base *BaseError) *ErrorBuilder {
	base.FormatReason(&eb.b)
	return eb
}

// Inner appends the inner error from BaseError if present.
func (eb *ErrorBuilder) Inner(base *BaseError) *ErrorBuilder {
	base.FormatInner(&eb.b)
	return eb
}

// String returns the built error message.
func (eb *ErrorBuilder) String() string {
	return eb.b.String()
}

// CLINotFoundError indicates the Claude CLI executable could not be found.
type CLINotFoundError struct {
	Path        string // Path that was checked
	Command     string // Command that failed to execute
	Suggestions []string
}

// Error returns a descriptive error message for CLINotFoundError.
func (e *CLINotFoundError) Error() string {
	var suggestions string
	if len(e.Suggestions) > 0 {
		suggestions = "\nSuggestions:\n"
		for _, suggestion := range e.Suggestions {
			suggestions += fmt.Sprintf("  - %s\n", suggestion)
		}
	}
	return fmt.Sprintf("Claude CLI not found at %s (command=%q): %s%s", e.Path, e.Command, e.innerError().Error(), suggestions)
}

// innerError returns the underlying OS error.
func (e *CLINotFoundError) innerError() error {
	if e.Command != "" {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("command -v %s", e.Command))
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("command not found: %v (output: %q)", err, string(output))
		}
	}
	return errors.New("no such file or directory")
}

// IsCLINotFound checks if an error is a CLINotFoundError.
func IsCLINotFound(err error) bool {
	_, ok := err.(*CLINotFoundError)
	return ok
}

// NewCLINotFoundError creates a new CLINotFoundError with suggestions.
func NewCLINotFoundError(path, command string) *CLINotFoundError {
	suggestions := []string{
		"Install Claude CLI: https://docs.anthropic.com/claude/docs/quickstart#installing-claude",
		"Add Claude CLI to your PATH: export PATH=\"$PATH:/path/to/claude\"",
		"Verify Claude CLI is executable: chmod +x /path/to/claude",
	}
	switch runtime.GOOS {
	case "darwin":
		suggestions = append(suggestions, "On macOS, Claude CLI is available via Homebrew: brew install claude")
	case "linux":
		suggestions = append(suggestions, "On Linux, download from: https://github.com/anthropics/claude-cli/releases")
	case "windows":
		suggestions = append(suggestions, "On Windows, use PowerShell or Command Prompt and add to PATH")
	}

	return &CLINotFoundError{
		Path:        path,
		Command:     command,
		Suggestions: suggestions,
	}
}

// ConnectionError indicates a failure to connect to or communicate with Claude CLI.
type ConnectionError struct {
	BaseError
}

// Error returns a descriptive error message for ConnectionError.
func (e *ConnectionError) Error() string {
	var b strings.Builder
	b.WriteString("failed to connect to Claude CLI")
	e.FormatReason(&b)
	e.FormatInner(&b)
	return b.String()
}

// NewConnectionError creates a new ConnectionError.
func NewConnectionError(reason string, inner error) *ConnectionError {
	return &ConnectionError{
		BaseError: BaseError{Reason: reason, Inner: inner},
	}
}

// TimeoutError indicates an operation timed out.
type TimeoutError struct {
	BaseError
	Operation string
	Timeout   string
}

// Error returns a descriptive error message for TimeoutError.
func (e *TimeoutError) Error() string {
	var b strings.Builder
	b.WriteString(e.Operation)
	b.WriteString(" timed out after ")
	b.WriteString(e.Timeout)
	return b.String()
}

// NewTimeoutError creates a new TimeoutError.
func NewTimeoutError(operation, timeout string) *TimeoutError {
	return &TimeoutError{
		Operation: operation,
		Timeout:   timeout,
	}
}

// ParserError indicates a JSON parsing failure.
type ParserError struct {
	BaseError
	Line   int    // Line number where parsing failed
	Offset int    // Character offset within the line
	Data   string // Raw data that failed to parse
}

// Error returns a descriptive error message for ParserError.
func (e *ParserError) Error() string {
	var b strings.Builder
	b.WriteString("failed to parse JSON")
	e.FormatReason(&b)
	if e.Line > 0 {
		b.WriteString(fmt.Sprintf(" (line %d", e.Line))
		if e.Offset > 0 {
			b.WriteString(fmt.Sprintf(", offset %d", e.Offset))
		}
		b.WriteString(")")
	}
	if e.Data != "" {
		// Show first 100 characters of data to avoid flooding logs
		data := e.Data
		if len(data) > 100 {
			data = data[:100] + "..."
		}
		b.WriteString(fmt.Sprintf(": %q", data))
	}
	return b.String()
}

// NewParserError creates a new ParserError.
func NewParserError(line, offset int, data, reason string) *ParserError {
	return &ParserError{
		BaseError: BaseError{Reason: reason},
		Line:      line,
		Offset:    offset,
		Data:      data,
	}
}

// ProtocolError indicates a protocol violation or invalid message received.
type ProtocolError struct {
	BaseError
	MessageType string
}

// Error returns a descriptive error message for ProtocolError.
func (e *ProtocolError) Error() string {
	var b strings.Builder
	b.WriteString("protocol error")
	if e.MessageType != "" {
		b.WriteString(fmt.Sprintf(" (type=%q)", e.MessageType))
	}
	e.FormatReason(&b)
	return b.String()
}

// NewProtocolError creates a new ProtocolError.
func NewProtocolError(messageType, reason string) *ProtocolError {
	return &ProtocolError{
		BaseError:   BaseError{Reason: reason},
		MessageType: messageType,
	}
}

// ConfigurationError indicates invalid configuration.
type ConfigurationError struct {
	BaseError
	Field string
	Value string
}

// Error returns a descriptive error message for ConfigurationError.
func (e *ConfigurationError) Error() string {
	var b strings.Builder
	b.WriteString("invalid configuration")
	if e.Field != "" {
		b.WriteString(fmt.Sprintf(" (field=%q", e.Field))
		if e.Value != "" {
			b.WriteString(fmt.Sprintf(", value=%q", e.Value))
		}
		b.WriteString(")")
	}
	e.FormatReason(&b)
	return b.String()
}

// NewConfigurationError creates a new ConfigurationError.
func NewConfigurationError(field, value, reason string) *ConfigurationError {
	return &ConfigurationError{
		BaseError: BaseError{Reason: reason},
		Field:     field,
		Value:     value,
	}
}

// ProcessError indicates a subprocess-related error.
type ProcessError struct {
	BaseError
	PID     int
	Command string
	Signal  string // Signal that caused the process to exit (if applicable)
}

// Error returns a descriptive error message for ProcessError.
func (e *ProcessError) Error() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("process %d failed", e.PID))
	e.FormatReason(&b)
	if e.Command != "" {
		b.WriteString(fmt.Sprintf(" (command=%q)", e.Command))
	}
	if e.Signal != "" {
		b.WriteString(fmt.Sprintf(" (signal=%s)", e.Signal))
	}
	return b.String()
}

// NewProcessError creates a new ProcessError.
func NewProcessError(pid int, command, reason, signal string) *ProcessError {
	return &ProcessError{
		BaseError: BaseError{Reason: reason},
		PID:       pid,
		Command:   command,
		Signal:    signal,
	}
}

// Error creates an error message from a string.
func Error(msg string) error {
	return errors.New(msg)
}

// Errorf creates a formatted error message.
func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// IsConnectionError checks if an error is a ConnectionError.
func IsConnectionError(err error) bool {
	_, ok := err.(*ConnectionError)
	return ok
}

// IsTimeoutError checks if an error is a TimeoutError.
func IsTimeoutError(err error) bool {
	_, ok := err.(*TimeoutError)
	return ok
}

// IsParserError checks if an error is a ParserError.
func IsParserError(err error) bool {
	_, ok := err.(*ParserError)
	return ok
}

// IsProtocolError checks if an error is a ProtocolError.
func IsProtocolError(err error) bool {
	_, ok := err.(*ProtocolError)
	return ok
}

// IsConfigurationError checks if an error is a ConfigurationError.
func IsConfigurationError(err error) bool {
	_, ok := err.(*ConfigurationError)
	return ok
}

// IsProcessError checks if an error is a ProcessError.
func IsProcessError(err error) bool {
	_, ok := err.(*ProcessError)
	return ok
}

// As*Error extraction helpers - use errors.As for wrapped error support

// AsCLINotFoundError extracts a CLINotFoundError from the error chain.
// Returns the error and true if found, nil and false otherwise.
func AsCLINotFoundError(err error) (*CLINotFoundError, bool) {
	var target *CLINotFoundError
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

// AsConnectionError extracts a ConnectionError from the error chain.
// Returns the error and true if found, nil and false otherwise.
func AsConnectionError(err error) (*ConnectionError, bool) {
	var target *ConnectionError
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

// AsTimeoutError extracts a TimeoutError from the error chain.
// Returns the error and true if found, nil and false otherwise.
func AsTimeoutError(err error) (*TimeoutError, bool) {
	var target *TimeoutError
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

// AsParserError extracts a ParserError from the error chain.
// Returns the error and true if found, nil and false otherwise.
func AsParserError(err error) (*ParserError, bool) {
	var target *ParserError
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

// AsProtocolError extracts a ProtocolError from the error chain.
// Returns the error and true if found, nil and false otherwise.
func AsProtocolError(err error) (*ProtocolError, bool) {
	var target *ProtocolError
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

// AsConfigurationError extracts a ConfigurationError from the error chain.
// Returns the error and true if found, nil and false otherwise.
func AsConfigurationError(err error) (*ConfigurationError, bool) {
	var target *ConfigurationError
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

// AsProcessError extracts a ProcessError from the error chain.
// Returns the error and true if found, nil and false otherwise.
func AsProcessError(err error) (*ProcessError, bool) {
	var target *ProcessError
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}

// CircuitBreaker interface defines the contract for a circuit breaker pattern.
// This is a stub implementation that can be completed later.
type CircuitBreaker interface {
	// Execute runs the given function if the circuit breaker allows it.
	Execute(ctx context.Context, fn func() error) error
	// State returns the current state of the circuit breaker.
	State() State
	// Reset manually resets the circuit breaker.
	Reset()
	// RecordFailure records a failure and potentially trips the circuit.
	RecordFailure()
	// RecordSuccess records a success and potentially resets the circuit.
	RecordSuccess()
}

// State represents the state of a circuit breaker.
type State int

const (
	// Closed state means the circuit breaker is operational and requests are allowed.
	Closed State = iota
	// Open state means the circuit breaker is tripped and requests are rejected.
	Open
	// HalfOpen state means the circuit breaker is testing after being open.
	HalfOpen
)

// String returns the string representation of the state.
func (s State) String() string {
	switch s {
	case Closed:
		return "CLOSED"
	case Open:
		return "OPEN"
	case HalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig holds configuration for a circuit breaker.
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of failures before tripping the circuit.
	FailureThreshold int
	// RecoveryTimeout is the time to wait before transitioning to half-open state.
	RecoveryTimeout time.Duration
	// HalfOpenMaxRequests is the maximum requests allowed in half-open state.
	HalfOpenMaxRequests int
}

// DefaultCircuitBreakerConfig returns sensible defaults for the circuit breaker.
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold:     5,
		RecoveryTimeout:      30 * time.Second,
		HalfOpenMaxRequests: 3,
	}
}

// StubCircuitBreaker is a basic circuit breaker implementation.
// This is a stub and can be enhanced with more sophisticated logic.
type StubCircuitBreaker struct {
	config CircuitBreakerConfig
	state  State
	mu     sync.RWMutex

	// Counters
	failures    int
	halfOpenOps int
	lastFailure time.Time
}

// NewStubCircuitBreaker creates a new stub circuit breaker.
func NewStubCircuitBreaker(config CircuitBreakerConfig) *StubCircuitBreaker {
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = DefaultCircuitBreakerConfig().FailureThreshold
	}
	if config.RecoveryTimeout <= 0 {
		config.RecoveryTimeout = DefaultCircuitBreakerConfig().RecoveryTimeout
	}
	if config.HalfOpenMaxRequests <= 0 {
		config.HalfOpenMaxRequests = DefaultCircuitBreakerConfig().HalfOpenMaxRequests
	}

	return &StubCircuitBreaker{
		config: config,
		state:  Closed,
	}
}

// Execute runs the function if the circuit breaker allows it.
func (cb *StubCircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if operation is allowed
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker %s: rejecting request", cb.state)
	}

	// Execute the function
	err := fn()

	if err != nil {
		cb.RecordFailure()
		return fmt.Errorf("circuit breaker: operation failed: %w", err)
	}

	cb.RecordSuccess()
	return nil
}

// canExecute checks if execution is allowed in the current state.
func (cb *StubCircuitBreaker) canExecute() bool {
	switch cb.state {
	case Closed:
		return true
	case HalfOpen:
		cb.halfOpenOps++
		return cb.halfOpenOps <= cb.config.HalfOpenMaxRequests
	case Open:
		// Check if enough time has passed to transition to half-open
		if time.Since(cb.lastFailure) > cb.config.RecoveryTimeout {
			cb.state = HalfOpen
			cb.halfOpenOps = 0
			return true
		}
		return false
	default:
		return false
	}
}

// State returns the current state of the circuit breaker.
func (cb *StubCircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker.
func (cb *StubCircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = Closed
	cb.failures = 0
	cb.halfOpenOps = 0
	cb.lastFailure = time.Time{}
}

// RecordFailure records a failure and potentially trips the circuit.
func (cb *StubCircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	// Trip circuit if threshold reached
	if cb.failures >= cb.config.FailureThreshold {
		cb.state = Open
		cb.failures = 0 // Reset counter for next cycle
	}
}

// RecordSuccess records a success and potentially resets the circuit.
func (cb *StubCircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == HalfOpen {
		// If enough successful operations in half-open, close the circuit
		if cb.halfOpenOps >= cb.config.HalfOpenMaxRequests {
			cb.state = Closed
			cb.failures = 0
			cb.halfOpenOps = 0
		}
	}
	// In other states, just reset the failure counter
	cb.failures = 0
}