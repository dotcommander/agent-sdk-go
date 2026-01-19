// Package subprocess provides tests for the subprocess transport.
package subprocess

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/parser"
	"github.com/dotcommander/agent-sdk-go/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTransport tests creating a new transport.
func TestNewTransport(t *testing.T) {
	config := &TransportConfig{
		Model:   "claude-sonnet-4-5-20250929",
		Timeout: 30 * time.Second,
	}

	transport, err := NewTransport(config)
	require.NoError(t, err)
	assert.NotNil(t, transport)
	assert.Equal(t, "claude-sonnet-4-5-20250929", transport.model)
	assert.Equal(t, 30*time.Second, transport.timeout)
	assert.Nil(t, transport.promptArg) // Interactive mode
}

// TestNewTransportWithDefaults tests creating a transport with default values.
func TestNewTransportWithDefaults(t *testing.T) {
	config := &TransportConfig{}

	transport, err := NewTransport(config)
	require.NoError(t, err)
	assert.NotNil(t, transport)
	assert.Equal(t, "claude-sonnet-4-5-20250929", transport.model) // Default model
	assert.Equal(t, defaultTimeout, transport.timeout)          // Default timeout
}

// TestNewTransportWithPrompt tests creating a transport for one-shot mode.
func TestNewTransportWithPrompt(t *testing.T) {
	config := &TransportConfig{
		Model:   "claude-sonnet-4-5-20250929",
		Timeout: 30 * time.Second,
	}

	prompt := "What is 2+2?"
	transport, err := NewTransportWithPrompt(config, prompt)
	require.NoError(t, err)
	assert.NotNil(t, transport)
	assert.NotNil(t, transport.promptArg)
	assert.Equal(t, prompt, *transport.promptArg)
}

// TestTransport_buildArgs tests building CLI arguments.
func TestTransport_buildArgs(t *testing.T) {
	tests := []struct {
		name       string
		transport  *Transport
		wantPrefix string
		wantModel  bool
	}{
		{
			name: "interactive mode",
			transport: &Transport{
				promptArg:    nil,
				model:        "claude-sonnet-4-5-20250929",
				systemPrompt: "You are helpful",
				customArgs:   []string{"--debug"},
			},
			wantPrefix: "",
			wantModel:  true,
		},
		{
			name: "one-shot mode",
			transport: &Transport{
				promptArg:    strPtr("What is 2+2?"),
				model:        "claude-sonnet-4-5-20250929",
				systemPrompt: "",
				customArgs:   nil,
			},
			wantPrefix: "",
			wantModel:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := tt.transport.buildArgs()
			assert.NotEmpty(t, args)

			// Check prefix
			if tt.wantPrefix != "" {
				assert.Contains(t, args[0], "prompt")
			}

			// Check model flag
			if tt.wantModel {
				found := false
				for i, arg := range args {
					if arg == "--model" && i+1 < len(args) {
						assert.Equal(t, tt.transport.model, args[i+1])
						found = true
						break
					}
				}
				assert.True(t, found, "model flag not found")
			}

			// Check custom args
			if len(tt.transport.customArgs) > 0 {
				for _, customArg := range tt.transport.customArgs {
					assert.Contains(t, args, customArg)
				}
			}
		})
	}
}

// TestParserRegistry_parseAssistantMessage tests parsing assistant messages via registry.
func TestParserRegistry_parseAssistantMessage(t *testing.T) {
	registry := parser.DefaultRegistry()

	jsonStr := `{"type": "assistant", "content": [{"type": "text", "text": "Hello!"}], "model": "claude-sonnet-4-5-20250929"}`

	msg, err := registry.Parse("assistant", jsonStr, 0)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, shared.MessageTypeAssistant, msg.Type())

	assistantMsg, ok := msg.(*shared.AssistantMessage)
	require.True(t, ok)
	assert.Equal(t, "claude-sonnet-4-5-20250929", assistantMsg.Model)
	assert.Len(t, assistantMsg.Content, 1)

	// Check content text
	contentText := shared.GetContentText(msg)
	assert.Equal(t, "Hello!", contentText)
}

// TestParserRegistry_parseResultMessage tests parsing result messages via registry.
func TestParserRegistry_parseResultMessage(t *testing.T) {
	registry := parser.DefaultRegistry()

	jsonStr := `{"type": "result", "subtype": "success", "result": "42", "duration_ms": 1000, "session_id": "test-session-123", "num_turns": 1, "is_error": false}`

	msg, err := registry.Parse("result", jsonStr, 0)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, shared.MessageTypeResult, msg.Type())

	resultMsg, ok := msg.(*shared.ResultMessage)
	require.True(t, ok)
	assert.Equal(t, "success", resultMsg.Subtype)
	require.NotNil(t, resultMsg.Result)
	assert.Equal(t, "42", *resultMsg.Result)
	assert.Equal(t, "test-session-123", resultMsg.SessionID)
}

// TestParserRegistry_parseStreamEvent tests parsing stream events via registry.
func TestParserRegistry_parseStreamEvent(t *testing.T) {
	registry := parser.DefaultRegistry()

	jsonStr := `{"type": "stream_event", "uuid": "test-uuid-456", "session_id": "test-session-789", "event": {"type": "content_block_delta", "index": 0, "delta": {"type": "text_delta", "text": "Hello"}}}`

	msg, err := registry.Parse("stream_event", jsonStr, 0)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, shared.MessageTypeStreamEvent, msg.Type())

	streamMsg, ok := msg.(*shared.StreamEvent)
	require.True(t, ok)
	assert.Equal(t, "test-uuid-456", streamMsg.UUID)
	assert.Equal(t, "test-session-789", streamMsg.SessionID)
	assert.Equal(t, "content_block_delta", streamMsg.Event["type"])
}

// TestParserRegistry_parseSystemMessage tests parsing system messages via registry.
func TestParserRegistry_parseSystemMessage(t *testing.T) {
	registry := parser.DefaultRegistry()

	jsonStr := `{"type": "system", "subtype": "status"}`

	msg, err := registry.Parse("system", jsonStr, 0)
	require.NoError(t, err)
	assert.NotNil(t, msg)
	assert.Equal(t, shared.MessageTypeSystem, msg.Type())

	systemMsg, ok := msg.(*shared.SystemMessage)
	require.True(t, ok)
	assert.Equal(t, "status", systemMsg.Subtype)
}

// TestTransport_IsConnected tests the IsConnected method.
func TestTransport_IsConnected(t *testing.T) {
	transport := &Transport{
		connected: false,
	}

	assert.False(t, transport.IsConnected())

	transport.mu.Lock()
	transport.connected = true
	transport.mu.Unlock()

	assert.True(t, transport.IsConnected())
}

// TestTransport_GetPID tests the GetPID method.
func TestTransport_GetPID(t *testing.T) {
	transport := &Transport{
		cmd: nil,
	}

	assert.Equal(t, 0, transport.GetPID())

	// Can't test actual PID without starting a real process
}

// TestTransport_Close tests closing an unconnected transport.
func TestTransport_Close_Unconnected(t *testing.T) {
	transport := &Transport{
		connected: false,
	}

	err := transport.Close()
	assert.NoError(t, err)
}

// TestMessageSerialization tests that our messages can be serialized to JSON.
func TestMessageSerialization(t *testing.T) {
	tests := []struct {
		name    string
		message shared.Message
	}{
		{
			name: "user message",
			message: &shared.UserMessage{
				MessageType: shared.MessageTypeUser,
				Content:     "Hello, Claude!",
			},
		},
		{
			name: "assistant message",
			message: &shared.AssistantMessage{
				MessageType: shared.MessageTypeAssistant,
				Content: []shared.ContentBlock{
					&shared.TextBlock{
						MessageType: "text",
						Text:        "Hello! How can I help?",
					},
				},
				Model: "claude-sonnet-4-5-20250929",
			},
		},
		{
			name: "result message",
			message: &shared.ResultMessage{
				MessageType: "result",
				Subtype:     "success",
				Result:      strPtr("Done!"),
				SessionID:   "test-123",
				DurationMs:  500,
				IsError:     false,
				NumTurns:    1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.message)
			require.NoError(t, err)
			assert.NotEmpty(t, data)

			// Verify it can be unmarshaled back
			var raw map[string]any
			err = json.Unmarshal(data, &raw)
			require.NoError(t, err)
			assert.Equal(t, tt.message.Type(), raw["type"])
		})
	}
}

// TestTransport_SendMessage_Unconnected tests sending message when not connected.
func TestTransport_SendMessage_Unconnected(t *testing.T) {
	transport := &Transport{
		connected: false,
	}

	ctx := context.Background()
	err := transport.SendMessage(ctx, "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

// TestTransport_SendMessage_OneShotMode tests sending message in one-shot mode.
func TestTransport_SendMessage_OneShotMode(t *testing.T) {
	transport := &Transport{
		connected: true,
		promptArg: strPtr("test"),
	}

	ctx := context.Background()
	err := transport.SendMessage(ctx, "Test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot send message in one-shot mode")
}

// strPtr returns a pointer to a string.
func strPtr(s string) *string {
	return &s
}
