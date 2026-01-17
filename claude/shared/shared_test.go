package shared

import (
	"testing"
)

func TestConstants(t *testing.T) {
	// Test message type constants
	if MessageTypeUser != "user" {
		t.Errorf("MessageTypeUser should be 'user', got %s", MessageTypeUser)
	}
	if MessageTypeAssistant != "assistant" {
		t.Errorf("MessageTypeAssistant should be 'assistant', got %s", MessageTypeAssistant)
	}
	if MessageTypeSystem != "system" {
		t.Errorf("MessageTypeSystem should be 'system', got %s", MessageTypeSystem)
	}
	if MessageTypeResult != "result" {
		t.Errorf("MessageTypeResult should be 'result', got %s", MessageTypeResult)
	}

	// Test content block type constants
	if ContentBlockTypeText != "text" {
		t.Errorf("ContentBlockTypeText should be 'text', got %s", ContentBlockTypeText)
	}
	if ContentBlockTypeThinking != "thinking" {
		t.Errorf("ContentBlockTypeThinking should be 'thinking', got %s", ContentBlockTypeThinking)
	}
	if ContentBlockTypeToolUse != "tool_use" {
		t.Errorf("ContentBlockTypeToolUse should be 'tool_use', got %s", ContentBlockTypeToolUse)
	}
	if ContentBlockTypeToolResult != "tool_result" {
		t.Errorf("ContentBlockTypeToolResult should be 'tool_result', got %s", ContentBlockTypeToolResult)
	}
}

func TestAssistantMessageErrorConstants(t *testing.T) {
	if AssistantMessageErrorAuthFailed != "authentication_failed" {
		t.Errorf("AssistantMessageErrorAuthFailed should be 'authentication_failed', got %s", AssistantMessageErrorAuthFailed)
	}
	if AssistantMessageErrorBilling != "billing_error" {
		t.Errorf("AssistantMessageErrorBilling should be 'billing_error', got %s", AssistantMessageErrorBilling)
	}
	if AssistantMessageErrorRateLimit != "rate_limit" {
		t.Errorf("AssistantMessageErrorRateLimit should be 'rate_limit', got %s", AssistantMessageErrorRateLimit)
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Model == "" {
		t.Error("Default model should not be empty")
	}
	if opts.Timeout == "" {
		t.Error("Default timeout should not be empty")
	}
	if opts.BufferSize <= 0 {
		t.Error("Default buffer size should be positive")
	}
}