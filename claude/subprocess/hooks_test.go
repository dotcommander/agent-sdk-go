package subprocess

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHookConfig_MatchesToolName(t *testing.T) {
	tests := []struct {
		name     string
		matcher  string
		toolName string
		want     bool
	}{
		{
			name:     "empty matcher matches all",
			matcher:  "",
			toolName: "Bash",
			want:     true,
		},
		{
			name:     "exact match",
			matcher:  "Write",
			toolName: "Write",
			want:     true,
		},
		{
			name:     "no match",
			matcher:  "Write",
			toolName: "Read",
			want:     false,
		},
		{
			name:     "regex OR match - first option",
			matcher:  "Write|Edit",
			toolName: "Write",
			want:     true,
		},
		{
			name:     "regex OR match - second option",
			matcher:  "Write|Edit",
			toolName: "Edit",
			want:     true,
		},
		{
			name:     "regex OR match - no match",
			matcher:  "Write|Edit",
			toolName: "Read",
			want:     false,
		},
		{
			name:     "regex wildcard",
			matcher:  ".*Tool.*",
			toolName: "MyToolHandler",
			want:     true,
		},
		{
			name:     "invalid regex returns false",
			matcher:  "[invalid",
			toolName: "anything",
			want:     false,
		},
		{
			name:     "case sensitive by default",
			matcher:  "write",
			toolName: "Write",
			want:     false,
		},
		{
			name:     "partial match within word",
			matcher:  "Edit",
			toolName: "MultiEdit",
			want:     true, // regex allows partial matching
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &shared.HookConfig{
				Event:   shared.HookEventPreToolUse,
				Matcher: tt.matcher,
			}
			got := cfg.MatchesToolName(tt.toolName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewHookExecutor(t *testing.T) {
	hooks := []shared.HookConfig{
		{Event: shared.HookEventPreToolUse, Matcher: "Write"},
		{Event: shared.HookEventPreToolUse, Matcher: "Edit"},
		{Event: shared.HookEventPostToolUse, Matcher: ""},
	}

	executor := NewHookExecutor(hooks)

	assert.True(t, executor.HasHooks(shared.HookEventPreToolUse))
	assert.True(t, executor.HasHooks(shared.HookEventPostToolUse))
	assert.False(t, executor.HasHooks(shared.HookEventSessionStart))
}

func TestHookExecutor_ExecuteHook_Continue(t *testing.T) {
	called := false
	hooks := []shared.HookConfig{
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				called = true
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		},
	}

	executor := NewHookExecutor(hooks)
	ctx := context.Background()

	output, err := executor.ExecuteHook(ctx, shared.HookEventPreToolUse, nil, "")
	require.NoError(t, err)
	assert.True(t, called)
	assert.True(t, output.Continue)
}

func TestHookExecutor_ExecuteHook_Block(t *testing.T) {
	hooks := []shared.HookConfig{
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				return &shared.SyncHookOutput{
					Decision: "block",
					Reason:   "not allowed",
				}, nil
			},
		},
	}

	executor := NewHookExecutor(hooks)
	ctx := context.Background()

	output, err := executor.ExecuteHook(ctx, shared.HookEventPreToolUse, nil, "")
	require.NoError(t, err)
	assert.Equal(t, "block", output.Decision)
	assert.Equal(t, "not allowed", output.Reason)
}

func TestHookExecutor_ExecuteHook_ChainExecution(t *testing.T) {
	var callOrder []int
	hooks := []shared.HookConfig{
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				callOrder = append(callOrder, 1)
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		},
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				callOrder = append(callOrder, 2)
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		},
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				callOrder = append(callOrder, 3)
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		},
	}

	executor := NewHookExecutor(hooks)
	ctx := context.Background()

	output, err := executor.ExecuteHook(ctx, shared.HookEventPreToolUse, nil, "")
	require.NoError(t, err)
	assert.True(t, output.Continue)
	assert.Equal(t, []int{1, 2, 3}, callOrder)
}

func TestHookExecutor_ExecuteHook_ShortCircuitOnBlock(t *testing.T) {
	var callOrder []int
	hooks := []shared.HookConfig{
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				callOrder = append(callOrder, 1)
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		},
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				callOrder = append(callOrder, 2)
				return &shared.SyncHookOutput{Decision: "block"}, nil
			},
		},
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				callOrder = append(callOrder, 3) // Should not be called
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		},
	}

	executor := NewHookExecutor(hooks)
	ctx := context.Background()

	output, err := executor.ExecuteHook(ctx, shared.HookEventPreToolUse, nil, "")
	require.NoError(t, err)
	assert.Equal(t, "block", output.Decision)
	assert.Equal(t, []int{1, 2}, callOrder) // Third hook not called
}

func TestHookExecutor_ExecuteHook_MatcherFiltering(t *testing.T) {
	var calledTools []string
	hooks := []shared.HookConfig{
		{
			Event:   shared.HookEventPreToolUse,
			Matcher: "Write|Edit",
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				if preInput, ok := input.(*shared.PreToolUseHookInput); ok {
					calledTools = append(calledTools, preInput.ToolName)
				}
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		},
	}

	executor := NewHookExecutor(hooks)
	ctx := context.Background()

	// Create typed inputs for different tools
	writeInput := &shared.PreToolUseHookInput{ToolName: "Write"}
	editInput := &shared.PreToolUseHookInput{ToolName: "Edit"}
	readInput := &shared.PreToolUseHookInput{ToolName: "Read"}

	// Execute for Write - should match
	_, _ = executor.ExecuteHook(ctx, shared.HookEventPreToolUse, writeInput, "Write")

	// Execute for Edit - should match
	_, _ = executor.ExecuteHook(ctx, shared.HookEventPreToolUse, editInput, "Edit")

	// Execute for Read - should NOT match
	_, _ = executor.ExecuteHook(ctx, shared.HookEventPreToolUse, readInput, "Read")

	assert.Equal(t, []string{"Write", "Edit"}, calledTools)
}

func TestHookExecutor_ExecuteHook_Timeout(t *testing.T) {
	hooks := []shared.HookConfig{
		{
			Event:   shared.HookEventPreToolUse,
			Timeout: 50 * time.Millisecond,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				// Sleep longer than timeout
				select {
				case <-time.After(200 * time.Millisecond):
					return &shared.SyncHookOutput{Continue: true}, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
		},
	}

	executor := NewHookExecutor(hooks)
	ctx := context.Background()

	output, err := executor.ExecuteHook(ctx, shared.HookEventPreToolUse, nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
	assert.Equal(t, "block", output.Decision) // Fail-closed
}

func TestHookExecutor_ExecuteHook_PanicRecovery(t *testing.T) {
	hooks := []shared.HookConfig{
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				panic("intentional panic for testing")
			},
		},
	}

	executor := NewHookExecutor(hooks)
	ctx := context.Background()

	output, err := executor.ExecuteHook(ctx, shared.HookEventPreToolUse, nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic")
	assert.Equal(t, "block", output.Decision) // Fail-closed
}

func TestHookExecutor_ExecuteHook_ErrorHandling(t *testing.T) {
	hooks := []shared.HookConfig{
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				return nil, errors.New("handler error")
			},
		},
	}

	executor := NewHookExecutor(hooks)
	ctx := context.Background()

	output, err := executor.ExecuteHook(ctx, shared.HookEventPreToolUse, nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "handler error")
	assert.Equal(t, "block", output.Decision) // Fail-closed
}

func TestHookExecutor_ExecuteHook_NoHooksRegistered(t *testing.T) {
	executor := NewHookExecutor(nil)
	ctx := context.Background()

	output, err := executor.ExecuteHook(ctx, shared.HookEventPreToolUse, nil, "")
	require.NoError(t, err)
	assert.True(t, output.Continue) // Default to continue
}

func TestHookExecutor_RegisterHook(t *testing.T) {
	executor := NewHookExecutor(nil)
	assert.False(t, executor.HasHooks(shared.HookEventPreToolUse))

	executor.RegisterHook(shared.HookConfig{
		Event: shared.HookEventPreToolUse,
		Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
			return &shared.SyncHookOutput{Continue: true}, nil
		},
	})

	assert.True(t, executor.HasHooks(shared.HookEventPreToolUse))
}

func TestHookExecutor_SetTimeout(t *testing.T) {
	executor := NewHookExecutor(nil)
	executor.SetTimeout(5 * time.Second)

	// Can't easily test internal state, but verify it doesn't panic
	assert.NotNil(t, executor)
}

func TestBuildTypedInput(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		wantType  string
	}{
		{"PreToolUse", "PreToolUse", "*shared.PreToolUseHookInput"},
		{"PostToolUse", "PostToolUse", "*shared.PostToolUseHookInput"},
		{"PostToolUseFailure", "PostToolUseFailure", "*shared.PostToolUseFailureHookInput"},
		{"SessionStart", "SessionStart", "*shared.SessionStartHookInput"},
		{"SessionEnd", "SessionEnd", "*shared.SessionEndHookInput"},
		{"PermissionRequest", "PermissionRequest", "*shared.PermissionRequestHookInput"},
		{"Notification", "Notification", "*shared.NotificationHookInput"},
		{"UserPromptSubmit", "UserPromptSubmit", "*shared.UserPromptSubmitHookInput"},
		{"Stop", "Stop", "*shared.StopHookInput"},
		{"SubagentStart", "SubagentStart", "*shared.SubagentStartHookInput"},
		{"SubagentStop", "SubagentStop", "*shared.SubagentStopHookInput"},
		{"PreCompact", "PreCompact", "*shared.PreCompactHookInput"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &shared.HookEventMessage{
				Type:           "hook_event",
				HookEventName:  tt.eventName,
				SessionID:      "test-session",
				TranscriptPath: "/path/to/transcript",
				Cwd:            "/working/dir",
				ToolName:       "TestTool",
				ToolInput:      map[string]any{"key": "value"},
				ToolUseID:      "tool-123",
			}

			input, err := BuildTypedInput(msg)
			require.NoError(t, err)
			assert.NotNil(t, input)

			// Verify common fields are populated
			switch v := input.(type) {
			case *shared.PreToolUseHookInput:
				assert.Equal(t, "test-session", v.SessionID)
				assert.Equal(t, "TestTool", v.ToolName)
			case *shared.PostToolUseHookInput:
				assert.Equal(t, "test-session", v.SessionID)
				assert.Equal(t, "TestTool", v.ToolName)
			}
		})
	}
}

func TestBuildTypedInput_UnknownEvent(t *testing.T) {
	msg := &shared.HookEventMessage{
		Type:          "hook_event",
		HookEventName: "UnknownEvent",
	}

	_, err := BuildTypedInput(msg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown hook event type")
}

func TestBuildHookResponse(t *testing.T) {
	t.Run("nil output returns continue", func(t *testing.T) {
		resp := BuildHookResponse(nil, "tool-123")
		assert.Equal(t, "hook_response", resp.Type)
		assert.Equal(t, "tool-123", resp.ToolUseID)
		assert.True(t, resp.Continue)
	})

	t.Run("block decision", func(t *testing.T) {
		output := &shared.SyncHookOutput{
			Decision: "block",
			Reason:   "not allowed",
		}
		resp := BuildHookResponse(output, "tool-456")
		assert.Equal(t, "block", resp.Decision)
		assert.Equal(t, "not allowed", resp.Reason)
	})

	t.Run("approve with all fields", func(t *testing.T) {
		output := &shared.SyncHookOutput{
			Continue:       true,
			SuppressOutput: true,
			Decision:       "approve",
			StopReason:     "user requested",
			SystemMessage:  "system msg",
			Reason:         "approved reason",
		}
		resp := BuildHookResponse(output, "tool-789")
		assert.True(t, resp.Continue)
		assert.True(t, resp.SuppressOutput)
		assert.Equal(t, "approve", resp.Decision)
		assert.Equal(t, "user requested", resp.StopReason)
		assert.Equal(t, "system msg", resp.SystemMessage)
		assert.Equal(t, "approved reason", resp.Reason)
	})
}

func TestHookExecutor_ContextCancellation(t *testing.T) {
	hooks := []shared.HookConfig{
		{
			Event: shared.HookEventPreToolUse,
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				// Wait for context cancellation
				<-ctx.Done()
				return nil, ctx.Err()
			},
		},
	}

	executor := NewHookExecutor(hooks)
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	output, err := executor.ExecuteHook(ctx, shared.HookEventPreToolUse, nil, "")
	require.Error(t, err)
	assert.Equal(t, "block", output.Decision) // Fail-closed
}
