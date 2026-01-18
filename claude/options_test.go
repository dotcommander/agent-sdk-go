package claude

import (
	"testing"

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
