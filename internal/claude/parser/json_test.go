package parser

import (
	"testing"

	"agent-sdk-go/internal/claude/shared"
)

func TestParserParseUserMessage(t *testing.T) {
	parser := NewParser()

	jsonStr := `{
		"type": "user",
		"content": "Hello, Claude!"
	}`

	msg, err := parser.ParseMessage(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse UserMessage: %v", err)
	}

	if userMsg, ok := msg.(*shared.UserMessage); !ok {
		t.Fatalf("Expected *UserMessage, got %T", msg)
	} else {
		if userMsg.Type() != shared.MessageTypeUser {
			t.Errorf("Expected type %s, got %s", shared.MessageTypeUser, userMsg.Type())
		}
		if content, ok := userMsg.Content.(string); !ok || content != "Hello, Claude!" {
			t.Errorf("Expected content 'Hello, Claude!', got %v", userMsg.Content)
		}
	}
}

func TestParserParseAssistantMessage(t *testing.T) {
	parser := NewParser()

	jsonStr := `{
		"type": "assistant",
		"content": [
			{
				"type": "text",
				"text": "Hello! I'm Claude."
			}
		],
		"model": "claude-3-5-sonnet-20241022"
	}`

	msg, err := parser.ParseMessage(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse AssistantMessage: %v", err)
	}

	if assistantMsg, ok := msg.(*shared.AssistantMessage); !ok {
		t.Fatalf("Expected *AssistantMessage, got %T", msg)
	} else {
		if assistantMsg.Type() != shared.MessageTypeAssistant {
			t.Errorf("Expected type %s, got %s", shared.MessageTypeAssistant, assistantMsg.Type())
		}
		if assistantMsg.Model != "claude-3-5-sonnet-20241022" {
			t.Errorf("Expected model 'claude-3-5-sonnet-20241022', got %s", assistantMsg.Model)
		}
		if len(assistantMsg.Content) != 1 {
			t.Errorf("Expected 1 content block, got %d", len(assistantMsg.Content))
		}
	}
}

func TestParserParseSystemMessage(t *testing.T) {
	parser := NewParser()

	jsonStr := `{
		"type": "system",
		"subtype": "file_change",
		"file": "test.txt",
		"content": "Hello, world!"
	}`

	msg, err := parser.ParseMessage(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse SystemMessage: %v", err)
	}

	if systemMsg, ok := msg.(*shared.SystemMessage); !ok {
		t.Fatalf("Expected *SystemMessage, got %T", msg)
	} else {
		if systemMsg.Type() != shared.MessageTypeSystem {
			t.Errorf("Expected type %s, got %s", shared.MessageTypeSystem, systemMsg.Type())
		}
		if systemMsg.Subtype != "file_change" {
			t.Errorf("Expected subtype 'file_change', got %s", systemMsg.Subtype)
		}
		if systemMsg.Data["file"] != "test.txt" {
			t.Errorf("Expected data['file'] = 'test.txt', got %v", systemMsg.Data["file"])
		}
	}
}

func TestParserParseResultMessage(t *testing.T) {
	parser := NewParser()

	jsonStr := `{
		"type": "result",
		"subtype": "final_result",
		"duration_ms": 1500,
		"duration_api_ms": 800,
		"is_error": false,
		"num_turns": 2,
		"session_id": "session-123"
	}`

	msg, err := parser.ParseMessage(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse ResultMessage: %v", err)
	}

	if resultMsg, ok := msg.(*shared.ResultMessage); !ok {
		t.Fatalf("Expected *ResultMessage, got %T", msg)
	} else {
		if resultMsg.Type() != shared.MessageTypeResult {
			t.Errorf("Expected type %s, got %s", shared.MessageTypeResult, resultMsg.Type())
		}
		if resultMsg.Subtype != "final_result" {
			t.Errorf("Expected subtype 'final_result', got %s", resultMsg.Subtype)
		}
		if resultMsg.DurationMs != 1500 {
			t.Errorf("Expected duration_ms 1500, got %d", resultMsg.DurationMs)
		}
		if resultMsg.SessionID != "session-123" {
			t.Errorf("Expected session_id 'session-123', got %s", resultMsg.SessionID)
		}
	}
}

func TestParserParseStreamEvent(t *testing.T) {
	parser := NewParser()

	jsonStr := `{
		"type": "stream_event",
		"uuid": "event-123",
		"session_id": "session-123",
		"event": {
			"type": "content_block_delta",
			"index": 0,
			"delta": {
				"text": "Hello"
			}
		}
	}`

	msg, err := parser.ParseMessage(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse StreamEvent: %v", err)
	}

	if streamEvent, ok := msg.(*shared.StreamEvent); !ok {
		t.Fatalf("Expected *StreamEvent, got %T", msg)
	} else {
		if streamEvent.Type() != shared.MessageTypeStreamEvent {
			t.Errorf("Expected type %s, got %s", shared.MessageTypeStreamEvent, streamEvent.Type())
		}
		if streamEvent.UUID != "event-123" {
			t.Errorf("Expected UUID 'event-123', got %s", streamEvent.UUID)
		}
		if streamEvent.SessionID != "session-123" {
			t.Errorf("Expected session_id 'session-123', got %s", streamEvent.SessionID)
		}
	}
}

func TestParserParseMultipleMessages(t *testing.T) {
	parser := NewParser()

	jsonStr := `{"type": "user", "content": "Hello"}{"type": "assistant", "content": [{"type": "text", "text": "Hi!"}]}`

	msgs, err := parser.ParseMessages(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse multiple messages: %v", err)
	}

	if len(msgs) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(msgs))
	}

	if userMsg, ok := msgs[0].(*shared.UserMessage); !ok || userMsg.Content.(string) != "Hello" {
		t.Errorf("First message should be UserMessage with content 'Hello'")
	}

	if assistantMsg, ok := msgs[1].(*shared.AssistantMessage); !ok || len(assistantMsg.Content) != 1 {
		t.Errorf("Second message should be AssistantMessage with 1 content block")
	}
}

func TestParserInvalidJSON(t *testing.T) {
	parser := NewParser()

	_, err := parser.ParseMessage("{invalid json}")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParserUnknownMessageType(t *testing.T) {
	parser := NewParser()

	_, err := parser.ParseMessage(`{"type": "unknown", "data": {}}`)
	if err == nil {
		t.Error("Expected error for unknown message type, got nil")
	}
}

func TestParserBufferManagement(t *testing.T) {
	parser := NewParser()

	// Test that buffer is properly managed
	jsonStr := `{"type": "user", "content": "Hello"}, {"type": "assistant", "content": [{"type": "text", "text": "Hi!"}]}`

	// Parse first message
	msgs, err := parser.ParseMessages(jsonStr)
	if err != nil {
		t.Fatalf("Failed to parse messages: %v", err)
	}

	if len(msgs) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(msgs))
	}

	// Check buffer size is reasonable
	if parser.GetBufferSize() > 1024 {
		t.Errorf("Buffer size too large: %d", parser.GetBufferSize())
	}

	// Check line number is incremented
	if parser.GetLineNumber() <= 0 {
		t.Errorf("Line number should be incremented, got %d", parser.GetLineNumber())
	}
}

func TestParserIncompleteJSON(t *testing.T) {
	parser := NewParser()

	// Incomplete JSON should not return an error, but store in buffer
	incomplete := `{"type": "user", "content": "Hello"`

	msg, err := parser.ParseMessage(incomplete)
	if err != nil {
		t.Fatalf("Failed to parse incomplete JSON: %v", err)
	}

	if msg != nil {
		t.Error("Expected nil message for incomplete JSON, got message")
	}

	// Buffer should contain the incomplete JSON
	if parser.GetBufferSize() == 0 {
		t.Error("Buffer should contain incomplete JSON")
	}

	// Complete JSON should now work - add closing brace
	complete := `}`
	msg, err = parser.ParseMessage(complete)
	if err != nil {
		t.Fatalf("Failed to parse complete JSON: %v", err)
	}

	if userMsg, ok := msg.(*shared.UserMessage); !ok {
		t.Fatalf("Expected *UserMessage, got %T", msg)
	} else {
		if userMsg.Content.(string) != "Hello" {
			t.Errorf("Expected content 'Hello', got %v", userMsg.Content)
		}
	}
}

// BenchmarkParser benchmarks the parser performance.
func BenchmarkParser(b *testing.B) {
	parser := NewParser()

	jsonStr := `{
		"type": "assistant",
		"content": [
			{
				"type": "text",
				"text": "Hello! I'm Claude."
			}
		],
		"model": "claude-3-5-sonnet-20241022"
	}`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := parser.ParseMessage(jsonStr)
			if err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkParserBufferManagement benchmarks buffer management performance.
func BenchmarkParserBufferManagement(b *testing.B) {
	parser := NewParser()

	jsonStr := `{"type": "user", "content": "Hello"}{"type": "assistant", "content": [{"type": "text", "text": "Hi!"}]}`

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := parser.ParseMessages(jsonStr)
			if err != nil {
				b.Error(err)
			}
			parser.Reset()
		}
	})
}