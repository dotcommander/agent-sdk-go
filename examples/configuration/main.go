// Package main demonstrates all configuration options for the Claude Agent SDK.
//
// This example shows:
// - Model selection and fallback configuration
// - Timeout and limit settings
// - Permission modes and tool restrictions
// - System prompts and context files
// - Environment variables and custom arguments
// - MCP server configuration
// - Sandbox and output format settings
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/configuration
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/claude/shared"
	"github.com/dotcommander/agent-sdk-go/claude/v2"
)

func main() {
	fmt.Println("=== Configuration Example ===")
	fmt.Println("This demonstrates all available configuration options.")
	fmt.Println()

	// Run all configuration demonstrations
	demonstrateBasicOptions()
	demonstrateLimitOptions()
	demonstrateToolOptions()
	demonstrateContextOptions()
	demonstrateAdvancedOptions()
	demonstrateMcpOptions()
	demonstrateEnvironmentOptions()

	// Interactive demo if CLI available
	if cli.IsCLIAvailable() && len(os.Args) > 1 && os.Args[1] == "demo" {
		runInteractiveDemo()
	}

	fmt.Println()
	fmt.Println("=== Configuration Example Complete ===")
}

// demonstrateBasicOptions shows model and timeout configuration.
func demonstrateBasicOptions() {
	fmt.Println("--- Basic Options ---")

	// Model selection examples
	fmt.Println("Model Options:")
	models := []string{
		"claude-sonnet-4-5-20250929", // Latest Sonnet
		"claude-3-5-haiku-20241022",  // Fast, cost-effective
		"claude-opus-4-20250514",     // Most capable
		"sonnet",                     // Short alias (resolved automatically)
		"haiku",                      // Short alias
		"opus",                       // Short alias
	}
	for _, model := range models {
		resolved := shared.ResolveModelName(model)
		if model != resolved {
			fmt.Printf("  %s -> %s\n", model, resolved)
		} else {
			fmt.Printf("  %s\n", model)
		}
	}
	fmt.Println()

	// Timeout configuration
	fmt.Println("Timeout Options:")
	timeouts := []time.Duration{
		30 * time.Second,  // Quick queries
		60 * time.Second,  // Standard operations
		120 * time.Second, // Complex tasks
		5 * time.Minute,   // Long-running jobs
	}
	for _, timeout := range timeouts {
		fmt.Printf("  %v - %s\n", timeout, describeTimeout(timeout))
	}
	fmt.Println()
}

// describeTimeout provides a description for a timeout value.
func describeTimeout(d time.Duration) string {
	switch {
	case d <= 30*time.Second:
		return "Quick queries, simple questions"
	case d <= 60*time.Second:
		return "Standard operations"
	case d <= 120*time.Second:
		return "Complex tasks, code generation"
	default:
		return "Long-running jobs, analysis"
	}
}

// demonstrateLimitOptions shows conversation and budget limits.
func demonstrateLimitOptions() {
	fmt.Println("--- Limit Options ---")

	// Max turns limit
	fmt.Println("MaxTurns - Limits conversation turns:")
	fmt.Println("  1   - Single response (no follow-up)")
	fmt.Println("  5   - Short conversations")
	fmt.Println("  10  - Extended discussions")
	fmt.Println("  nil - Unlimited (default)")
	fmt.Println()

	// Max thinking tokens
	fmt.Println("MaxThinkingTokens - Limits internal reasoning:")
	fmt.Println("  1000  - Quick decisions")
	fmt.Println("  4096  - Moderate thinking")
	fmt.Println("  16384 - Deep reasoning")
	fmt.Println("  nil   - Model default")
	fmt.Println()

	// Budget limits
	fmt.Println("MaxBudgetUSD - Cost control:")
	fmt.Println("  0.10 - Minimal testing")
	fmt.Println("  1.00 - Development")
	fmt.Println("  10.00 - Production batch")
	fmt.Println("  nil  - No limit (use with caution)")
	fmt.Println()

	// Example usage
	fmt.Println("Example: Creating session with limits")
	fmt.Println(`
  session, err := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-5-20250929"),
      v2.WithMaxTurns(5),
      v2.WithMaxThinkingTokens(4096),
      v2.WithMaxBudgetUSD(1.00),
  )`)
	fmt.Println()
}

// demonstrateToolOptions shows tool restriction configuration.
func demonstrateToolOptions() {
	fmt.Println("--- Tool Options ---")

	// Allowed tools
	fmt.Println("AllowedTools - Restrict to specific tools:")
	allowedExamples := [][]string{
		{"Read", "Write"},                 // File operations only
		{"Bash"},                          // Shell only
		{"Grep", "Glob"},                  // Search only
		{"Read", "Grep", "Glob"},          // Read-only exploration
		{"Read", "Write", "Edit", "Bash"}, // Full development
	}
	for _, tools := range allowedExamples {
		fmt.Printf("  %v - %s\n", tools, describeToolSet(tools))
	}
	fmt.Println()

	// Disallowed tools
	fmt.Println("DisallowedTools - Block specific tools:")
	disallowedExamples := [][]string{
		{"Bash"},                  // No shell access
		{"Write", "Edit"},         // Read-only mode
		{"WebFetch", "WebSearch"}, // No internet
	}
	for _, tools := range disallowedExamples {
		fmt.Printf("  %v - %s\n", tools, describeBlockedTools(tools))
	}
	fmt.Println()

	// Example usage
	fmt.Println("Example: Creating read-only session")
	fmt.Println(`
  session, err := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-5-20250929"),
      v2.WithAllowedTools("Read", "Grep", "Glob"),
      v2.WithDisallowedTools("Write", "Edit", "Bash"),
  )`)
	fmt.Println()
}

// describeToolSet provides a description for a set of allowed tools.
func describeToolSet(tools []string) string {
	switch len(tools) {
	case 2:
		if tools[0] == "Read" && tools[1] == "Write" {
			return "File operations only"
		}
		if tools[0] == "Grep" && tools[1] == "Glob" {
			return "Search operations only"
		}
	case 1:
		if tools[0] == "Bash" {
			return "Shell commands only"
		}
	case 3:
		return "Read-only exploration"
	case 4:
		return "Full development capabilities"
	}
	return "Custom tool set"
}

// describeBlockedTools provides a description for blocked tools.
func describeBlockedTools(tools []string) string {
	switch len(tools) {
	case 1:
		if tools[0] == "Bash" {
			return "Disable shell for safety"
		}
	case 2:
		if tools[0] == "Write" {
			return "Read-only mode"
		}
		if tools[0] == "WebFetch" {
			return "No internet access"
		}
	}
	return "Custom restrictions"
}

// demonstrateContextOptions shows context and prompt configuration.
func demonstrateContextOptions() {
	fmt.Println("--- Context Options ---")

	// System prompt
	fmt.Println("SystemPrompt - Custom instructions:")
	systemPrompts := []string{
		"You are a helpful coding assistant.",
		"Respond in JSON format only.",
		"You are a security expert reviewing code.",
		"Keep responses under 100 words.",
	}
	for _, prompt := range systemPrompts {
		fmt.Printf("  \"%s\"\n", prompt)
	}
	fmt.Println()

	// Context files
	fmt.Println("ContextFiles - Include files in context:")
	fmt.Println("  Single file:    []string{\"main.go\"}")
	fmt.Println("  Multiple files: []string{\"main.go\", \"go.mod\", \"README.md\"}")
	fmt.Println("  Config:         []string{\".env\", \"config.yaml\"}")
	fmt.Println()

	// Additional directories
	fmt.Println("AdditionalDirectories - Accessible paths:")
	fmt.Println("  []string{\"/tmp\"}")
	fmt.Println("  []string{\"/home/user/project\", \"/var/data\"}")
	fmt.Println()

	// Example usage
	fmt.Println("Example: Creating session with context")
	fmt.Println(`
  session, err := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-5-20250929"),
      v2.WithSystemPrompt("You are a Go expert. Follow Go idioms."),
      v2.WithContextFiles("main.go", "go.mod"),
      v2.WithAdditionalDirectories("/tmp"),
  )`)
	fmt.Println()
}

// demonstrateAdvancedOptions shows advanced configuration.
func demonstrateAdvancedOptions() {
	fmt.Println("--- Advanced Options ---")

	// Fallback model
	fmt.Println("FallbackModel - Automatic failover:")
	fmt.Println(`  v2.WithFallbackModel("claude-3-5-haiku-20241022")`)
	fmt.Println("  Falls back if primary model fails or is unavailable")
	fmt.Println()

	// Session persistence
	fmt.Println("PersistSession - Save sessions to disk:")
	fmt.Println("  v2.WithPersistSession(true)  - Save session (default)")
	fmt.Println("  v2.WithPersistSession(false) - Ephemeral session")
	fmt.Println()

	// File checkpointing
	fmt.Println("EnableFileCheckpointing - Track file changes:")
	fmt.Println("  v2.WithEnableFileCheckpointing(true)")
	fmt.Println("  Enables reverting files to previous states")
	fmt.Println()

	// Output format
	fmt.Println("OutputFormat - Structured responses:")
	fmt.Println(`
  format := &shared.OutputFormat{
      Type: "json_schema",
      Schema: map[string]any{
          "type": "object",
          "properties": map[string]any{
              "answer": map[string]any{"type": "string"},
              "confidence": map[string]any{"type": "number"},
          },
      },
  }
  v2.WithOutputFormat(format)`)
	fmt.Println()

	// Beta features
	fmt.Println("Betas - Enable experimental features:")
	fmt.Println(`  v2.WithBetas("context-1m-2025-08-07")`)
	fmt.Println()
}

// demonstrateMcpOptions shows MCP server configuration.
func demonstrateMcpOptions() {
	fmt.Println("--- MCP Server Options ---")

	// Stdio server
	fmt.Println("Stdio MCP Server:")
	stdioConfig := shared.McpStdioServerConfig{
		Type:    "stdio",
		Command: "npx",
		Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
	}
	fmt.Printf("  Type: %s\n", stdioConfig.Type)
	fmt.Printf("  Command: %s %v\n", stdioConfig.Command, stdioConfig.Args)
	fmt.Println()

	// SSE server
	fmt.Println("SSE MCP Server:")
	sseConfig := shared.McpSSEServerConfig{
		Type: "sse",
		URL:  "http://localhost:3000/sse",
	}
	fmt.Printf("  Type: %s\n", sseConfig.Type)
	fmt.Printf("  URL: %s\n", sseConfig.URL)
	fmt.Println()

	// HTTP server
	fmt.Println("HTTP MCP Server:")
	httpConfig := shared.McpHttpServerConfig{
		Type: "http",
		URL:  "http://localhost:8080/mcp",
	}
	fmt.Printf("  Type: %s\n", httpConfig.Type)
	fmt.Printf("  URL: %s\n", httpConfig.URL)
	fmt.Println()

	// Example usage
	fmt.Println("Example: Session with MCP servers")
	fmt.Println(`
  servers := map[string]shared.McpServerConfig{
      "filesystem": {
          Type:    "stdio",
          Command: "npx",
          Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
      },
      "custom": {
          Type: "http",
          URL:  "http://localhost:8080/mcp",
      },
  }
  session, err := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-5-20250929"),
      v2.WithMcpServers(servers),
  )`)
	fmt.Println()
}

// demonstrateEnvironmentOptions shows environment and subprocess configuration.
func demonstrateEnvironmentOptions() {
	fmt.Println("--- Environment Options ---")

	// Environment variables
	fmt.Println("Env - Set subprocess environment:")
	fmt.Println(`
  env := map[string]string{
      "CUSTOM_VAR": "value",
      "DEBUG":      "true",
  }
  v2.WithEnv(env)`)
	fmt.Println()

	// Custom arguments
	fmt.Println("CustomArgs - Additional CLI arguments:")
	fmt.Println(`  v2.WithCustomArgs("--verbose", "--no-color")`)
	fmt.Println()

	// Extra args (key-value)
	fmt.Println("ExtraArgs - Named CLI arguments:")
	fmt.Println(`
  args := map[string]string{
      "--config": "/path/to/config.yaml",
      "--log-level": "debug",
  }
  v2.WithExtraArgs(args)`)
	fmt.Println()
}

// runInteractiveDemo runs a live demonstration with the Claude CLI.
func runInteractiveDemo() {
	fmt.Println()
	fmt.Println("=== Interactive Demo ===")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create session with various options
	session, err := v2.CreateSession(ctx,
		// Model configuration
		v2.WithModel("claude-sonnet-4-5-20250929"),
		v2.WithFallbackModel("claude-3-5-haiku-20241022"),
		v2.WithTimeout(60*time.Second),

		// Limits
		v2.WithMaxTurns(3),
		v2.WithMaxThinkingTokens(4096),

		// Context
		v2.WithSystemPrompt("You are a helpful assistant. Keep responses concise."),

		// Session persistence
		v2.WithPersistSession(false), // Don't save this demo session
	)
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		return
	}
	defer func() { _ = session.Close() }()

	fmt.Printf("Session created: %s\n", session.SessionID())

	// Send a test query
	prompt := "What are the key benefits of Go's concurrency model? Answer briefly."
	fmt.Printf("Prompt: %s\n\n", prompt)

	if err := session.Send(ctx, prompt); err != nil {
		log.Printf("Send failed: %v", err)
		return
	}

	fmt.Println("Response:")
	fmt.Println("----------------------------------------")
	for msg := range session.Receive(ctx) {
		switch msg.Type() {
		case v2.V2EventTypeAssistant:
			text := v2.ExtractAssistantText(msg)
			if text != "" {
				fmt.Print(text)
			}
		case v2.V2EventTypeStreamDelta:
			text := v2.ExtractDeltaText(msg)
			if text != "" {
				fmt.Print(text)
			}
		case v2.V2EventTypeResult:
			text := v2.ExtractResultText(msg)
			if text != "" {
				fmt.Print(text)
			}
		case v2.V2EventTypeError:
			fmt.Printf("\nError: %s\n", v2.ExtractErrorMessage(msg))
		}
	}
	fmt.Println("\n----------------------------------------")
}
