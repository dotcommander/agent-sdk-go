package shared

import (
	"encoding/json"
	"testing"
)

func TestUserMessage(t *testing.T) {
	// Test string content
	userMsg := &UserMessage{
		Content: "Hello, Claude!",
	}

	// Test Type() method
	if userMsg.Type() != MessageTypeUser {
		t.Errorf("Expected MessageTypeUser, got %s", userMsg.Type())
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(userMsg)
	if err != nil {
		t.Fatalf("Failed to marshal UserMessage: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled UserMessage
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal UserMessage: %v", err)
	}

	if unmarshaled.Type() != MessageTypeUser {
		t.Errorf("Unmarshaled message type: expected %s, got %s", MessageTypeUser, unmarshaled.Type())
	}
}

func TestAssistantMessage(t *testing.T) {
	// Test with content blocks
	content := []ContentBlock{
		&TextBlock{Text: "Hello! I'm Claude."},
	}

	assistantMsg := &AssistantMessage{
		Content: content,
		Model:   "claude-3-5-sonnet-20241022",
	}

	// Test Type() method
	if assistantMsg.Type() != MessageTypeAssistant {
		t.Errorf("Expected MessageTypeAssistant, got %s", assistantMsg.Type())
	}

	// Test HasError() with no error
	if assistantMsg.HasError() {
		t.Error("Expected HasError() to return false when no error is set")
	}

	// Test with error
	rateLimitErr := AssistantMessageErrorRateLimit
	assistantMsg.Error = &rateLimitErr
	if !assistantMsg.HasError() {
		t.Error("Expected HasError() to return true when error is set")
	}

	if !assistantMsg.IsRateLimited() {
		t.Error("Expected IsRateLimited() to return true when error is rate limit")
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(assistantMsg)
	if err != nil {
		t.Fatalf("Failed to marshal AssistantMessage: %v", err)
	}

	// Verify type is set in JSON
	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		t.Fatalf("Failed to parse JSON for verification: %v", err)
	}

	if temp["type"] != MessageTypeAssistant {
		t.Errorf("JSON type field: expected %s, got %s", MessageTypeAssistant, temp["type"])
	}
}

func TestSystemMessage(t *testing.T) {
	systemMsg := &SystemMessage{
		Subtype: "file_change",
		Data: map[string]any{
			"file":    "test.txt",
			"content": "Hello, world!",
		},
	}

	// Test Type() method
	if systemMsg.Type() != MessageTypeSystem {
		t.Errorf("Expected MessageTypeSystem, got %s", systemMsg.Type())
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(systemMsg)
	if err != nil {
		t.Fatalf("Failed to marshal SystemMessage: %v", err)
	}

	// Verify type and subtype are preserved in JSON
	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		t.Fatalf("Failed to parse JSON for verification: %v", err)
	}

	if temp["type"] != MessageTypeSystem {
		t.Errorf("JSON type field: expected %s, got %s", MessageTypeSystem, temp["type"])
	}
	if temp["subtype"] != "file_change" {
		t.Errorf("JSON subtype field: expected file_change, got %s", temp["subtype"])
	}

	// Verify data is preserved
	data, ok := temp["file"].(string)
	if !ok || data != "test.txt" {
		t.Errorf("Expected data field to be preserved, got: %v", temp["file"])
	}
}

func TestResultMessage(t *testing.T) {
	resultMsg := &ResultMessage{
		Subtype:     "final_result",
		DurationMs:  1500,
		NumTurns:    2,
		SessionID:  "session-123",
	}

	// Test Type() method
	if resultMsg.Type() != MessageTypeResult {
		t.Errorf("Expected MessageTypeResult, got %s", resultMsg.Type())
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(resultMsg)
	if err != nil {
		t.Fatalf("Failed to marshal ResultMessage: %v", err)
	}

	// Verify type is set in JSON
	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		t.Fatalf("Failed to parse JSON for verification: %v", err)
	}

	if temp["type"] != MessageTypeResult {
		t.Errorf("JSON type field: expected %s, got %s", MessageTypeResult, temp["type"])
	}
}

func TestTextBlock(t *testing.T) {
	textBlock := &TextBlock{Text: "Hello, world!"}

	// Test BlockType() method
	if textBlock.BlockType() != ContentBlockTypeText {
		t.Errorf("Expected ContentBlockTypeText, got %s", textBlock.BlockType())
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(textBlock)
	if err != nil {
		t.Fatalf("Failed to marshal TextBlock: %v", err)
	}

	// Verify type is set in JSON
	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		t.Fatalf("Failed to parse JSON for verification: %v", err)
	}

	if temp["type"] != ContentBlockTypeText {
		t.Errorf("JSON type field: expected %s, got %s", ContentBlockTypeText, temp["type"])
	}
}

func TestThinkingBlock(t *testing.T) {
	thinkingBlock := &ThinkingBlock{
		Thinking:  "I need to analyze this request carefully.",
		Signature: "thinking-123",
	}

	// Test BlockType() method
	if thinkingBlock.BlockType() != ContentBlockTypeThinking {
		t.Errorf("Expected ContentBlockTypeThinking, got %s", thinkingBlock.BlockType())
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(thinkingBlock)
	if err != nil {
		t.Fatalf("Failed to marshal ThinkingBlock: %v", err)
	}

	// Verify type is set in JSON
	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		t.Fatalf("Failed to parse JSON for verification: %v", err)
	}

	if temp["type"] != ContentBlockTypeThinking {
		t.Errorf("JSON type field: expected %s, got %s", ContentBlockTypeThinking, temp["type"])
	}
}

func TestToolUseBlock(t *testing.T) {
	toolUseBlock := &ToolUseBlock{
		ToolUseID: "tool-use-123",
		Name:      "calculator",
		Input:     map[string]any{"expression": "2 + 2"},
	}

	// Test BlockType() method
	if toolUseBlock.BlockType() != ContentBlockTypeToolUse {
		t.Errorf("Expected ContentBlockTypeToolUse, got %s", toolUseBlock.BlockType())
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(toolUseBlock)
	if err != nil {
		t.Fatalf("Failed to marshal ToolUseBlock: %v", err)
	}

	// Verify type is set in JSON
	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		t.Fatalf("Failed to parse JSON for verification: %v", err)
	}

	if temp["type"] != ContentBlockTypeToolUse {
		t.Errorf("JSON type field: expected %s, got %s", ContentBlockTypeToolUse, temp["type"])
	}
}

func TestToolResultBlock(t *testing.T) {
	toolResultBlock := &ToolResultBlock{
		ToolUseID: "tool-use-123",
		Content:    "4",
	}

	// Test BlockType() method
	if toolResultBlock.BlockType() != ContentBlockTypeToolResult {
		t.Errorf("Expected ContentBlockTypeToolResult, got %s", toolResultBlock.BlockType())
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(toolResultBlock)
	if err != nil {
		t.Fatalf("Failed to marshal ToolResultBlock: %v", err)
	}

	// Verify type is set in JSON
	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		t.Fatalf("Failed to parse JSON for verification: %v", err)
	}

	if temp["type"] != ContentBlockTypeToolResult {
		t.Errorf("JSON type field: expected %s, got %s", ContentBlockTypeToolResult, temp["type"])
	}
}

func TestStreamEvent(t *testing.T) {
	streamEvent := &StreamEvent{
		UUID: "event-123",
		Event: map[string]any{
			"type": "content_block_delta",
			"delta": map[string]any{
				"text": "Hello",
			},
		},
	}

	// Test Type() method
	if streamEvent.Type() != MessageTypeStreamEvent {
		t.Errorf("Expected MessageTypeStreamEvent, got %s", streamEvent.Type())
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(streamEvent)
	if err != nil {
		t.Fatalf("Failed to marshal StreamEvent: %v", err)
	}

	// Verify type is set in JSON
	var temp map[string]interface{}
	if err := json.Unmarshal(jsonData, &temp); err != nil {
		t.Fatalf("Failed to parse JSON for verification: %v", err)
	}

	if temp["type"] != MessageTypeStreamEvent {
		t.Errorf("JSON type field: expected %s, got %s", MessageTypeStreamEvent, temp["type"])
	}
}

// TestMessageInterface tests that all message types implement the Message interface.
func TestMessageInterface(t *testing.T) {
	messages := []Message{
		&UserMessage{Content: "test"},
		&AssistantMessage{Content: []ContentBlock{&TextBlock{Text: "test"}}},
		&SystemMessage{Subtype: "test"},
		&ResultMessage{Subtype: "final_result"},
		&StreamEvent{UUID: "test"},
	}

	for _, msg := range messages {
		if msg.Type() == "" {
			t.Errorf("Message %T has empty Type()", msg)
		}
	}
}

// TestContentBlockInterface tests that all content block types implement the ContentBlock interface.
func TestContentBlockInterface(t *testing.T) {
	blocks := []ContentBlock{
		&TextBlock{Text: "test"},
		&ThinkingBlock{Thinking: "test", Signature: "sig"},
		&ToolUseBlock{ToolUseID: "id", Name: "name"},
		&ToolResultBlock{ToolUseID: "id"},
	}

	for _, block := range blocks {
		if block.BlockType() == "" {
			t.Errorf("ContentBlock %T has empty BlockType()", block)
		}
	}
}

// BenchmarkMessageUnmarshaling benchmarks JSON unmarshaling of messages.
func BenchmarkMessageUnmarshaling(b *testing.B) {
	userMsg := &UserMessage{
		Content: "Hello, Claude! How are you today?",
	}

	jsonData, err := json.Marshal(userMsg)
	if err != nil {
		b.Fatalf("Failed to marshal UserMessage: %v", err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var msg UserMessage
			if err := json.Unmarshal(jsonData, &msg); err != nil {
				b.Error(err)
			}
		}
	})
}

// BenchmarkMessageMarshaling benchmarks JSON marshaling of messages.
func BenchmarkMessageMarshaling(b *testing.B) {
	msg := &AssistantMessage{
		Content: []ContentBlock{
			&TextBlock{Text: "Hello! I'm Claude."},
		},
		Model: "claude-3-5-sonnet-20241022",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := json.Marshal(msg)
			if err != nil {
				b.Error(err)
			}
		}
	})
}