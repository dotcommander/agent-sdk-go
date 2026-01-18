// Package main demonstrates hook lifecycle patterns with the Claude Agent SDK.
//
// This example shows:
// - All 12 hook event types and their inputs
// - Hook configuration patterns
// - Pre/Post tool use hooks for validation
// - Session lifecycle hooks
// - Permission request handling
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world
package main

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

func main() {
	fmt.Println("=== Hook Lifecycle Example ===")
	fmt.Println("This demonstrates all 12 hook event types in the SDK.")
	fmt.Println()

	// Show all available hook events
	demonstrateHookEvents()

	// Show hook input structures
	demonstratePreToolUseHook()
	demonstratePostToolUseHook()
	demonstrateSessionLifecycleHooks()
	demonstrateSubagentHooks()
	demonstratePermissionHook()

	// Show hook output patterns
	demonstrateHookOutputs()

	fmt.Println()
	fmt.Println("=== Hook Lifecycle Example Complete ===")
}

// demonstrateHookEvents lists all available hook events.
func demonstrateHookEvents() {
	fmt.Println("--- Available Hook Events ---")

	events := shared.AllHookEvents()
	for i, event := range events {
		fmt.Printf("  %2d. %s\n", i+1, event)
	}
	fmt.Println()
}

// demonstratePreToolUseHook shows PreToolUse hook input structure.
func demonstratePreToolUseHook() {
	fmt.Println("--- PreToolUse Hook ---")
	fmt.Println("Called BEFORE a tool executes. Can block, modify, or approve.")
	fmt.Println()

	input := shared.PreToolUseHookInput{
		BaseHookInput: shared.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user/project",
			PermissionMode: "default",
		},
		HookEventName: string(shared.HookEventPreToolUse),
		ToolName:      "Write",
		ToolInput: map[string]any{
			"file_path": "/home/user/project/main.go",
			"content":   "package main\n\nfunc main() {}\n",
		},
		ToolUseID: "tool-use-abc123",
	}

	printJSON("Input", input)

	// Example: File path validation hook
	fmt.Println("Example: Block writes outside project directory")
	fmt.Print(`
  filePath := input.ToolInput["file_path"].(string)
  if !strings.HasPrefix(filePath, input.Cwd) {
      return shared.SyncHookOutput{
          Decision:   "block",
          StopReason: "Cannot write outside project directory",
      }
  }
  return shared.SyncHookOutput{Continue: true}
`)
	fmt.Println()
}

// demonstratePostToolUseHook shows PostToolUse and PostToolUseFailure hooks.
func demonstratePostToolUseHook() {
	fmt.Println("--- PostToolUse Hook ---")
	fmt.Println("Called AFTER a tool executes successfully. For logging, metrics, side effects.")
	fmt.Println()

	input := shared.PostToolUseHookInput{
		BaseHookInput: shared.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user/project",
		},
		HookEventName: string(shared.HookEventPostToolUse),
		ToolName:      "Read",
		ToolInput: map[string]any{
			"file_path": "/home/user/project/config.yaml",
		},
		ToolResponse: map[string]any{
			"content": "server:\n  port: 8080\n",
			"lines":   2,
		},
		ToolUseID: "tool-use-def456",
	}

	printJSON("Input", input)

	fmt.Println("--- PostToolUseFailure Hook ---")
	fmt.Println("Called when a tool execution fails. For error tracking, recovery.")
	fmt.Println()

	failureInput := shared.PostToolUseFailureHookInput{
		BaseHookInput: shared.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user/project",
		},
		HookEventName: string(shared.HookEventPostToolUseFailure),
		ToolName:      "Bash",
		ToolInput: map[string]any{
			"command": "rm -rf /important",
		},
		ToolUseID:   "tool-use-ghi789",
		Error:       "Permission denied",
		IsInterrupt: false,
	}

	printJSON("Input", failureInput)
	fmt.Println()
}

// demonstrateSessionLifecycleHooks shows session start/end hooks.
func demonstrateSessionLifecycleHooks() {
	fmt.Println("--- Session Lifecycle Hooks ---")
	fmt.Println()

	fmt.Println("SessionStart: Called when a session begins")
	startInput := shared.SessionStartHookInput{
		BaseHookInput: shared.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user/project",
		},
		HookEventName: string(shared.HookEventSessionStart),
		Source:        "startup", // "startup" | "resume" | "clear" | "compact"
		AgentType:     "claude-sonnet-4",
		Model:         "claude-sonnet-4-20250514",
	}
	printJSON("SessionStart Input", startInput)

	fmt.Println("SessionEnd: Called when a session ends")
	endInput := shared.SessionEndHookInput{
		BaseHookInput: shared.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user/project",
		},
		HookEventName: string(shared.HookEventSessionEnd),
		Reason:        "user_exit", // or "error", "timeout", etc.
	}
	printJSON("SessionEnd Input", endInput)

	fmt.Println("Stop: Called when execution is stopping")
	stopInput := shared.StopHookInput{
		BaseHookInput: shared.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user/project",
		},
		HookEventName:  string(shared.HookEventStop),
		StopHookActive: true,
	}
	printJSON("Stop Input", stopInput)
	fmt.Println()
}

// demonstrateSubagentHooks shows subagent start/stop hooks.
func demonstrateSubagentHooks() {
	fmt.Println("--- Subagent Hooks ---")
	fmt.Println()

	fmt.Println("SubagentStart: Called when a subagent is spawned")
	startInput := shared.SubagentStartHookInput{
		BaseHookInput: shared.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user/project",
		},
		HookEventName: string(shared.HookEventSubagentStart),
		AgentID:       "agent-xyz789",
		AgentType:     "code-review",
	}
	printJSON("SubagentStart Input", startInput)

	fmt.Println("SubagentStop: Called when a subagent completes")
	stopInput := shared.SubagentStopHookInput{
		BaseHookInput: shared.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user/project",
		},
		HookEventName:       string(shared.HookEventSubagentStop),
		StopHookActive:      false,
		AgentID:             "agent-xyz789",
		AgentTranscriptPath: "/tmp/agent-xyz789-transcript.json",
	}
	printJSON("SubagentStop Input", stopInput)
	fmt.Println()
}

// demonstratePermissionHook shows permission request handling.
func demonstratePermissionHook() {
	fmt.Println("--- Permission Request Hook ---")
	fmt.Println("Called when a tool requests elevated permissions.")
	fmt.Println()

	input := shared.PermissionRequestHookInput{
		BaseHookInput: shared.BaseHookInput{
			SessionID:      "session-123",
			TranscriptPath: "/tmp/transcript.json",
			Cwd:            "/home/user/project",
		},
		HookEventName: string(shared.HookEventPermissionRequest),
		ToolName:      "Bash",
		ToolInput: map[string]any{
			"command": "sudo apt-get install nodejs",
		},
		PermissionSuggestions: []shared.PermissionUpdate{
			{
				Type:     "addRules",
				Behavior: shared.PermissionBehaviorAllow,
				Rules: []shared.PermissionRuleValue{
					{ToolName: "Bash"},
				},
			},
		},
	}

	printJSON("Input", input)

	fmt.Println("Example: Auto-approve safe commands, require confirmation for others")
	fmt.Print(`
  cmd := input.ToolInput["command"].(string)
  safeCommands := []string{"ls", "cat", "head", "tail", "grep"}

  for _, safe := range safeCommands {
      if strings.HasPrefix(cmd, safe) {
          return shared.SyncHookOutput{Decision: "approve"}
      }
  }

  // Require manual approval for other commands
  return shared.SyncHookOutput{
      Decision: "block",
      Reason:   "Command requires manual approval",
  }
`)
	fmt.Println()
}

// demonstrateHookOutputs shows hook output patterns.
func demonstrateHookOutputs() {
	fmt.Println("--- Hook Output Patterns ---")
	fmt.Println()

	fmt.Println("1. Continue (allow tool execution):")
	printJSON("Output", shared.SyncHookOutput{Continue: true})

	fmt.Println("2. Block (prevent tool execution):")
	printJSON("Output", shared.SyncHookOutput{
		Decision:   "block",
		StopReason: "File path not allowed",
	})

	fmt.Println("3. Approve with system message:")
	printJSON("Output", shared.SyncHookOutput{
		Decision:      "approve",
		SystemMessage: "Tool execution approved by security policy",
	})

	fmt.Println("4. Suppress output (hide from user):")
	printJSON("Output", shared.SyncHookOutput{
		Continue:       true,
		SuppressOutput: true,
	})

	fmt.Println("5. Async hook (for long-running validation):")
	printJSON("Output", shared.AsyncHookOutput{
		Async:        true,
		AsyncTimeout: 30, // seconds
	})

	fmt.Println("6. Custom hook-specific data:")
	printJSON("Output", shared.SyncHookOutput{
		Continue: true,
		HookSpecificOutput: map[string]any{
			"validated":     true,
			"validatedAt":   "2024-01-01T00:00:00Z",
			"validatedBy":   "security-policy-v2",
			"auditLogEntry": "audit-123",
		},
	})
	fmt.Println()
}

// printJSON prints a labeled JSON object.
func printJSON(label string, v any) {
	data, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		fmt.Printf("  %s: (error: %v)\n", label, err)
		return
	}
	fmt.Printf("  %s:\n  %s\n\n", label, string(data))
}
