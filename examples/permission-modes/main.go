// Package main demonstrates permission mode patterns with the Claude Agent SDK.
//
// This example shows:
// - All 6 permission modes and their behaviors
// - Permission behaviors (allow, deny, ask)
// - Permission update destinations
// - Permission rules configuration
// - Dynamic permission updates
//
// Based on TypeScript SDK permission system.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude"
)

func main() {
	fmt.Println("=== Permission Modes Example ===")
	fmt.Println("This demonstrates the permission system in the SDK.")
	fmt.Println()

	// Show all permission modes
	demonstratePermissionModes()

	// Show permission behaviors
	demonstratePermissionBehaviors()

	// Show permission update destinations
	demonstratePermissionDestinations()

	// Show permission rules
	demonstratePermissionRules()

	// Show permission results
	demonstratePermissionResults()

	// Show session configuration examples
	demonstrateSessionConfigurations()

	fmt.Println()
	fmt.Println("=== Permission Modes Example Complete ===")
}

// demonstratePermissionModes shows all available permission modes.
func demonstratePermissionModes() {
	fmt.Println("--- Permission Modes ---")
	fmt.Println()

	modes := []struct {
		mode        claude.PermissionMode
		description string
		useCase     string
	}{
		{
			mode:        claude.PermissionModeDefault,
			description: "Standard permission checking with prompts",
			useCase:     "Interactive sessions where user approves actions",
		},
		{
			mode:        claude.PermissionModeAcceptEdits,
			description: "Automatically accept file edits, prompt for others",
			useCase:     "Coding assistants where edits are expected",
		},
		{
			mode:        claude.PermissionModeBypassPermissions,
			description: "Skip all permission checks",
			useCase:     "Fully automated pipelines, CI/CD environments",
		},
		{
			mode:        claude.PermissionModePlan,
			description: "Plan mode - no execution, only planning",
			useCase:     "Reviewing proposed changes before execution",
		},
		{
			mode:        claude.PermissionModeDelegate,
			description: "Delegate permission decisions to hooks",
			useCase:     "Custom permission logic via PreToolUse hooks",
		},
		{
			mode:        claude.PermissionModeDontAsk,
			description: "Never prompt, deny if not pre-approved",
			useCase:     "Headless automation with pre-configured rules",
		},
	}

	for _, m := range modes {
		fmt.Printf("  %s:\n", m.mode)
		fmt.Printf("    Description: %s\n", m.description)
		fmt.Printf("    Use case: %s\n", m.useCase)
		fmt.Println()
	}
}

// demonstratePermissionBehaviors shows permission behavior options.
func demonstratePermissionBehaviors() {
	fmt.Println("--- Permission Behaviors ---")
	fmt.Println()

	behaviors := []struct {
		behavior    claude.PermissionBehavior
		description string
	}{
		{
			behavior:    claude.PermissionBehaviorAllow,
			description: "Allow the tool to execute without prompting",
		},
		{
			behavior:    claude.PermissionBehaviorDeny,
			description: "Deny the tool execution",
		},
		{
			behavior:    claude.PermissionBehaviorAsk,
			description: "Prompt the user for permission",
		},
	}

	for _, b := range behaviors {
		fmt.Printf("  %s: %s\n", b.behavior, b.description)
	}
	fmt.Println()
}

// demonstratePermissionDestinations shows where permissions are stored.
func demonstratePermissionDestinations() {
	fmt.Println("--- Permission Update Destinations ---")
	fmt.Println()

	destinations := []struct {
		dest        claude.PermissionUpdateDestination
		description string
		persistence string
	}{
		{
			dest:        claude.PermissionDestUserSettings,
			description: "User-level settings (~/.claude/settings.json)",
			persistence: "Persists across all sessions and projects",
		},
		{
			dest:        claude.PermissionDestProjectSettings,
			description: "Project-level settings (.claude/settings.json)",
			persistence: "Persists for this project only",
		},
		{
			dest:        claude.PermissionDestLocalSettings,
			description: "Local settings (not committed)",
			persistence: "Persists locally, not shared with team",
		},
		{
			dest:        claude.PermissionDestSession,
			description: "Session-only permissions",
			persistence: "Lost when session ends",
		},
		{
			dest:        claude.PermissionDestCLIArg,
			description: "Command-line argument",
			persistence: "Single execution only",
		},
	}

	for _, d := range destinations {
		fmt.Printf("  %s:\n", d.dest)
		fmt.Printf("    Description: %s\n", d.description)
		fmt.Printf("    Persistence: %s\n", d.persistence)
		fmt.Println()
	}
}

// demonstratePermissionRules shows permission rule configuration.
func demonstratePermissionRules() {
	fmt.Println("--- Permission Rules ---")
	fmt.Println()

	// Example: Add rules
	addRulesUpdate := claude.PermissionUpdate{
		Type:     "addRules",
		Behavior: claude.PermissionBehaviorAllow,
		Rules: []claude.PermissionRuleValue{
			{ToolName: "Read"},
			{ToolName: "Glob"},
			{ToolName: "Grep"},
		},
		Destination: claude.PermissionDestSession,
	}
	fmt.Println("1. Add rules (allow read-only tools):")
	printJSON("Update", addRulesUpdate)

	// Example: Add rules with content filter
	content := "*.go"
	addRulesWithContent := claude.PermissionUpdate{
		Type:     "addRules",
		Behavior: claude.PermissionBehaviorAllow,
		Rules: []claude.PermissionRuleValue{
			{ToolName: "Edit", RuleContent: &content},
		},
		Destination: claude.PermissionDestProjectSettings,
	}
	fmt.Println("2. Add rules with content filter (allow editing .go files):")
	printJSON("Update", addRulesWithContent)

	// Example: Replace all rules
	replaceRulesUpdate := claude.PermissionUpdate{
		Type:     "replaceRules",
		Behavior: claude.PermissionBehaviorDeny,
		Rules: []claude.PermissionRuleValue{
			{ToolName: "Bash"},
		},
		Destination: claude.PermissionDestSession,
	}
	fmt.Println("3. Replace rules (deny Bash commands):")
	printJSON("Update", replaceRulesUpdate)

	// Example: Set mode
	setModeUpdate := claude.PermissionUpdate{
		Type:        "setMode",
		Mode:        claude.PermissionModeAcceptEdits,
		Destination: claude.PermissionDestSession,
	}
	fmt.Println("4. Set permission mode:")
	printJSON("Update", setModeUpdate)

	// Example: Add allowed directories
	addDirsUpdate := claude.PermissionUpdate{
		Type:        "addDirectories",
		Directories: []string{"/home/user/project", "/tmp/workspace"},
		Destination: claude.PermissionDestSession,
	}
	fmt.Println("5. Add allowed directories:")
	printJSON("Update", addDirsUpdate)
}

// demonstratePermissionResults shows permission check results.
func demonstratePermissionResults() {
	fmt.Println("--- Permission Results ---")
	fmt.Println()

	// Allow result
	allowResult := claude.PermissionResult{
		Behavior:  claude.PermissionBehaviorAllow,
		ToolUseID: "tool-use-123",
	}
	fmt.Println("1. Allow result:")
	printJSON("Result", allowResult)

	// Allow with modified input
	allowModifiedResult := claude.PermissionResult{
		Behavior: claude.PermissionBehaviorAllow,
		UpdatedInput: map[string]any{
			"command": "ls -la", // Sanitized from "ls -la && rm -rf /"
		},
		ToolUseID: "tool-use-456",
	}
	fmt.Println("2. Allow with modified input (sanitized):")
	printJSON("Result", allowModifiedResult)

	// Deny result
	denyResult := claude.PermissionResult{
		Behavior:  claude.PermissionBehaviorDeny,
		Message:   "Operation not permitted: file outside allowed directories",
		ToolUseID: "tool-use-789",
		Interrupt: false,
	}
	fmt.Println("3. Deny result:")
	printJSON("Result", denyResult)

	// Deny with interrupt
	denyInterruptResult := claude.PermissionResult{
		Behavior:  claude.PermissionBehaviorDeny,
		Message:   "Critical security violation detected",
		ToolUseID: "tool-use-abc",
		Interrupt: true, // Stops the entire session
	}
	fmt.Println("4. Deny with interrupt (stops session):")
	printJSON("Result", denyInterruptResult)

	// Allow with permission updates
	allowWithUpdates := claude.PermissionResult{
		Behavior: claude.PermissionBehaviorAllow,
		UpdatedPermissions: []claude.PermissionUpdate{
			{
				Type:     "addRules",
				Behavior: claude.PermissionBehaviorAllow,
				Rules: []claude.PermissionRuleValue{
					{ToolName: "Bash"},
				},
				Destination: claude.PermissionDestSession,
			},
		},
		ToolUseID: "tool-use-def",
	}
	fmt.Println("5. Allow with permission updates (remember choice):")
	printJSON("Result", allowWithUpdates)
}

// demonstrateSessionConfigurations shows permission configuration for sessions.
func demonstrateSessionConfigurations() {
	fmt.Println("--- Session Configuration Examples ---")
	fmt.Println()

	fmt.Println("1. Read-only session (exploration mode):")
	fmt.Print(`
  session, err := v2.CreateSession(ctx,
      v2.WithPermissionMode(claude.PermissionModeDefault),
      v2.WithAllowedTools([]string{"Read", "Glob", "Grep", "LSP"}),
  )
`)
	fmt.Println()

	fmt.Println("2. Code editing session (accept edits):")
	fmt.Print(`
  session, err := v2.CreateSession(ctx,
      v2.WithPermissionMode(claude.PermissionModeAcceptEdits),
      v2.WithAllowedDirectories([]string{"/home/user/project"}),
  )
`)
	fmt.Println()

	fmt.Println("3. Fully automated session (CI/CD):")
	fmt.Print(`
  session, err := v2.CreateSession(ctx,
      v2.WithPermissionMode(claude.PermissionModeBypassPermissions),
      v2.WithMaxTurns(50),
      v2.WithTimeout(5*time.Minute),
  )
`)
	fmt.Println()

	fmt.Println("4. Plan-only session (review before execute):")
	fmt.Print(`
  session, err := v2.CreateSession(ctx,
      v2.WithPermissionMode(claude.PermissionModePlan),
  )
  // Claude will propose changes but not execute them
`)
	fmt.Println()

	fmt.Println("5. Custom permission logic (delegate to hooks):")
	fmt.Print(`
  session, err := v2.CreateSession(ctx,
      v2.WithPermissionMode(claude.PermissionModeDelegate),
      v2.WithHooks(claude.HooksConfig{
          PreToolUse: []claude.Hook{
              {
                  Matcher: ".*",
                  Handler: myPermissionHandler,
              },
          },
      }),
  )
`)
	fmt.Println()
}

// printJSON prints a labeled JSON object.
func printJSON(label string, v any) {
	data, err := json.MarshalIndent(v, "    ", "  ")
	if err != nil {
		fmt.Printf("    %s: (error: %v)\n", label, err)
		return
	}
	fmt.Printf("    %s:\n    %s\n\n", label, string(data))
}
