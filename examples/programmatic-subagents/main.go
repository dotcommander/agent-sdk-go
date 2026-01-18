// Package main demonstrates programmatic subagents with the Claude Agent SDK.
//
// Programmatic subagents allow you to:
// - Define custom agent types for specialized tasks
// - Configure agent models and tool permissions
// - Create agent hierarchies for complex workflows
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos
package main

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

func main() {
	fmt.Println("=== Programmatic Subagents Example ===")
	fmt.Println("This demonstrates creating and configuring custom subagents.")
	fmt.Println()

	// Show basic agent definition
	demonstrateBasicAgent()

	// Show agent model configuration
	demonstrateAgentModels()

	// Show agent tool restrictions
	demonstrateToolRestrictions()

	// Show agent hierarchies
	demonstrateAgentHierarchies()

	// Show practical patterns
	demonstratePracticalPatterns()

	fmt.Println()
	fmt.Println("=== Programmatic Subagents Example Complete ===")
}

// demonstrateBasicAgent shows basic agent definition.
func demonstrateBasicAgent() {
	fmt.Println("--- Basic Agent Definition ---")
	fmt.Println()

	// Note: AgentDefinition uses Description, Prompt, and Model (not Name)
	// The agent is registered by key in the Agents map
	agent := shared.AgentDefinition{
		Description: "Reviews code for quality, security, and best practices",
		Prompt:      "You are a senior code reviewer. Focus on security, performance, and maintainability.",
		Model:       shared.AgentModelSonnet,
	}

	printJSON("Agent Definition", agent)

	fmt.Println(`
  // Register with client using a map (key becomes the agent name)
  client, err := claude.NewClient(
      claude.WithAgents(map[string]shared.AgentDefinition{
          "code-reviewer": agent,
      }),
  )

  // Claude will use this agent when Task tool specifies "code-reviewer"
`)
}

// demonstrateAgentModels shows model configuration for agents.
func demonstrateAgentModels() {
	fmt.Println("--- Agent Model Configuration ---")
	fmt.Println()

	fmt.Println("1. Using model aliases (recommended):")
	agents := map[string]shared.AgentDefinition{
		"research-agent": {
			Description: "Performs deep research and analysis",
			Prompt:      "You are a research specialist. Be thorough and cite sources.",
			Model:       shared.AgentModelOpus, // Use Opus for complex reasoning
		},
		"quick-lookup": {
			Description: "Fast lookups and simple queries",
			Prompt:      "Provide quick, concise answers.",
			Model:       shared.AgentModelHaiku, // Use Haiku for speed
		},
		"balanced-agent": {
			Description: "General purpose agent",
			Prompt:      "Help with various tasks effectively.",
			Model:       shared.AgentModelSonnet, // Default balanced choice
		},
	}

	for name, agent := range agents {
		fmt.Printf("  %s: Model=%s\n", name, agent.Model)
	}
	fmt.Println()

	fmt.Println("Available model aliases:")
	fmt.Println("  - AgentModelSonnet  -> claude-sonnet-4")
	fmt.Println("  - AgentModelOpus    -> claude-opus-4")
	fmt.Println("  - AgentModelHaiku   -> claude-haiku-3.5")
	fmt.Println("  - AgentModelInherit -> inherits from parent")
	fmt.Println()
}

// demonstrateToolRestrictions shows agent tool permissions.
func demonstrateToolRestrictions() {
	fmt.Println("--- Agent Tool Restrictions ---")
	fmt.Println()

	fmt.Println("1. Read-only agent:")
	readOnlyAgent := shared.AgentDefinition{
		Description:     "Can only read files, not modify",
		Prompt:          "Analyze code without making changes.",
		Tools:           []string{"Read", "Glob", "Grep", "LSP"},
		Model:           shared.AgentModelSonnet,
	}
	printJSON("Read-Only Agent", readOnlyAgent)

	fmt.Println("2. No-bash agent:")
	noBashAgent := shared.AgentDefinition{
		Description:     "Cannot execute shell commands",
		Prompt:          "Work without shell access for safety.",
		DisallowedTools: []string{"Bash", "KillShell"},
		Model:           shared.AgentModelSonnet,
	}
	printJSON("No-Bash Agent", noBashAgent)

	fmt.Println("3. Minimal tools agent:")
	minimalAgent := shared.AgentDefinition{
		Description: "Only uses specific tools",
		Prompt:      "Focus on file editing tasks only.",
		Tools:       []string{"Read", "Edit", "Write"},
		Model:       shared.AgentModelSonnet,
	}
	printJSON("Minimal Tools Agent", minimalAgent)

	fmt.Println("Note: Tools takes precedence - only listed tools are available.")
	fmt.Println("      DisallowedTools removes specific tools from the default set.")
	fmt.Println()
}

// demonstrateAgentHierarchies shows creating agent hierarchies.
func demonstrateAgentHierarchies() {
	fmt.Println("--- Agent Hierarchies ---")
	fmt.Println()

	agents := map[string]shared.AgentDefinition{
		"orchestrator": {
			Description: "Coordinates complex tasks by delegating to specialist agents",
			Prompt:      "You are a project manager. Delegate tasks to appropriate specialists.",
			Model:       shared.AgentModelOpus,
		},
		"frontend-dev": {
			Description: "Specializes in React, TypeScript, and CSS",
			Prompt:      "You are a frontend specialist. Focus on UI/UX and client-side code.",
			Model:       shared.AgentModelSonnet,
			Tools:       []string{"Read", "Write", "Edit", "Glob", "Grep", "Bash"},
		},
		"backend-dev": {
			Description: "Specializes in Go, databases, and APIs",
			Prompt:      "You are a backend specialist. Focus on server-side code and APIs.",
			Model:       shared.AgentModelSonnet,
			Tools:       []string{"Read", "Write", "Edit", "Glob", "Grep", "Bash", "LSP"},
		},
		"tester": {
			Description: "Writes and runs tests",
			Prompt:      "You are a QA specialist. Write thorough tests and verify behavior.",
			Model:       shared.AgentModelSonnet,
			Tools:       []string{"Read", "Write", "Edit", "Glob", "Grep", "Bash"},
		},
	}

	fmt.Println("Team of Specialized Agents:")
	for name, agent := range agents {
		fmt.Printf("  - %s (%s)\n", name, agent.Model)
		fmt.Printf("    %s\n", agent.Description)
	}
	fmt.Println()

	fmt.Println(`
  // Register all agents
  client, err := claude.NewClient(
      claude.WithAgentMap(agents),
  )

  // The orchestrator can delegate to specialists:
  // "Use the frontend-dev agent to fix the React component,
  //  then use the tester agent to verify the fix"
`)
}

// demonstratePracticalPatterns shows practical agent patterns.
func demonstratePracticalPatterns() {
	fmt.Println("--- Practical Agent Patterns ---")
	fmt.Println()

	fmt.Println("1. Security Review Pipeline:")
	fmt.Println(`
  agents := map[string]shared.AgentDefinition{
      "security-scanner": {
          Description: "Scans code for security vulnerabilities",
          Prompt:      "Focus on OWASP top 10 and common security issues.",
          Model:       shared.AgentModelSonnet,
          Tools:       []string{"Read", "Glob", "Grep"}, // Read-only
      },
      "security-fixer": {
          Description: "Fixes identified security issues",
          Prompt:      "Apply secure coding practices when fixing issues.",
          Model:       shared.AgentModelOpus,
          Tools:       []string{"Read", "Write", "Edit", "Glob", "Grep"},
      },
  }
`)

	fmt.Println("2. Documentation Generator:")
	fmt.Println(`
  agents := map[string]shared.AgentDefinition{
      "code-analyzer": {
          Description: "Analyzes code structure and extracts documentation",
          Prompt:      "Extract function signatures, types, and patterns.",
          Model:       shared.AgentModelHaiku, // Fast analysis
          Tools:       []string{"Read", "Glob", "Grep", "LSP"},
      },
      "doc-writer": {
          Description: "Writes comprehensive documentation",
          Prompt:      "Write clear, comprehensive docs with examples.",
          Model:       shared.AgentModelSonnet,
          Tools:       []string{"Read", "Write", "Edit"},
      },
  }
`)

	fmt.Println("3. Test-Driven Development:")
	fmt.Println(`
  agents := map[string]shared.AgentDefinition{
      "test-writer": {
          Description: "Writes tests for specified functionality",
          Prompt:      "Write failing tests first (red phase).",
          Model:       shared.AgentModelSonnet,
      },
      "implementer": {
          Description: "Implements code to pass the tests",
          Prompt:      "Write minimal code to pass tests (green phase).",
          Model:       shared.AgentModelSonnet,
      },
      "refactorer": {
          Description: "Refactors passing code for quality",
          Prompt:      "Improve code quality while keeping tests green (refactor phase).",
          Model:       shared.AgentModelOpus, // Complex reasoning
      },
  }

  // Workflow: test-writer -> implementer -> refactorer (red-green-refactor)
`)
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
