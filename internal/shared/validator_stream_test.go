package shared

import (
	"sync"
	"testing"
)

func TestNewStreamValidator(t *testing.T) {
	v := NewStreamValidator()
	if v == nil {
		t.Fatal("NewStreamValidator returned nil")
	}

	// Check initial state
	stats := v.GetStats()
	if stats.TotalMessages != 0 {
		t.Errorf("Expected 0 total messages, got %d", stats.TotalMessages)
	}
	if stats.PartialMessages != 0 {
		t.Errorf("Expected 0 partial messages, got %d", stats.PartialMessages)
	}
	if stats.Errors != 0 {
		t.Errorf("Expected 0 errors, got %d", stats.Errors)
	}
}

func TestStreamValidatorTrackMessage(t *testing.T) {
	tests := []struct {
		name            string
		messages        []Message
		wantTotal       int
		wantPartial     int
		wantErrors      int
		wantToolReq     int
		wantToolResp    int
		wantPending     int
		wantIssueCount  int
	}{
		{
			name:           "nil message",
			messages:       []Message{nil},
			wantTotal:      1,
			wantPartial:    0,
			wantErrors:     0,
			wantToolReq:    0,
			wantToolResp:   0,
			wantPending:    0,
			wantIssueCount: 0, // nil message doesn't count as an issue, it's tracked as total message
		},
		{
			name: "assistant with tool use",
			messages: []Message{
				&AssistantMessage{
					Content: []ContentBlock{
						&ToolUseBlock{ToolUseID: "tool-1", Name: "bash", Input: map[string]any{"command": "ls"}},
					},
					Model: "claude-test",
				},
			},
			wantTotal:      1,
			wantPartial:    0,
			wantErrors:     0,
			wantToolReq:    1,
			wantToolResp:   0,
			wantPending:    1,
			wantIssueCount: 1, // incomplete_tool
		},
		{
			name: "assistant with tool use and response",
			messages: []Message{
				&AssistantMessage{
					Content: []ContentBlock{
						&ToolUseBlock{ToolUseID: "tool-1", Name: "bash", Input: map[string]any{"command": "ls"}},
					},
					Model: "claude-test",
				},
				&UserMessage{
					Content: []ContentBlock{
						&ToolResultBlock{ToolUseID: "tool-1", Content: "file1.txt"},
					},
				},
			},
			wantTotal:      2,
			wantPartial:    0,
			wantErrors:     0,
			wantToolReq:    1,
			wantToolResp:   1,
			wantPending:    0,
			wantIssueCount: 0,
		},
		{
			name: "stream event increments partial",
			messages: []Message{
				&StreamEvent{
					UUID:      "uuid-1",
					SessionID: "sess-1",
					Event:     map[string]any{"type": "content_block_delta"},
				},
			},
			wantTotal:      1,
			wantPartial:    1,
			wantErrors:     0,
			wantToolReq:    0,
			wantToolResp:   0,
			wantPending:    0,
			wantIssueCount: 0,
		},
		{
			name: "result with error",
			messages: []Message{
				&ResultMessage{
					Subtype:   "error",
					IsError:   true,
					SessionID: "sess-1",
				},
			},
			wantTotal:      1,
			wantPartial:    0,
			wantErrors:     1,
			wantToolReq:    0,
			wantToolResp:   0,
			wantPending:    0,
			wantIssueCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewStreamValidator()

			for _, msg := range tt.messages {
				v.TrackMessage(msg)
			}

			v.MarkStreamEnd()

			stats := v.GetStats()
			if stats.TotalMessages != tt.wantTotal {
				t.Errorf("TotalMessages: got %d, want %d", stats.TotalMessages, tt.wantTotal)
			}
			if stats.PartialMessages != tt.wantPartial {
				t.Errorf("PartialMessages: got %d, want %d", stats.PartialMessages, tt.wantPartial)
			}
			if stats.Errors != tt.wantErrors {
				t.Errorf("Errors: got %d, want %d", stats.Errors, tt.wantErrors)
			}

			pending := v.PendingToolCount()
			if pending != tt.wantPending {
				t.Errorf("PendingToolCount: got %d, want %d", pending, tt.wantPending)
			}

			issues := v.GetIssues()
			if len(issues) != tt.wantIssueCount {
				t.Errorf("GetIssues: got %d issues, want %d. Issues: %+v", len(issues), tt.wantIssueCount, issues)
			}
		})
	}
}

func TestStreamValidatorTrackToolManually(t *testing.T) {
	v := NewStreamValidator()

	// Track request
	v.TrackToolRequest("tool-123")

	// Should have 1 pending
	if v.PendingToolCount() != 1 {
		t.Errorf("Expected 1 pending tool, got %d", v.PendingToolCount())
	}

	// Track result
	v.TrackToolResult("tool-123")

	// Should have 0 pending
	if v.PendingToolCount() != 0 {
		t.Errorf("Expected 0 pending tools, got %d", v.PendingToolCount())
	}

	// Should be complete after marking end
	v.MarkStreamEnd()
	if !v.IsComplete() {
		t.Error("Expected IsComplete to return true")
	}

	// Should have no issues
	issues := v.GetIssues()
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d: %+v", len(issues), issues)
	}
}

func TestStreamValidatorTrackError(t *testing.T) {
	v := NewStreamValidator()

	v.TrackError()
	v.TrackError()
	v.TrackError()

	stats := v.GetStats()
	if stats.Errors != 3 {
		t.Errorf("Expected 3 errors, got %d", stats.Errors)
	}
}

func TestStreamValidatorReset(t *testing.T) {
	v := NewStreamValidator()

	// Add some state
	v.TrackMessage(&AssistantMessage{
		Content: []ContentBlock{
			&ToolUseBlock{ToolUseID: "tool-1", Name: "bash", Input: map[string]any{}},
		},
		Model: "test",
	})
	v.TrackError()
	v.MarkStreamEnd()

	// Verify state
	stats := v.GetStats()
	if stats.TotalMessages != 1 {
		t.Errorf("Before reset: expected 1 message, got %d", stats.TotalMessages)
	}

	// Reset
	v.Reset()

	// Verify reset
	stats = v.GetStats()
	if stats.TotalMessages != 0 {
		t.Errorf("After reset: expected 0 messages, got %d", stats.TotalMessages)
	}
	if stats.Errors != 0 {
		t.Errorf("After reset: expected 0 errors, got %d", stats.Errors)
	}
	if v.PendingToolCount() != 0 {
		t.Errorf("After reset: expected 0 pending tools, got %d", v.PendingToolCount())
	}
	if v.IsComplete() {
		t.Error("After reset: expected IsComplete to return false (stream not ended)")
	}
}

func TestStreamValidatorIsComplete(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(*StreamValidator)
		wantComplete bool
	}{
		{
			name:        "not ended",
			setup:       func(v *StreamValidator) {},
			wantComplete: false,
		},
		{
			name: "ended with no tools",
			setup: func(v *StreamValidator) {
				v.MarkStreamEnd()
			},
			wantComplete: true,
		},
		{
			name: "ended with matched tools",
			setup: func(v *StreamValidator) {
				v.TrackToolRequest("t1")
				v.TrackToolResult("t1")
				v.MarkStreamEnd()
			},
			wantComplete: true,
		},
		{
			name: "ended with unmatched tool",
			setup: func(v *StreamValidator) {
				v.TrackToolRequest("t1")
				v.MarkStreamEnd()
			},
			wantComplete: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewStreamValidator()
			tt.setup(v)

			if got := v.IsComplete(); got != tt.wantComplete {
				t.Errorf("IsComplete: got %v, want %v", got, tt.wantComplete)
			}
		})
	}
}

func TestStreamValidatorIssueTypes(t *testing.T) {
	t.Run("incomplete_tool", func(t *testing.T) {
		v := NewStreamValidator()
		v.TrackToolRequest("orphan-request")
		v.MarkStreamEnd()

		issues := v.GetIssues()
		found := false
		for _, issue := range issues {
			if issue.Type == "incomplete_tool" && issue.UUID == "orphan-request" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected incomplete_tool issue for 'orphan-request', got: %+v", issues)
		}
	})

	t.Run("orphan_result", func(t *testing.T) {
		v := NewStreamValidator()
		v.TrackToolResult("orphan-result")
		v.MarkStreamEnd()

		issues := v.GetIssues()
		found := false
		for _, issue := range issues {
			if issue.Type == "orphan_result" && issue.UUID == "orphan-result" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected orphan_result issue for 'orphan-result', got: %+v", issues)
		}
	})

	t.Run("empty_stream", func(t *testing.T) {
		v := NewStreamValidator()
		v.MarkStreamEnd()

		issues := v.GetIssues()
		found := false
		for _, issue := range issues {
			if issue.Type == "empty_stream" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected empty_stream issue, got: %+v", issues)
		}
	})
}

func TestStreamValidatorConcurrency(t *testing.T) {
	v := NewStreamValidator()
	var wg sync.WaitGroup

	// Concurrent message tracking
	for i := range 100 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			v.TrackMessage(&AssistantMessage{
				Content: []ContentBlock{
					&TextBlock{Text: "hello"},
				},
				Model: "test",
			})
		}(i)
	}

	// Concurrent tool tracking
	for i := range 50 {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			v.TrackToolRequest("concurrent-tool")
		}(i)
		go func(id int) {
			defer wg.Done()
			v.TrackToolResult("concurrent-tool")
		}(i)
	}

	// Concurrent error tracking
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v.TrackError()
		}()
	}

	// Concurrent reads
	for range 20 {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = v.GetStats()
		}()
		go func() {
			defer wg.Done()
			_ = v.GetIssues()
		}()
		go func() {
			defer wg.Done()
			_ = v.PendingToolCount()
		}()
	}

	wg.Wait()

	// Verify no panic and reasonable state
	stats := v.GetStats()
	if stats.TotalMessages != 100 {
		t.Errorf("Expected 100 messages, got %d", stats.TotalMessages)
	}
	if stats.Errors != 10 {
		t.Errorf("Expected 10 errors, got %d", stats.Errors)
	}
}

func TestStreamValidatorProcessingTime(t *testing.T) {
	v := NewStreamValidator()

	// Track a message to ensure time passes
	v.TrackMessage(&AssistantMessage{
		Content: []ContentBlock{&TextBlock{Text: "test"}},
		Model:   "test",
	})

	stats := v.GetStats()
	if stats.ProcessingTime == "" {
		t.Error("ProcessingTime should not be empty")
	}
}
