package claude

import (
	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

// ClientError represents a client-related error.
type ClientError struct {
	shared.BaseError
	Operation string
}

// Error returns a descriptive error message.
func (e *ClientError) Error() string {
	return shared.NewErrorBuilder("client error").
		InField("in", e.Operation).
		Reason(&e.BaseError).
		Inner(&e.BaseError).
		String()
}

// NewClientError creates a new ClientError.
func NewClientError(operation, reason string, inner error) *ClientError {
	return &ClientError{
		BaseError: shared.BaseError{Reason: reason, Inner: inner},
		Operation: operation,
	}
}

// IsClientError checks if an error is a ClientError.
func IsClientError(err error) bool {
	_, ok := err.(*ClientError)
	return ok
}

// SessionError represents a session-related error.
type SessionError struct {
	shared.BaseError
	SessionID string
}

// Error returns a descriptive error message.
func (e *SessionError) Error() string {
	return shared.NewErrorBuilder("session error").
		Field("session_id", e.SessionID, false).
		Reason(&e.BaseError).
		Inner(&e.BaseError).
		String()
}

// NewSessionError creates a new SessionError.
func NewSessionError(sessionID, reason string, inner error) *SessionError {
	return &SessionError{
		BaseError: shared.BaseError{Reason: reason, Inner: inner},
		SessionID: sessionID,
	}
}

// IsSessionError checks if an error is a SessionError.
func IsSessionError(err error) bool {
	_, ok := err.(*SessionError)
	return ok
}

// QueryError represents a query-related error.
type QueryError struct {
	shared.BaseError
	Prompt string
	Model  string
}

// Error returns a descriptive error message.
func (e *QueryError) Error() string {
	prompt := e.Prompt
	if len(prompt) > 50 {
		prompt = prompt[:50] + "..."
	}
	return shared.NewErrorBuilder("query error").
		Field("prompt", prompt, true).
		Field("model", e.Model, false).
		Reason(&e.BaseError).
		Inner(&e.BaseError).
		String()
}

// NewQueryError creates a new QueryError.
func NewQueryError(prompt, model, reason string, inner error) *QueryError {
	return &QueryError{
		BaseError: shared.BaseError{Reason: reason, Inner: inner},
		Prompt:    prompt,
		Model:     model,
	}
}

// IsQueryError checks if an error is a QueryError.
func IsQueryError(err error) bool {
	_, ok := err.(*QueryError)
	return ok
}

// StreamError represents a stream-related error.
type StreamError struct {
	shared.BaseError
	Prompt string
	Model  string
}

// Error returns a descriptive error message.
func (e *StreamError) Error() string {
	prompt := e.Prompt
	if len(prompt) > 50 {
		prompt = prompt[:50] + "..."
	}
	return shared.NewErrorBuilder("stream error").
		Field("prompt", prompt, true).
		Field("model", e.Model, false).
		Reason(&e.BaseError).
		Inner(&e.BaseError).
		String()
}

// NewStreamError creates a new StreamError.
func NewStreamError(prompt, model, reason string, inner error) *StreamError {
	return &StreamError{
		BaseError: shared.BaseError{Reason: reason, Inner: inner},
		Prompt:    prompt,
		Model:     model,
	}
}

// IsStreamError checks if an error is a StreamError.
func IsStreamError(err error) bool {
	_, ok := err.(*StreamError)
	return ok
}

// ConfigurationError represents a configuration-related error.
type ConfigurationError struct {
	shared.BaseError
	Field string
	Value string
}

// Error returns a descriptive error message.
func (e *ConfigurationError) Error() string {
	return shared.NewErrorBuilder("configuration error").
		Field("field", e.Field, true).
		Field("value", e.Value, true).
		Reason(&e.BaseError).
		Inner(&e.BaseError).
		String()
}

// NewConfigurationError creates a new ConfigurationError.
func NewConfigurationError(field, value, reason string, inner error) *ConfigurationError {
	return &ConfigurationError{
		BaseError: shared.BaseError{Reason: reason, Inner: inner},
		Field:     field,
		Value:     value,
	}
}

// IsConfigurationError checks if an error is a ConfigurationError.
func IsConfigurationError(err error) bool {
	_, ok := err.(*ConfigurationError)
	return ok
}

// AsError converts any error to a more specific error type if possible.
func AsError(err error) error {
	if err == nil {
		return nil
	}

	// Check for shared errors first
	if shared.IsCLINotFound(err) {
		return NewClientError("connect", "Claude CLI not found", err)
	}

	if shared.IsConnectionError(err) {
		return NewClientError("connect", "failed to connect", err)
	}

	if shared.IsTimeoutError(err) {
		return NewClientError("query", "operation timed out", err)
	}

	if shared.IsParserError(err) {
		return NewClientError("parse", "failed to parse message", err)
	}

	if shared.IsProtocolError(err) {
		return NewClientError("protocol", "protocol violation", err)
	}

	// Return as-is if not a specific type
	return err
}
