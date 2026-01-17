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
