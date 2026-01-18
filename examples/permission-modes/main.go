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

	"github.com/dotcommander/agent-sdk-go/claude/shared"
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
		mode        shared.PermissionMode
		description string
		useCase     string
	}{
		{
			mode:        shared.PermissionModeDefault,
			description: "Standard permission checking with prompts",
			useCase:     "Interactive sessions where user approves actions",
		},
		{
			mode:        shared.PermissionModeAcceptEdits,
			description: "Automatically accept file edits, prompt for others",
			useCase:     "Coding assistants where edits are expected",
		},
		{
			mode:        shared.PermissionModeBypassPermissions,
			description: "Skip all permission checks",
			useCase:     "Fully automated pipelines, CI/CD environments",
		},
		{
			mode:        shared.PermissionModePlan,
			description: "Plan mode - no execution, only planning",
			useCase:     "Reviewing proposed changes before execution",
		},
		{
			mode:        shared.PermissionModeDelegate,
			description: "Delegate permission decisions to hooks",
			useCase:     "Custom permission logic via PreToolUse hooks",
		},
		{
			mode:        shared.PermissionModeDontAsk,
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
		behavior    shared.PermissionBehavior
		description string
	}{
		{
			behavior:    shared.PermissionBehaviorAllow,
			description: "Allow the tool to execute without prompting",
		},
		{
			behavior:    shared.PermissionBehaviorDeny,
			description: "Deny the tool execution",
		},
		{
			behavior:    shared.PermissionBehaviorAsk,
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
		dest        shared.PermissionUpdateDestination
		description string
		persistence string
	}{
		{
			dest:        shared.PermissionDestUserSettings,
			description: "User-level settings (~/.claude/settings.json)",
			persistence: "Persists across all sessions and projects",
		},
		{
			dest:        shared.PermissionDestProjectSettings,
			description: "Project-level settings (.claude/settings.json)",
			persistence: "Persists for this project only",
		},
		{
			dest:        shared.PermissionDestLocalSettings,
			description: "Local settings (not committed)",
			persistence: "Persists locally, not shared with team",
		},
		{
			dest:        shared.PermissionDestSession,
			description: "Session-only permissions",
			persistence: "Lost when session ends",
		},
		{
			dest:        shared.PermissionDestCLIArg,
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
	addRulesUpdate := shared.PermissionUpdate{
		Type:     "addRules",
		Behavior: shared.PermissionBehaviorAllow,
		Rules: []shared.PermissionRuleValue{
			{ToolName: "Read"},
			{ToolName: "Glob"},
			{ToolName: "Grep"},
		},
		Destination: shared.PermissionDestSession,
	}
	fmt.Println("1. Add rules (allow read-only tools):")
	printJSON("Update", addRulesUpdate)

	// Example: Add rules with content filter
	content := "*.go"
	addRulesWithContent := shared.PermissionUpdate{
		Type:     "addRules",
		Behavior: shared.PermissionBehaviorAllow,
		Rules: []shared.PermissionRuleValue{
			{ToolName: "Edit", RuleContent: &content},
		},
		Destination: shared.PermissionDestProjectSettings,
	}
	fmt.Println("2. Add rules with content filter (allow editing .go files):")
	printJSON("Update", addRulesWithContent)

	// Example: Replace all rules
	replaceRulesUpdate := shared.PermissionUpdate{
		Type:     "replaceRules",
		Behavior: shared.PermissionBehaviorDeny,
		Rules: []shared.PermissionRuleValue{
			{ToolName: "Bash"},
		},
		Destination: shared.PermissionDestSession,
	}
	fmt.Println("3. Replace rules (deny Bash commands):")
	printJSON("Update", replaceRulesUpdate)

	// Example: Set mode
	setModeUpdate := shared.PermissionUpdate{
		Type:        "setMode",
		Mode:        shared.PermissionModeAcceptEdits,
		Destination: shared.PermissionDestSession,
	}
	fmt.Println("4. Set permission mode:")
	printJSON("Update", setModeUpdate)

	// Example: Add allowed directories
	addDirsUpdate := shared.PermissionUpdate{
		Type:        "addDirectories",
		Directories: []string{"/home/user/project", "/tmp/workspace"},
		Destination: shared.PermissionDestSession,
	}
	fmt.Println("5. Add allowed directories:")
	printJSON("Update", addDirsUpdate)
}

// demonstratePermissionResults shows permission check results.
func demonstratePermissionResults() {
	fmt.Println("--- Permission Results ---")
	fmt.Println()

	// Allow result
	allowResult := shared.PermissionResult{
		Behavior:  shared.PermissionBehaviorAllow,
		ToolUseID: "tool-use-123",
	}
	fmt.Println("1. Allow result:")
	printJSON("Result", allowResult)

	// Allow with modified input
	allowModifiedResult := shared.PermissionResult{
		Behavior: shared.PermissionBehaviorAllow,
		UpdatedInput: map[string]any{
			"command": "ls -la", // Sanitized from "ls -la && rm -rf /"
		},
		ToolUseID: "tool-use-456",
	}
	fmt.Println("2. Allow with modified input (sanitized):")
	printJSON("Result", allowModifiedResult)

	// Deny result
	denyResult := shared.PermissionResult{
		Behavior:  shared.PermissionBehaviorDeny,
		Message:   "Operation not permitted: file outside allowed directories",
		ToolUseID: "tool-use-789",
		Interrupt: false,
	}
	fmt.Println("3. Deny result:")
	printJSON("Result", denyResult)

	// Deny with interrupt
	denyInterruptResult := shared.PermissionResult{
		Behavior:  shared.PermissionBehaviorDeny,
		Message:   "Critical security violation detected",
		ToolUseID: "tool-use-abc",
		Interrupt: true, // Stops the entire session
	}
	fmt.Println("4. Deny with interrupt (stops session):")
	printJSON("Result", denyInterruptResult)

	// Allow with permission updates
	allowWithUpdates := shared.PermissionResult{
		Behavior: shared.PermissionBehaviorAllow,
		UpdatedPermissions: []shared.PermissionUpdate{
			{
				Type:     "addRules",
				Behavior: shared.PermissionBehaviorAllow,
				Rules: []shared.PermissionRuleValue{
					{ToolName: "Bash"},
				},
				Destination: shared.PermissionDestSession,
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
      v2.WithPermissionMode(shared.PermissionModeDefault),
      v2.WithAllowedTools([]string{"Read", "Glob", "Grep", "LSP"}),
  )
`)
	fmt.Println()

	fmt.Println("2. Code editing session (accept edits):")
	fmt.Print(`
  session, err := v2.CreateSession(ctx,
      v2.WithPermissionMode(shared.PermissionModeAcceptEdits),
      v2.WithAllowedDirectories([]string{"/home/user/project"}),
  )
`)
	fmt.Println()

	fmt.Println("3. Fully automated session (CI/CD):")
	fmt.Print(`
  session, err := v2.CreateSession(ctx,
      v2.WithPermissionMode(shared.PermissionModeBypassPermissions),
      v2.WithMaxTurns(50),
      v2.WithTimeout(5*time.Minute),
  )
`)
	fmt.Println()

	fmt.Println("4. Plan-only session (review before execute):")
	fmt.Print(`
  session, err := v2.CreateSession(ctx,
      v2.WithPermissionMode(shared.PermissionModePlan),
  )
  // Claude will propose changes but not execute them
`)
	fmt.Println()

	fmt.Println("5. Custom permission logic (delegate to hooks):")
	fmt.Print(`
  session, err := v2.CreateSession(ctx,
      v2.WithPermissionMode(shared.PermissionModeDelegate),
      v2.WithHooks(shared.HooksConfig{
          PreToolUse: []shared.Hook{
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
