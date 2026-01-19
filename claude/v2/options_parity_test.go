package v2

import (
	"context"
	"testing"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
	"github.com/stretchr/testify/assert"
)

// TestWithCwd tests the working directory option.
func TestWithCwd(t *testing.T) {
	t.Parallel()

	opts := DefaultSessionOptions()
	WithCwd("/custom/path")(opts)

	assert.Equal(t, "/custom/path", opts.Cwd)
}

// TestWithPromptCwd tests the working directory option for prompts.
func TestWithPromptCwd(t *testing.T) {
	t.Parallel()

	opts := DefaultPromptOptions()
	WithPromptCwd("/custom/path")(opts)

	assert.Equal(t, "/custom/path", opts.Cwd)
}

// TestWithTools_Preset tests the tools preset option.
func TestWithTools_Preset(t *testing.T) {
	t.Parallel()

	opts := DefaultSessionOptions()
	tools := shared.ToolsPreset("claude_code")
	WithTools(tools)(opts)

	assert.NotNil(t, opts.Tools)
	assert.Equal(t, "preset", opts.Tools.Type)
	assert.Equal(t, "claude_code", opts.Tools.Preset)
}

// TestWithTools_Explicit tests the explicit tools option.
func TestWithTools_Explicit(t *testing.T) {
	t.Parallel()

	opts := DefaultSessionOptions()
	tools := shared.ToolsExplicit("Read", "Write", "Bash")
	WithTools(tools)(opts)

	assert.NotNil(t, opts.Tools)
	assert.Equal(t, "explicit", opts.Tools.Type)
	assert.Equal(t, []string{"Read", "Write", "Bash"}, opts.Tools.Tools)
}

// TestWithPromptTools tests the tools option for prompts.
func TestWithPromptTools(t *testing.T) {
	t.Parallel()

	opts := DefaultPromptOptions()
	tools := shared.ToolsPreset("claude_code")
	WithPromptTools(tools)(opts)

	assert.NotNil(t, opts.Tools)
	assert.Equal(t, "preset", opts.Tools.Type)
}

// TestWithStderr tests the stderr callback option.
func TestWithStderr(t *testing.T) {
	t.Parallel()

	opts := DefaultSessionOptions()
	called := false
	callback := func(line string) {
		called = true
	}
	WithStderr(callback)(opts)

	assert.NotNil(t, opts.Stderr)
	// Test the callback
	opts.Stderr("test line")
	assert.True(t, called)
}

// TestWithPromptStderr tests the stderr callback option for prompts.
func TestWithPromptStderr(t *testing.T) {
	t.Parallel()

	opts := DefaultPromptOptions()
	called := false
	callback := func(line string) {
		called = true
	}
	WithPromptStderr(callback)(opts)

	assert.NotNil(t, opts.Stderr)
	opts.Stderr("test line")
	assert.True(t, called)
}

// TestWithCanUseTool tests the permission callback option.
func TestWithCanUseTool(t *testing.T) {
	t.Parallel()

	opts := DefaultSessionOptions()
	callback := func(ctx context.Context, toolName string, toolInput map[string]any, options shared.CanUseToolOptions) (shared.PermissionResult, error) {
		return shared.PermissionResult{Behavior: shared.PermissionBehaviorAllow}, nil
	}
	WithCanUseTool(callback)(opts)

	assert.NotNil(t, opts.CanUseTool)

	// Test the callback
	result, err := opts.CanUseTool(context.Background(), "Read", nil, shared.CanUseToolOptions{})
	assert.NoError(t, err)
	assert.Equal(t, shared.PermissionBehaviorAllow, result.Behavior)
}

// TestWithPromptCanUseTool tests the permission callback option for prompts.
func TestWithPromptCanUseTool(t *testing.T) {
	t.Parallel()

	opts := DefaultPromptOptions()
	callback := func(ctx context.Context, toolName string, toolInput map[string]any, options shared.CanUseToolOptions) (shared.PermissionResult, error) {
		return shared.PermissionResult{Behavior: shared.PermissionBehaviorDeny, Message: "denied"}, nil
	}
	WithPromptCanUseTool(callback)(opts)

	assert.NotNil(t, opts.CanUseTool)

	// Test the callback
	result, err := opts.CanUseTool(context.Background(), "Write", nil, shared.CanUseToolOptions{})
	assert.NoError(t, err)
	assert.Equal(t, shared.PermissionBehaviorDeny, result.Behavior)
	assert.Equal(t, "denied", result.Message)
}

// TestToolsPreset tests the ToolsPreset constructor.
func TestToolsPreset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		preset string
	}{
		{"claude_code", "claude_code"},
		{"custom", "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := shared.ToolsPreset(tt.preset)
			assert.Equal(t, "preset", config.Type)
			assert.Equal(t, tt.preset, config.Preset)
			assert.Nil(t, config.Tools)
		})
	}
}

// TestToolsExplicit tests the ToolsExplicit constructor.
func TestToolsExplicit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tools []string
	}{
		{"single", []string{"Read"}},
		{"multiple", []string{"Read", "Write", "Bash", "Grep"}},
		{"empty", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := shared.ToolsExplicit(tt.tools...)
			assert.Equal(t, "explicit", config.Type)
			assert.Equal(t, "", config.Preset)
			assert.Equal(t, tt.tools, config.Tools)
		})
	}
}

// TestCanUseToolCallback_AllowBehavior tests allow behavior.
func TestCanUseToolCallback_AllowBehavior(t *testing.T) {
	t.Parallel()

	callback := func(_ context.Context, _ string, _ map[string]any, _ shared.CanUseToolOptions) (shared.PermissionResult, error) {
		return shared.PermissionResult{
			Behavior: shared.PermissionBehaviorAllow,
		}, nil
	}

	result, err := callback(context.Background(), "Read", map[string]any{"file_path": "/test"}, shared.CanUseToolOptions{})
	assert.NoError(t, err)
	assert.Equal(t, shared.PermissionBehaviorAllow, result.Behavior)
}

// TestCanUseToolCallback_DenyBehavior tests deny behavior.
func TestCanUseToolCallback_DenyBehavior(t *testing.T) {
	t.Parallel()

	callback := func(_ context.Context, _ string, _ map[string]any, _ shared.CanUseToolOptions) (shared.PermissionResult, error) {
		return shared.PermissionResult{
			Behavior: shared.PermissionBehaviorDeny,
			Message:  "Access denied to this path",
		}, nil
	}

	result, err := callback(context.Background(), "Write", map[string]any{"file_path": "/etc/passwd"}, shared.CanUseToolOptions{})
	assert.NoError(t, err)
	assert.Equal(t, shared.PermissionBehaviorDeny, result.Behavior)
	assert.Equal(t, "Access denied to this path", result.Message)
}

// TestCanUseToolCallback_ModifyInput tests input modification.
func TestCanUseToolCallback_ModifyInput(t *testing.T) {
	t.Parallel()

	callback := func(_ context.Context, _ string, toolInput map[string]any, _ shared.CanUseToolOptions) (shared.PermissionResult, error) {
		// Sanitize the input
		modified := make(map[string]any)
		for k, v := range toolInput {
			modified[k] = v
		}
		modified["sanitized"] = true

		return shared.PermissionResult{
			Behavior:     shared.PermissionBehaviorAllow,
			UpdatedInput: modified,
		}, nil
	}

	result, err := callback(context.Background(), "Bash", map[string]any{"command": "ls"}, shared.CanUseToolOptions{})
	assert.NoError(t, err)
	assert.Equal(t, shared.PermissionBehaviorAllow, result.Behavior)
	assert.NotNil(t, result.UpdatedInput)
	assert.Equal(t, true, result.UpdatedInput["sanitized"])
}

// TestOptionsChaining tests that multiple options can be chained.
func TestOptionsChaining(t *testing.T) {
	t.Parallel()

	opts := DefaultSessionOptions()

	// Apply multiple options
	WithCwd("/project")(opts)
	WithTools(shared.ToolsPreset("claude_code"))(opts)
	WithStderr(func(line string) {})(opts)
	WithModel("claude-sonnet-4-20250514")(opts)

	assert.Equal(t, "/project", opts.Cwd)
	assert.NotNil(t, opts.Tools)
	assert.NotNil(t, opts.Stderr)
	assert.Equal(t, "claude-sonnet-4-20250514", opts.Model)
}
