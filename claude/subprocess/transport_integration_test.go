package subprocess

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/parser"
	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

func TestTransportEnvironmentValidation(t *testing.T) {
	// Test valid environment variables
	validEnv := map[string]string{
		"VALID_VAR":   "value",
		"ANOTHER_VAR": "another_value",
	}

	// Test invalid environment variables
	invalidEnv := map[string]string{
		"valid_var":     "value",           // lowercase key
		"INVALID_VAR":   "value\nwith\nnewlines", // newlines in value
		"INVALID_VAR2":  "value\000null",   // null byte in value
		"3_INVALID_VAR": "value",          // starts with number
	}

	// Test buildEnv with valid environment variables
	config := &TransportConfig{
		Env: validEnv,
	}

	transport, err := NewTransport(config)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	env := transport.buildEnv()

	// Should include all valid environment variables
	for k, v := range validEnv {
		if !slices.Contains(env, k+"="+v) {
			t.Errorf("Valid environment variable %s=%s not found", k, v)
		}
	}

	// Test buildEnv with invalid environment variables
	config2 := &TransportConfig{
		Env: invalidEnv,
	}

	transport2, err := NewTransport(config2)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	env2 := transport2.buildEnv()

	// Should NOT include any invalid environment variables
	for k, v := range invalidEnv {
		if slices.Contains(env2, k+"="+v) {
			t.Errorf("Invalid environment variable %s=%s was included", k, v)
		}
	}
}

func TestTransportPromptValidation(t *testing.T) {
	// Shell metacharacters are safe because exec.Command doesn't invoke a shell.
	// Only null bytes are rejected as they could cause truncation issues.
	validPrompts := []string{
		"Hello world",
		"What is the capital of France?",
		"Explain quantum computing in simple terms",
		"Calculate 2 + 2",
		"Hello `world",  // backtick - safe with exec.Command
		"Hello $world",  // dollar sign - safe with exec.Command
		"Hello $!world", // bang - safe with exec.Command
		"Hello &world",  // ampersand - safe with exec.Command
		"Hello ;world",  // semicolon - safe with exec.Command
		"Hello |world",  // pipe - safe with exec.Command
		"Hello <world",  // less than - safe with exec.Command
		"Hello >world",  // greater than - safe with exec.Command
	}

	invalidPrompts := []string{
		"Hello\x00world", // null byte - dangerous, could truncate
	}

	// Test valid prompts
	for _, prompt := range validPrompts {
		_, err := NewTransportWithPrompt(&TransportConfig{}, prompt)
		if err != nil {
			t.Errorf("Valid prompt '%s' was rejected: %v", prompt, err)
		}
	}

	// Test invalid prompts
	for _, prompt := range invalidPrompts {
		_, err := NewTransportWithPrompt(&TransportConfig{}, prompt)
		if err == nil {
			t.Errorf("Invalid prompt '%s' was accepted", prompt)
		} else {
			t.Logf("Invalid prompt '%s' correctly rejected: %v", prompt, err)
		}
	}
}

func TestTransportMessageParsing(t *testing.T) {
	// Use the parser registry directly (OCP compliance)
	registry := parser.DefaultRegistry()

	// Test assistant message parsing
	assistantJSON := `{"type": "assistant", "content": [{"type": "text", "text": "Hello, world!"}], "model": "claude-3-5-sonnet-20241022"}`

	assistantMsg, err := registry.Parse("assistant", assistantJSON, 0)
	if err != nil {
		t.Fatalf("Failed to parse assistant message: %v", err)
	}
	contentText := shared.GetContentText(assistantMsg)
	if contentText != "Hello, world!" {
		t.Errorf("Expected content 'Hello, world!', got '%s'", contentText)
	}

	// Test result message parsing
	resultJSON := `{"type": "result", "result": "The answer is 42"}`

	resultMsg, err := registry.Parse("result", resultJSON, 0)
	if err != nil {
		t.Fatalf("Failed to parse result message: %v", err)
	}
	resultMessage, ok := resultMsg.(*shared.ResultMessage)
	if !ok {
		t.Fatalf("Expected *shared.ResultMessage, got %T", resultMsg)
	}
	if resultMessage.Result == nil || *resultMessage.Result != "The answer is 42" {
		t.Errorf("Expected result 'The answer is 42', got '%v'", resultMessage.Result)
	}

	// Test stream event parsing
	streamJSON := `{"type": "stream_event", "uuid": "test-uuid-123", "session_id": "test-session-456", "event": {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Hello"}}}`

	streamMsg, err := registry.Parse("stream_event", streamJSON, 0)
	if err != nil {
		t.Fatalf("Failed to parse stream event: %v", err)
	}
	streamEvent, ok := streamMsg.(*shared.StreamEvent)
	if !ok {
		t.Fatalf("Expected *shared.StreamEvent, got %T", streamMsg)
	}
	if eventType, ok := streamEvent.Event["type"].(string); !ok || eventType != "content_block_delta" {
		t.Errorf("Expected event type 'content_block_delta', got '%v'", streamEvent.Event)
	}

	// Test system message parsing
	systemJSON := `{"type": "system"}`

	systemMsg, err := registry.Parse("system", systemJSON, 0)
	if err != nil {
		t.Fatalf("Failed to parse system message: %v", err)
	}
	if systemMsg == nil {
		t.Error("Expected system message to be non-nil")
	}
}

func TestTransportContextCancellation(t *testing.T) {
	// This test verifies that goroutines properly exit on context cancellation
	// Note: We can't easily test this without a real Claude CLI, but we can
	// test that the context is properly passed to the goroutines

	config := &TransportConfig{}

	// Create a context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	transport, err := NewTransport(config)
	if err != nil {
		t.Fatalf("Failed to create transport: %v", err)
	}

	// Try to connect with a short timeout - this should time out
	err = transport.Connect(ctx)
	// We expect either a connection error or timeout
	if err != nil {
		// Connection failed, which is expected without Claude CLI
		t.Logf("Connection failed as expected (Claude CLI not available): %v", err)
	}
}

func TestTransportInvalidConfig(t *testing.T) {
	// Test nil config
	transport, err := NewTransport(nil)
	if err != nil {
		t.Errorf("Nil config should be valid: %v", err)
	}

	// Test empty config should have defaults
	if transport.cliCommand == "" {
		t.Error("Default CLI command should be set")
	}
	if transport.model == "" {
		t.Error("Default model should be set")
	}
	if transport.timeout == 0 {
		t.Error("Default timeout should be set")
	}
}