package claude

import (
	"errors"
	"strings"
	"testing"
)

func TestClientError(t *testing.T) {
	inner := errors.New("connection refused")
	err := NewClientError("connect", "failed to establish connection", inner)

	msg := err.Error()
	if !strings.Contains(msg, "client error") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, "in connect") {
		t.Error("Error message should contain operation")
	}
	if !strings.Contains(msg, "failed to establish connection") {
		t.Error("Error message should contain reason")
	}
	if !strings.Contains(msg, "connection refused") {
		t.Error("Error message should contain inner error")
	}

	// Test error unwrapping
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find inner error")
	}

	// Test type checker
	if !IsClientError(err) {
		t.Error("IsClientError should return true")
	}
	if IsClientError(errors.New("random")) {
		t.Error("IsClientError should return false for other errors")
	}
}

func TestSessionError(t *testing.T) {
	inner := errors.New("session expired")
	err := NewSessionError("sess_123", "session no longer valid", inner)

	msg := err.Error()
	if !strings.Contains(msg, "session error") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, "session_id=sess_123") {
		t.Error("Error message should contain session ID")
	}
	if !strings.Contains(msg, "session no longer valid") {
		t.Error("Error message should contain reason")
	}
	if !strings.Contains(msg, "session expired") {
		t.Error("Error message should contain inner error")
	}

	// Test error unwrapping
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find inner error")
	}

	// Test type checker
	if !IsSessionError(err) {
		t.Error("IsSessionError should return true")
	}
}

func TestQueryError(t *testing.T) {
	inner := errors.New("rate limited")
	err := NewQueryError("What is the meaning of life?", "claude-3-opus", "request failed", inner)

	msg := err.Error()
	if !strings.Contains(msg, "query error") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, "What is the meaning of life?") {
		t.Error("Error message should contain prompt")
	}
	if !strings.Contains(msg, "model=claude-3-opus") {
		t.Error("Error message should contain model")
	}
	if !strings.Contains(msg, "request failed") {
		t.Error("Error message should contain reason")
	}

	// Test error unwrapping
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find inner error")
	}

	// Test type checker
	if !IsQueryError(err) {
		t.Error("IsQueryError should return true")
	}
}

func TestQueryErrorTruncatesLongPrompt(t *testing.T) {
	longPrompt := strings.Repeat("x", 100)
	err := NewQueryError(longPrompt, "claude-3", "test", nil)

	msg := err.Error()
	// Should truncate to 50 chars + "..."
	if strings.Contains(msg, strings.Repeat("x", 51)) {
		t.Error("Error message should truncate long prompts")
	}
	if !strings.Contains(msg, "...") {
		t.Error("Error message should indicate truncation")
	}
}

func TestStreamError(t *testing.T) {
	inner := errors.New("connection reset")
	err := NewStreamError("Hello Claude", "claude-3-sonnet", "stream interrupted", inner)

	msg := err.Error()
	if !strings.Contains(msg, "stream error") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, "Hello Claude") {
		t.Error("Error message should contain prompt")
	}
	if !strings.Contains(msg, "model=claude-3-sonnet") {
		t.Error("Error message should contain model")
	}
	if !strings.Contains(msg, "stream interrupted") {
		t.Error("Error message should contain reason")
	}

	// Test error unwrapping
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find inner error")
	}

	// Test type checker
	if !IsStreamError(err) {
		t.Error("IsStreamError should return true")
	}
}

func TestConfigurationErrorClaude(t *testing.T) {
	inner := errors.New("invalid format")
	err := NewConfigurationError("APIKey", "", "API key cannot be empty", inner)

	msg := err.Error()
	if !strings.Contains(msg, "configuration error") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, `field="APIKey"`) {
		t.Error("Error message should contain field")
	}
	if !strings.Contains(msg, "API key cannot be empty") {
		t.Error("Error message should contain reason")
	}

	// Test error unwrapping
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find inner error")
	}

	// Test type checker
	if !IsConfigurationError(err) {
		t.Error("IsConfigurationError should return true")
	}
}

func TestAsError(t *testing.T) {
	// Test nil error
	if AsError(nil) != nil {
		t.Error("AsError(nil) should return nil")
	}

	// Test unknown error type (should return as-is)
	unknownErr := errors.New("unknown error")
	if AsError(unknownErr) != unknownErr {
		t.Error("AsError should return unknown errors as-is")
	}
}

// Test that all error types satisfy the error interface
func TestErrorInterface(t *testing.T) {
	var _ error = NewClientError("", "", nil)
	var _ error = NewSessionError("", "", nil)
	var _ error = NewQueryError("", "", "", nil)
	var _ error = NewStreamError("", "", "", nil)
	var _ error = NewConfigurationError("", "", "", nil)
}
