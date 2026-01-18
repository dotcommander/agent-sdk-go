package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHookEventConstants(t *testing.T) {
	events := AllHookEvents()
	assert.Len(t, events, 12)
	assert.Contains(t, events, HookEventPreToolUse)
	assert.Contains(t, events, HookEventPostToolUse)
	assert.Contains(t, events, HookEventSessionStart)
	assert.Contains(t, events, HookEventSessionEnd)
}

func TestBaseHookInput(t *testing.T) {
	input := BaseHookInput{
		SessionID:      "sess-123",
		TranscriptPath: "/path/to/transcript",
		Cwd:            "/working/dir",
		PermissionMode: "default",
	}
	assert.Equal(t, "sess-123", input.SessionID)
	assert.Equal(t, "/path/to/transcript", input.TranscriptPath)
	assert.Equal(t, "/working/dir", input.Cwd)
	assert.Equal(t, "default", input.PermissionMode)
}

func TestPreToolUseHookInput(t *testing.T) {
	input := PreToolUseHookInput{
		BaseHookInput: BaseHookInput{
			SessionID: "sess-123",
		},
		HookEventName: "PreToolUse",
		ToolName:      "Bash",
		ToolInput:     map[string]any{"command": "ls"},
		ToolUseID:     "tool-456",
	}
	assert.Equal(t, "PreToolUse", input.HookEventName)
	assert.Equal(t, "Bash", input.ToolName)
	assert.Equal(t, "ls", input.ToolInput["command"])
	assert.Equal(t, "tool-456", input.ToolUseID)
}

func TestHookConfig_MatchesToolName(t *testing.T) {
	tests := []struct {
		name     string
		matcher  string
		toolName string
		want     bool
	}{
		{"empty matcher matches all", "", "Bash", true},
		{"exact match", "Write", "Write", true},
		{"no match", "Write", "Read", false},
		{"regex OR - first", "Write|Edit", "Write", true},
		{"regex OR - second", "Write|Edit", "Edit", true},
		{"regex OR - no match", "Write|Edit", "Read", false},
		{"invalid regex", "[invalid", "anything", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &HookConfig{
				Event:   HookEventPreToolUse,
				Matcher: tt.matcher,
			}
			assert.Equal(t, tt.want, cfg.MatchesToolName(tt.toolName))
		})
	}
}

func TestHookEventMessage(t *testing.T) {
	msg := HookEventMessage{
		Type:           "hook_event",
		HookEventName:  "PreToolUse",
		SessionID:      "sess-123",
		TranscriptPath: "/path/transcript",
		Cwd:            "/working/dir",
		ToolName:       "Write",
		ToolInput:      map[string]any{"file_path": "/test/file.txt"},
		ToolUseID:      "tool-456",
	}
	assert.Equal(t, "hook_event", msg.Type)
	assert.Equal(t, "PreToolUse", msg.HookEventName)
	assert.Equal(t, "Write", msg.ToolName)
}

func TestHookOutput(t *testing.T) {
	resp := HookOutput{
		Type:      "hook_response",
		ToolUseID: "tool-456",
		Continue:  true,
		Decision:  "approve",
	}
	assert.Equal(t, "hook_response", resp.Type)
	assert.True(t, resp.Continue)
	assert.Equal(t, "approve", resp.Decision)
}
