package claude

import (
	"bytes"
	"context"
	"testing"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
	"github.com/stretchr/testify/assert"
)

// TestWithToolsPreset tests the tools preset option.
func TestWithToolsPreset(t *testing.T) {
	t.Parallel()

	t.Run("adds preset flag to custom args", func(t *testing.T) {
		opts := &ClientOptions{}
		WithToolsPreset("claude_code")(opts)

		assert.Contains(t, opts.CustomArgs, "--tools")
		assert.Contains(t, opts.CustomArgs, "preset:claude_code")
	})
}

// TestWithClaudeCodeTools tests the Claude Code tools convenience option.
func TestWithClaudeCodeTools(t *testing.T) {
	t.Parallel()

	t.Run("sets claude_code preset", func(t *testing.T) {
		opts := &ClientOptions{}
		WithClaudeCodeTools()(opts)

		assert.Contains(t, opts.CustomArgs, "--tools")
		assert.Contains(t, opts.CustomArgs, "preset:claude_code")
	})
}

// TestWithDisallowedTools tests the disallowed tools option.
func TestWithDisallowedTools(t *testing.T) {
	t.Parallel()

	t.Run("adds disallowed tools flags", func(t *testing.T) {
		opts := &ClientOptions{}
		WithDisallowedTools("Write", "Edit")(opts)

		count := 0
		for _, arg := range opts.CustomArgs {
			if arg == "--disallowed-tools" {
				count++
			}
		}
		assert.Equal(t, 2, count) // One for each tool
		assert.Contains(t, opts.CustomArgs, "Write")
		assert.Contains(t, opts.CustomArgs, "Edit")
	})
}

// TestWithContinue tests the continue session option.
func TestWithContinue(t *testing.T) {
	t.Parallel()

	t.Run("adds continue flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithContinue()(opts)

		assert.Contains(t, opts.CustomArgs, "--continue")
	})
}

// TestWithResume tests the resume session option.
func TestWithResume(t *testing.T) {
	t.Parallel()

	t.Run("adds resume flag with session ID", func(t *testing.T) {
		opts := &ClientOptions{}
		WithResume("test-session-id")(opts)

		assert.Contains(t, opts.CustomArgs, "--resume")
		assert.Contains(t, opts.CustomArgs, "test-session-id")
	})
}

// TestWithSystemPrompt tests the system prompt option.
func TestWithSystemPrompt(t *testing.T) {
	t.Parallel()

	t.Run("adds system prompt flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithSystemPrompt("You are a helpful assistant.")(opts)

		assert.Contains(t, opts.CustomArgs, "--system-prompt")
		assert.Contains(t, opts.CustomArgs, "You are a helpful assistant.")
	})
}

// TestWithAppendSystemPrompt tests the append system prompt option.
func TestWithAppendSystemPrompt(t *testing.T) {
	t.Parallel()

	t.Run("adds append system prompt flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithAppendSystemPrompt("Focus on Go.")(opts)

		assert.Contains(t, opts.CustomArgs, "--append-system-prompt")
		assert.Contains(t, opts.CustomArgs, "Focus on Go.")
	})
}

// TestWithMaxTurns tests the max turns option.
func TestWithMaxTurns(t *testing.T) {
	t.Parallel()

	t.Run("adds max turns flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithMaxTurns(10)(opts)

		assert.Contains(t, opts.CustomArgs, "--max-turns")
		assert.Contains(t, opts.CustomArgs, "10")
	})
}

// TestWithMaxBudgetUSD tests the max budget option.
func TestWithMaxBudgetUSD(t *testing.T) {
	t.Parallel()

	t.Run("adds max budget flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithMaxBudgetUSD(5.50)(opts)

		assert.Contains(t, opts.CustomArgs, "--max-budget-usd")
		assert.Contains(t, opts.CustomArgs, "5.50")
	})
}

// TestWithMaxThinkingTokens tests the max thinking tokens option.
func TestWithMaxThinkingTokens(t *testing.T) {
	t.Parallel()

	t.Run("adds max thinking tokens flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithMaxThinkingTokens(4096)(opts)

		assert.Contains(t, opts.CustomArgs, "--max-thinking-tokens")
		assert.Contains(t, opts.CustomArgs, "4096")
	})
}

// TestWithWorkingDirectory tests the working directory option.
func TestWithWorkingDirectory(t *testing.T) {
	t.Parallel()

	t.Run("adds cwd flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithWorkingDirectory("/path/to/project")(opts)

		assert.Contains(t, opts.CustomArgs, "--cwd")
		assert.Contains(t, opts.CustomArgs, "/path/to/project")
	})
}

// TestWithAdditionalDirectories tests the additional directories option.
func TestWithAdditionalDirectories(t *testing.T) {
	t.Parallel()

	t.Run("adds multiple dir flags", func(t *testing.T) {
		opts := &ClientOptions{}
		WithAdditionalDirectories("/path/one", "/path/two")(opts)

		count := 0
		for _, arg := range opts.CustomArgs {
			if arg == "--add-dir" {
				count++
			}
		}
		assert.Equal(t, 2, count)
		assert.Contains(t, opts.CustomArgs, "/path/one")
		assert.Contains(t, opts.CustomArgs, "/path/two")
	})
}

// TestWithAgent tests the agent option.
func TestWithAgent(t *testing.T) {
	t.Parallel()

	t.Run("adds agent flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithAgent("code-review")(opts)

		assert.Contains(t, opts.CustomArgs, "--agent")
		assert.Contains(t, opts.CustomArgs, "code-review")
	})
}

// TestWithFileCheckpointing tests the file checkpointing option.
func TestWithFileCheckpointing(t *testing.T) {
	t.Parallel()

	t.Run("adds file checkpointing flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithFileCheckpointing()(opts)

		assert.Contains(t, opts.CustomArgs, "--enable-file-checkpointing")
	})
}

// TestOptionsComposition tests that options compose correctly.
func TestOptionsComposition(t *testing.T) {
	t.Parallel()

	t.Run("multiple options compose", func(t *testing.T) {
		opts := DefaultClientOptions()

		WithModel("claude-opus-4-1-20250805")(opts)
		WithToolsPreset("claude_code")(opts)
		WithMaxTurns(10)(opts)
		WithSystemPrompt("Test prompt")(opts)

		assert.Equal(t, "claude-opus-4-1-20250805", opts.Model)
		assert.Contains(t, opts.CustomArgs, "preset:claude_code")
		assert.Contains(t, opts.CustomArgs, "10")
		assert.Contains(t, opts.CustomArgs, "Test prompt")
	})
}

// =============================================================================
// P0 Parity Options Tests
// =============================================================================

// Note: TestWithAllowedTools is in mcp_test.go alongside the function definition.

// TestWithBetas tests the beta features option.
func TestWithBetas(t *testing.T) {
	t.Parallel()

	t.Run("adds single beta flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithBetas("context-1m-2025-08-07")(opts)

		assert.Contains(t, opts.CustomArgs, "--beta")
		assert.Contains(t, opts.CustomArgs, "context-1m-2025-08-07")
	})

	t.Run("adds multiple beta flags", func(t *testing.T) {
		opts := &ClientOptions{}
		WithBetas("feature-a", "feature-b")(opts)

		// Count --beta flags
		count := 0
		for _, arg := range opts.CustomArgs {
			if arg == "--beta" {
				count++
			}
		}
		assert.Equal(t, 2, count)
		assert.Contains(t, opts.CustomArgs, "feature-a")
		assert.Contains(t, opts.CustomArgs, "feature-b")
	})

	t.Run("empty betas list does nothing", func(t *testing.T) {
		opts := &ClientOptions{}
		WithBetas()(opts)

		assert.NotContains(t, opts.CustomArgs, "--beta")
	})
}

// TestWithCanUseTool tests the permission callback option.
func TestWithCanUseTool(t *testing.T) {
	t.Parallel()

	t.Run("sets callback on options", func(t *testing.T) {
		opts := &ClientOptions{}

		callback := func(ctx context.Context, toolName string, toolInput map[string]any, cbOpts shared.CanUseToolOptions) (shared.PermissionResult, error) {
			return shared.NewPermissionResultAllow(), nil
		}

		WithCanUseTool(callback)(opts)

		assert.NotNil(t, opts.CanUseTool)
	})

	t.Run("nil callback clears existing", func(t *testing.T) {
		opts := &ClientOptions{}
		opts.CanUseTool = func(ctx context.Context, toolName string, toolInput map[string]any, cbOpts shared.CanUseToolOptions) (shared.PermissionResult, error) {
			return shared.NewPermissionResultDeny("test"), nil
		}

		WithCanUseTool(nil)(opts)

		assert.Nil(t, opts.CanUseTool)
	})
}

// TestWithHooks tests the hooks registration option.
func TestWithHooks(t *testing.T) {
	t.Parallel()

	t.Run("sets hooks map on options", func(t *testing.T) {
		opts := &ClientOptions{}

		hooks := map[shared.HookEvent][]shared.HookConfig{
			shared.HookEventPreToolUse: {
				{
					Event: shared.HookEventPreToolUse,
					Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
						return &shared.SyncHookOutput{Continue: true}, nil
					},
				},
			},
		}

		WithHooks(hooks)(opts)

		assert.NotNil(t, opts.Hooks)
		assert.Len(t, opts.Hooks[shared.HookEventPreToolUse], 1)
	})

	t.Run("empty hooks map is valid", func(t *testing.T) {
		opts := &ClientOptions{}
		WithHooks(map[shared.HookEvent][]shared.HookConfig{})(opts)

		assert.NotNil(t, opts.Hooks)
		assert.Len(t, opts.Hooks, 0)
	})
}

// TestWithHook tests the single hook registration option.
func TestWithHook(t *testing.T) {
	t.Parallel()

	t.Run("adds single hook to empty options", func(t *testing.T) {
		opts := &ClientOptions{}

		config := shared.HookConfig{
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		}

		WithHook(shared.HookEventPreToolUse, config)(opts)

		assert.NotNil(t, opts.Hooks)
		assert.Len(t, opts.Hooks[shared.HookEventPreToolUse], 1)
		// Verify event was set on config
		assert.Equal(t, shared.HookEventPreToolUse, opts.Hooks[shared.HookEventPreToolUse][0].Event)
	})

	t.Run("multiple hooks for same event append", func(t *testing.T) {
		opts := &ClientOptions{}

		config1 := shared.HookConfig{
			Matcher: "Read",
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		}
		config2 := shared.HookConfig{
			Matcher: "Write",
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		}

		WithHook(shared.HookEventPreToolUse, config1)(opts)
		WithHook(shared.HookEventPreToolUse, config2)(opts)

		assert.Len(t, opts.Hooks[shared.HookEventPreToolUse], 2)
	})

	t.Run("different events get separate lists", func(t *testing.T) {
		opts := &ClientOptions{}

		preConfig := shared.HookConfig{
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		}
		postConfig := shared.HookConfig{
			Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
				return &shared.SyncHookOutput{Continue: true}, nil
			},
		}

		WithHook(shared.HookEventPreToolUse, preConfig)(opts)
		WithHook(shared.HookEventPostToolUse, postConfig)(opts)

		assert.Len(t, opts.Hooks[shared.HookEventPreToolUse], 1)
		assert.Len(t, opts.Hooks[shared.HookEventPostToolUse], 1)
	})
}

// TestWithPreToolUseHook tests the pre-tool-use convenience option.
func TestWithPreToolUseHook(t *testing.T) {
	t.Parallel()

	t.Run("registers handler for PreToolUse event", func(t *testing.T) {
		opts := &ClientOptions{}

		WithPreToolUseHook(func(ctx context.Context, input *shared.PreToolUseHookInput) (*shared.SyncHookOutput, error) {
			return &shared.SyncHookOutput{Continue: true}, nil
		})(opts)

		assert.NotNil(t, opts.Hooks)
		assert.Len(t, opts.Hooks[shared.HookEventPreToolUse], 1)
	})
}

// TestWithPostToolUseHook tests the post-tool-use convenience option.
func TestWithPostToolUseHook(t *testing.T) {
	t.Parallel()

	t.Run("registers handler for PostToolUse event", func(t *testing.T) {
		opts := &ClientOptions{}

		WithPostToolUseHook(func(ctx context.Context, input *shared.PostToolUseHookInput) (*shared.SyncHookOutput, error) {
			return &shared.SyncHookOutput{Continue: true}, nil
		})(opts)

		assert.NotNil(t, opts.Hooks)
		assert.Len(t, opts.Hooks[shared.HookEventPostToolUse], 1)
	})
}

// TestWithSessionStartHook tests the session start convenience option.
func TestWithSessionStartHook(t *testing.T) {
	t.Parallel()

	t.Run("registers handler for SessionStart event", func(t *testing.T) {
		opts := &ClientOptions{}

		WithSessionStartHook(func(ctx context.Context, input *shared.SessionStartHookInput) (*shared.SyncHookOutput, error) {
			return &shared.SyncHookOutput{Continue: true}, nil
		})(opts)

		assert.NotNil(t, opts.Hooks)
		assert.Len(t, opts.Hooks[shared.HookEventSessionStart], 1)
	})
}

// TestWithSessionEndHook tests the session end convenience option.
func TestWithSessionEndHook(t *testing.T) {
	t.Parallel()

	t.Run("registers handler for SessionEnd event", func(t *testing.T) {
		opts := &ClientOptions{}

		WithSessionEndHook(func(ctx context.Context, input *shared.SessionEndHookInput) (*shared.SyncHookOutput, error) {
			return &shared.SyncHookOutput{Continue: true}, nil
		})(opts)

		assert.NotNil(t, opts.Hooks)
		assert.Len(t, opts.Hooks[shared.HookEventSessionEnd], 1)
	})
}

// TestP0OptionsComposition tests that P0 options compose with existing options.
func TestP0OptionsComposition(t *testing.T) {
	t.Parallel()

	t.Run("all P0 options compose correctly", func(t *testing.T) {
		opts := DefaultClientOptions()

		// Apply all P0 options
		WithModel("claude-sonnet-4")(opts)
		WithAllowedTools("Read", "Write")(opts)
		WithBetas("context-1m-2025-08-07")(opts)
		WithCanUseTool(func(ctx context.Context, toolName string, toolInput map[string]any, cbOpts shared.CanUseToolOptions) (shared.PermissionResult, error) {
			return shared.NewPermissionResultAllow(), nil
		})(opts)
		WithPreToolUseHook(func(ctx context.Context, input *shared.PreToolUseHookInput) (*shared.SyncHookOutput, error) {
			return &shared.SyncHookOutput{Continue: true}, nil
		})(opts)

		// Verify all options are set
		assert.Equal(t, "claude-sonnet-4", opts.Model)
		assert.Contains(t, opts.CustomArgs, "--allowed-tools")
		assert.Contains(t, opts.CustomArgs, "--beta")
		assert.NotNil(t, opts.CanUseTool)
		assert.NotNil(t, opts.Hooks)
		assert.Len(t, opts.Hooks[shared.HookEventPreToolUse], 1)
	})
}

// =============================================================================
// P1 Options Tests
// =============================================================================

// TestWithJSONSchema tests the JSON schema output option.
func TestWithJSONSchema(t *testing.T) {
	t.Parallel()

	t.Run("sets output format with schema", func(t *testing.T) {
		opts := &ClientOptions{}
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		}

		WithJSONSchema(schema)(opts)

		assert.NotNil(t, opts.OutputFormat)
		assert.Equal(t, "json_schema", opts.OutputFormat.Type)
		assert.Equal(t, schema, opts.OutputFormat.Schema)
	})
}

// TestWithOutputFormat tests the output format option.
func TestWithOutputFormat(t *testing.T) {
	t.Parallel()

	t.Run("sets output format directly", func(t *testing.T) {
		opts := &ClientOptions{}
		format := &shared.OutputFormat{
			Type:   "json_schema",
			Schema: map[string]any{"type": "string"},
		}

		WithOutputFormat(format)(opts)

		assert.Equal(t, format, opts.OutputFormat)
	})

	t.Run("nil format clears existing", func(t *testing.T) {
		opts := &ClientOptions{
			OutputFormat: &shared.OutputFormat{Type: "test"},
		}

		WithOutputFormat(nil)(opts)

		assert.Nil(t, opts.OutputFormat)
	})
}

// TestWithCwd tests the working directory alias option.
func TestWithCwd(t *testing.T) {
	t.Parallel()

	t.Run("adds cwd flag (alias)", func(t *testing.T) {
		opts := &ClientOptions{}
		WithCwd("/tmp/test")(opts)

		assert.Contains(t, opts.CustomArgs, "--cwd")
		assert.Contains(t, opts.CustomArgs, "/tmp/test")
	})
}

// TestWithAddDirs tests the additional directories alias option.
func TestWithAddDirs(t *testing.T) {
	t.Parallel()

	t.Run("adds multiple dir flags (alias)", func(t *testing.T) {
		opts := &ClientOptions{}
		WithAddDirs("/data", "/logs")(opts)

		count := 0
		for _, arg := range opts.CustomArgs {
			if arg == "--add-dir" {
				count++
			}
		}
		assert.Equal(t, 2, count)
	})
}

// TestWithForkSession tests the fork session option.
func TestWithForkSession(t *testing.T) {
	t.Parallel()

	t.Run("adds resume and fork flags", func(t *testing.T) {
		opts := &ClientOptions{}
		WithForkSession("session-123")(opts)

		assert.Contains(t, opts.CustomArgs, "--resume")
		assert.Contains(t, opts.CustomArgs, "session-123")
		assert.Contains(t, opts.CustomArgs, "--fork")
	})
}

// TestWithFallbackModel tests the fallback model option.
func TestWithFallbackModel(t *testing.T) {
	t.Parallel()

	t.Run("adds fallback model flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithFallbackModel("claude-sonnet-4")(opts)

		assert.Contains(t, opts.CustomArgs, "--fallback-model")
		assert.Contains(t, opts.CustomArgs, "claude-sonnet-4")
	})
}

// TestWithUser tests the user option.
func TestWithUser(t *testing.T) {
	t.Parallel()

	t.Run("adds user flag", func(t *testing.T) {
		opts := &ClientOptions{}
		WithUser("user-12345")(opts)

		assert.Contains(t, opts.CustomArgs, "--user")
		assert.Contains(t, opts.CustomArgs, "user-12345")
	})
}

// TestWithDebugWriter tests the debug writer option.
func TestWithDebugWriter(t *testing.T) {
	t.Parallel()

	t.Run("sets debug writer", func(t *testing.T) {
		opts := &ClientOptions{}
		var buf bytes.Buffer

		WithDebugWriter(&buf)(opts)

		assert.Equal(t, &buf, opts.DebugWriter)
	})

	t.Run("nil clears debug writer", func(t *testing.T) {
		var buf bytes.Buffer
		opts := &ClientOptions{DebugWriter: &buf}

		WithDebugWriter(nil)(opts)

		assert.Nil(t, opts.DebugWriter)
	})
}

// TestWithStderrCallback tests the stderr callback option.
func TestWithStderrCallback(t *testing.T) {
	t.Parallel()

	t.Run("sets stderr callback", func(t *testing.T) {
		opts := &ClientOptions{}
		called := false
		callback := func(line string) { called = true }

		WithStderrCallback(callback)(opts)

		assert.NotNil(t, opts.StderrCallback)
		opts.StderrCallback("test")
		assert.True(t, called)
	})
}

// TestWithAgents tests the agents option.
func TestWithAgents(t *testing.T) {
	t.Parallel()

	t.Run("sets agents map", func(t *testing.T) {
		opts := &ClientOptions{}
		agents := map[string]shared.AgentDefinition{
			"coder":    {Description: "Code writer", Model: shared.AgentModelSonnet},
			"reviewer": {Description: "Code reviewer", Model: shared.AgentModelOpus},
		}

		WithAgents(agents)(opts)

		assert.Equal(t, agents, opts.Agents)
		assert.Len(t, opts.Agents, 2)
	})
}

// TestWithSettingSources tests the setting sources option.
func TestWithSettingSources(t *testing.T) {
	t.Parallel()

	t.Run("sets setting sources", func(t *testing.T) {
		opts := &ClientOptions{}

		WithSettingSources(shared.SettingSourceUser, shared.SettingSourceProject)(opts)

		assert.Len(t, opts.SettingSources, 2)
		assert.Equal(t, shared.SettingSourceUser, opts.SettingSources[0])
		assert.Equal(t, shared.SettingSourceProject, opts.SettingSources[1])
	})
}

// TestWithSandboxSettings tests the sandbox settings option.
func TestWithSandboxSettings(t *testing.T) {
	t.Parallel()

	t.Run("sets sandbox settings", func(t *testing.T) {
		opts := &ClientOptions{}
		sandbox := &shared.SandboxSettings{
			Enabled:    true,
			Type:       "docker",
			WorkingDir: "/workspace",
		}

		WithSandboxSettings(sandbox)(opts)

		assert.Equal(t, sandbox, opts.Sandbox)
		assert.True(t, opts.Sandbox.Enabled)
		assert.Equal(t, "docker", opts.Sandbox.Type)
	})
}

// TestWithTypedPermissionMode tests the typed permission mode option.
func TestWithTypedPermissionMode(t *testing.T) {
	t.Parallel()

	t.Run("sets permission mode from constant", func(t *testing.T) {
		opts := &ClientOptions{}
		WithTypedPermissionMode(shared.PermissionModeAcceptEdits)(opts)

		assert.Equal(t, "acceptEdits", opts.PermissionMode)
	})

	t.Run("all permission mode constants work", func(t *testing.T) {
		modes := []shared.PermissionMode{
			shared.PermissionModeDefault,
			shared.PermissionModeAcceptEdits,
			shared.PermissionModeBypassPermissions,
			shared.PermissionModePlan,
			shared.PermissionModeDelegate,
			shared.PermissionModeDontAsk,
		}

		for _, mode := range modes {
			opts := &ClientOptions{}
			WithTypedPermissionMode(mode)(opts)
			assert.Equal(t, string(mode), opts.PermissionMode)
		}
	})
}

// TestP1OptionsComposition tests that P1 options compose correctly.
func TestP1OptionsComposition(t *testing.T) {
	t.Parallel()

	t.Run("all P1 options compose correctly", func(t *testing.T) {
		opts := DefaultClientOptions()

		// Apply P1 options
		WithCwd("/project")(opts)
		WithForkSession("session-abc")(opts)
		WithFallbackModel("claude-haiku-3")(opts)
		WithUser("test-user")(opts)
		WithJSONSchema(map[string]any{"type": "object"})(opts)
		WithAgents(map[string]shared.AgentDefinition{
			"agent1": {Description: "Test agent"},
		})(opts)
		WithSandboxSettings(&shared.SandboxSettings{Enabled: true})(opts)

		// Verify
		assert.Contains(t, opts.CustomArgs, "--cwd")
		assert.Contains(t, opts.CustomArgs, "--fork")
		assert.Contains(t, opts.CustomArgs, "--fallback-model")
		assert.Contains(t, opts.CustomArgs, "--user")
		assert.NotNil(t, opts.OutputFormat)
		assert.NotNil(t, opts.Agents)
		assert.NotNil(t, opts.Sandbox)
	})
}
