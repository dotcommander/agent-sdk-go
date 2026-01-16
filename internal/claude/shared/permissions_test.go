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
