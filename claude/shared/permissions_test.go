package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionModeConstants(t *testing.T) {
	assert.Equal(t, PermissionMode("default"), PermissionModeDefault)
	assert.Equal(t, PermissionMode("acceptEdits"), PermissionModeAcceptEdits)
	assert.Equal(t, PermissionMode("bypassPermissions"), PermissionModeBypassPermissions)
	assert.Equal(t, PermissionMode("plan"), PermissionModePlan)
	assert.Equal(t, PermissionMode("delegate"), PermissionModeDelegate)
	assert.Equal(t, PermissionMode("dontAsk"), PermissionModeDontAsk)
}

func TestPermissionBehaviorConstants(t *testing.T) {
	assert.Equal(t, PermissionBehavior("allow"), PermissionBehaviorAllow)
	assert.Equal(t, PermissionBehavior("deny"), PermissionBehaviorDeny)
	assert.Equal(t, PermissionBehavior("ask"), PermissionBehaviorAsk)
}

func TestPermissionUpdateDestinationConstants(t *testing.T) {
	assert.Equal(t, PermissionUpdateDestination("userSettings"), PermissionDestUserSettings)
	assert.Equal(t, PermissionUpdateDestination("projectSettings"), PermissionDestProjectSettings)
	assert.Equal(t, PermissionUpdateDestination("localSettings"), PermissionDestLocalSettings)
	assert.Equal(t, PermissionUpdateDestination("session"), PermissionDestSession)
	assert.Equal(t, PermissionUpdateDestination("cliArg"), PermissionDestCLIArg)
}

func TestPermissionResult(t *testing.T) {
	t.Run("allow result", func(t *testing.T) {
		result := PermissionResult{
			Behavior:     PermissionBehaviorAllow,
			UpdatedInput: map[string]any{"key": "value"},
			ToolUseID:    "test-123",
		}
		assert.Equal(t, PermissionBehaviorAllow, result.Behavior)
		assert.Equal(t, "value", result.UpdatedInput["key"])
		assert.Equal(t, "test-123", result.ToolUseID)
	})

	t.Run("deny result", func(t *testing.T) {
		result := PermissionResult{
			Behavior:  PermissionBehaviorDeny,
			Message:   "Permission denied",
			Interrupt: true,
		}
		assert.Equal(t, PermissionBehaviorDeny, result.Behavior)
		assert.Equal(t, "Permission denied", result.Message)
		assert.True(t, result.Interrupt)
	})
}

// TestNewPermissionResultAllow tests the typed allow result constructor.
func TestNewPermissionResultAllow(t *testing.T) {
	t.Parallel()

	t.Run("basic allow", func(t *testing.T) {
		result := NewPermissionResultAllow()
		assert.Equal(t, PermissionBehaviorAllow, result.Behavior)
		assert.Empty(t, result.Message)
		assert.Nil(t, result.UpdatedInput)
		assert.Nil(t, result.UpdatedPermissions)
		assert.False(t, result.Interrupt)
	})

	t.Run("allow with updated input", func(t *testing.T) {
		input := map[string]any{"file_path": "/safe/path", "content": "data"}
		result := NewPermissionResultAllow(WithUpdatedInput(input))

		assert.Equal(t, PermissionBehaviorAllow, result.Behavior)
		assert.Equal(t, "/safe/path", result.UpdatedInput["file_path"])
		assert.Equal(t, "data", result.UpdatedInput["content"])
	})

	t.Run("allow with permission updates", func(t *testing.T) {
		updates := []PermissionUpdate{
			{Type: "addRules", Rules: []PermissionRuleValue{{ToolName: "Write"}}},
		}
		result := NewPermissionResultAllow(WithPermissionUpdates(updates...))

		assert.Equal(t, PermissionBehaviorAllow, result.Behavior)
		assert.Len(t, result.UpdatedPermissions, 1)
		assert.Equal(t, "addRules", result.UpdatedPermissions[0].Type)
	})

	t.Run("allow with tool use ID", func(t *testing.T) {
		result := NewPermissionResultAllow(WithToolUseID("tool-123"))
		assert.Equal(t, "tool-123", result.ToolUseID)
	})

	t.Run("allow with multiple options", func(t *testing.T) {
		result := NewPermissionResultAllow(
			WithUpdatedInput(map[string]any{"key": "value"}),
			WithToolUseID("tool-456"),
		)
		assert.Equal(t, PermissionBehaviorAllow, result.Behavior)
		assert.Equal(t, "value", result.UpdatedInput["key"])
		assert.Equal(t, "tool-456", result.ToolUseID)
	})
}

// TestNewPermissionResultDeny tests the typed deny result constructor.
func TestNewPermissionResultDeny(t *testing.T) {
	t.Parallel()

	t.Run("basic deny", func(t *testing.T) {
		result := NewPermissionResultDeny("Access denied")
		assert.Equal(t, PermissionBehaviorDeny, result.Behavior)
		assert.Equal(t, "Access denied", result.Message)
		assert.False(t, result.Interrupt)
	})

	t.Run("deny with interrupt", func(t *testing.T) {
		result := NewPermissionResultDeny("Critical security violation", WithInterrupt(true))
		assert.Equal(t, PermissionBehaviorDeny, result.Behavior)
		assert.Equal(t, "Critical security violation", result.Message)
		assert.True(t, result.Interrupt)
	})

	t.Run("deny with permission updates", func(t *testing.T) {
		updates := []PermissionUpdate{
			{Type: "setMode", Mode: PermissionModePlan},
		}
		result := NewPermissionResultDeny("Needs review", WithPermissionUpdates(updates...))

		assert.Equal(t, PermissionBehaviorDeny, result.Behavior)
		assert.Equal(t, "Needs review", result.Message)
		assert.Len(t, result.UpdatedPermissions, 1)
	})

	t.Run("deny with tool use ID", func(t *testing.T) {
		result := NewPermissionResultDeny("Not allowed", WithToolUseID("tool-789"))
		assert.Equal(t, "tool-789", result.ToolUseID)
	})
}

// TestPermissionResultOptions tests individual option functions.
func TestPermissionResultOptions(t *testing.T) {
	t.Parallel()

	t.Run("WithUpdatedInput modifies input", func(t *testing.T) {
		result := PermissionResult{}
		WithUpdatedInput(map[string]any{"a": 1, "b": 2})(&result)

		assert.NotNil(t, result.UpdatedInput)
		assert.Equal(t, 1, result.UpdatedInput["a"])
		assert.Equal(t, 2, result.UpdatedInput["b"])
	})

	t.Run("WithPermissionUpdates appends updates", func(t *testing.T) {
		result := PermissionResult{}
		update1 := PermissionUpdate{Type: "addRules"}
		update2 := PermissionUpdate{Type: "removeRules"}
		WithPermissionUpdates(update1, update2)(&result)

		assert.Len(t, result.UpdatedPermissions, 2)
		assert.Equal(t, "addRules", result.UpdatedPermissions[0].Type)
		assert.Equal(t, "removeRules", result.UpdatedPermissions[1].Type)
	})

	t.Run("WithInterrupt sets interrupt flag", func(t *testing.T) {
		result := PermissionResult{}
		WithInterrupt(true)(&result)
		assert.True(t, result.Interrupt)

		WithInterrupt(false)(&result)
		assert.False(t, result.Interrupt)
	})

	t.Run("WithToolUseID sets tool use ID", func(t *testing.T) {
		result := PermissionResult{}
		WithToolUseID("custom-id")(&result)
		assert.Equal(t, "custom-id", result.ToolUseID)
	})
}
