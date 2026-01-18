// Package main demonstrates session control patterns with the Claude Agent SDK.
//
// This example shows:
// - Session resume and persistence
// - Session limits (max turns, budget, thinking tokens)
// - Session interruption and graceful shutdown
// - Session forking
// - Session state inspection
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world-v2
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/claude/v2"
)

func main() {
	fmt.Println("=== Session Control Example ===")
	fmt.Println("This demonstrates session lifecycle and control patterns.")
	fmt.Println()

	// Check CLI availability
	if !cli.IsCLIAvailable() {
		fmt.Println("Claude CLI not available. Showing patterns only.")
		fmt.Println()
		demonstratePatterns()
		return
	}

	// Run demonstrations
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Basic session lifecycle
	demonstrateBasicLifecycle(ctx)

	// Session resume
	demonstrateSessionResume(ctx)

	// Session limits
	demonstrateSessionLimits()

	// Graceful shutdown
	demonstrateGracefulShutdown()

	fmt.Println()
	fmt.Println("=== Session Control Example Complete ===")
}

// demonstratePatterns shows session control patterns without actual execution.
func demonstratePatterns() {
	fmt.Println("--- Session Control Patterns ---")
	fmt.Println()

	fmt.Println("1. Basic Session Lifecycle:")
	fmt.Print(`
  // Create session
  session, err := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-20250514"),
      v2.WithTimeout(60*time.Second),
  )
  if err != nil {
      log.Fatal(err)
  }
  defer session.Close()

  // Get session ID for later resume
  sessionID := session.SessionID()

  // Send message and receive response
  session.Send(ctx, "Hello!")
  for msg := range session.Receive(ctx) {
      // Process messages
  }
`)
	fmt.Println()

	fmt.Println("2. Session Resume:")
	fmt.Print(`
  // Resume an existing session
  session, err := v2.ResumeSession(ctx, savedSessionID,
      v2.WithModel("claude-sonnet-4-20250514"),
  )
  if err != nil {
      log.Fatal(err)
  }
  defer session.Close()

  // Check if session was resumed
  if session.IsResumed() {
      fmt.Println("Resumed existing session")
  }

  // Continue conversation
  session.Send(ctx, "What were we discussing?")
`)
	fmt.Println()

	fmt.Println("3. Session Limits:")
	fmt.Print(`
  // Create session with limits
  session, err := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-20250514"),
      v2.WithMaxTurns(10),              // Maximum conversation turns
      v2.WithMaxBudget(1.00),           // Maximum cost in USD
      v2.WithMaxThinkingTokens(4096),   // Limit thinking tokens
      v2.WithTimeout(5*time.Minute),    // Overall timeout
  )
`)
	fmt.Println()

	fmt.Println("4. Graceful Shutdown:")
	fmt.Print(`
  // Handle interrupts gracefully
  ctx, cancel := signal.NotifyContext(context.Background(),
      syscall.SIGINT, syscall.SIGTERM)
  defer cancel()

  session, _ := v2.CreateSession(ctx, ...)
  defer session.Close()

  // Send message
  session.Send(ctx, prompt)

  // Receive with cancellation support
  for msg := range session.Receive(ctx) {
      select {
      case <-ctx.Done():
          fmt.Println("Interrupted, cleaning up...")
          return
      default:
          // Process message
      }
  }
`)
	fmt.Println()

	fmt.Println("5. Session State Inspection:")
	fmt.Println("  // Get session information")
	sessionIDCode := `  fmt.Printf("Session ID: %s\n", session.SessionID())`
	isResumedCode := `  fmt.Printf("Is Resumed: %v\n", session.IsResumed())`
	stringCode := `  fmt.Printf("String: %s\n", session.String())`
	fmt.Println(sessionIDCode)
	fmt.Println(isResumedCode)
	fmt.Println(stringCode)
	fmt.Println()
	fmt.Println("  // Get underlying client for advanced operations")
	fmt.Println("  client := session.GetClient()")
	fmt.Println()
}

// demonstrateBasicLifecycle shows basic session creation and usage.
func demonstrateBasicLifecycle(ctx context.Context) {
	fmt.Println("--- Basic Session Lifecycle ---")

	// Create session
	session, err := v2.CreateSession(ctx,
		v2.WithModel("claude-sonnet-4-20250514"),
		v2.WithTimeout(30*time.Second),
	)
	if err != nil {
		fmt.Printf("  Failed to create session: %v\n", err)
		fmt.Println("  (This is expected if Claude CLI is not fully configured)")
		fmt.Println()
		return
	}
	defer func() { _ = session.Close() }()

	fmt.Printf("  Created session: %s\n", session.SessionID())

	// Simple query
	if err := session.Send(ctx, "Say 'Hello' and nothing else."); err != nil {
		fmt.Printf("  Send error: %v\n", err)
		fmt.Println()
		return
	}

	fmt.Println("  Response:")
	for msg := range session.Receive(ctx) {
		if msg.Type() == v2.V2EventTypeAssistant {
			text := v2.ExtractAssistantText(msg)
			if text != "" {
				fmt.Printf("    %s\n", text)
			}
		}
	}
	fmt.Println()
}

// demonstrateSessionResume shows session persistence and resume.
func demonstrateSessionResume(ctx context.Context) {
	fmt.Println("--- Session Resume ---")

	// Create initial session
	session1, err := v2.CreateSession(ctx,
		v2.WithModel("claude-sonnet-4-20250514"),
		v2.WithTimeout(30*time.Second),
	)
	if err != nil {
		fmt.Printf("  Failed to create session: %v\n", err)
		fmt.Println()
		return
	}

	// Save session ID
	sessionID := session1.SessionID()
	fmt.Printf("  Created session 1: %s\n", sessionID)

	// Establish context
	if err := session1.Send(ctx, "Remember: the secret word is 'banana'. Just confirm you remember."); err != nil {
		fmt.Printf("  Send error: %v\n", err)
		_ = session1.Close()
		fmt.Println()
		return
	}

	// Get response
	for msg := range session1.Receive(ctx) {
		if msg.Type() == v2.V2EventTypeAssistant {
			text := v2.ExtractAssistantText(msg)
			if text != "" {
				fmt.Printf("  Session 1 response: %s\n", truncate(text, 80))
			}
		}
	}

	// Close first session
	_ = session1.Close()
	fmt.Println("  Session 1 closed")

	// Resume session
	fmt.Println("  Resuming session...")
	session2, err := v2.ResumeSession(ctx, sessionID,
		v2.WithModel("claude-sonnet-4-20250514"),
		v2.WithTimeout(30*time.Second),
	)
	if err != nil {
		fmt.Printf("  Failed to resume session: %v\n", err)
		fmt.Println()
		return
	}
	defer func() { _ = session2.Close() }()

	fmt.Printf("  Resumed session 2: %s\n", session2.SessionID())

	// Verify context was preserved
	if err := session2.Send(ctx, "What is the secret word?"); err != nil {
		fmt.Printf("  Send error: %v\n", err)
		fmt.Println()
		return
	}

	for msg := range session2.Receive(ctx) {
		if msg.Type() == v2.V2EventTypeAssistant {
			text := v2.ExtractAssistantText(msg)
			if text != "" {
				fmt.Printf("  Session 2 response: %s\n", truncate(text, 80))
			}
		}
	}
	fmt.Println()
}

// demonstrateSessionLimits shows session limit configurations.
func demonstrateSessionLimits() {
	fmt.Println("--- Session Limits ---")
	fmt.Println()

	// Show available limit options
	limits := []struct {
		option      string
		description string
		example     string
	}{
		{
			option:      "WithMaxTurns(n)",
			description: "Maximum number of conversation turns",
			example:     "v2.WithMaxTurns(10) // Limit to 10 exchanges",
		},
		{
			option:      "WithMaxBudget(usd)",
			description: "Maximum cost in USD",
			example:     "v2.WithMaxBudget(1.00) // Limit to $1.00",
		},
		{
			option:      "WithMaxThinkingTokens(n)",
			description: "Maximum thinking/reasoning tokens",
			example:     "v2.WithMaxThinkingTokens(4096) // Limit thinking",
		},
		{
			option:      "WithTimeout(duration)",
			description: "Overall session timeout",
			example:     "v2.WithTimeout(5*time.Minute) // 5 minute timeout",
		},
	}

	for _, l := range limits {
		fmt.Printf("  %s:\n", l.option)
		fmt.Printf("    %s\n", l.description)
		fmt.Printf("    Example: %s\n", l.example)
		fmt.Println()
	}
}

// demonstrateGracefulShutdown shows interrupt handling patterns.
func demonstrateGracefulShutdown() {
	fmt.Println("--- Graceful Shutdown ---")
	fmt.Println()

	fmt.Println("  Best practices for graceful shutdown:")
	fmt.Println()
	fmt.Println("  1. Use signal.NotifyContext for interrupt handling:")
	fmt.Println("     ctx, cancel := signal.NotifyContext(context.Background(),")
	fmt.Println("         syscall.SIGINT, syscall.SIGTERM)")
	fmt.Println("     defer cancel()")
	fmt.Println()
	fmt.Println("  2. Always defer session.Close():")
	fmt.Println("     session, _ := v2.CreateSession(ctx, ...)")
	fmt.Println("     defer session.Close()")
	fmt.Println()
	fmt.Println("  3. Check context in receive loops:")
	fmt.Println("     for msg := range session.Receive(ctx) {")
	fmt.Println("         select {")
	fmt.Println("         case <-ctx.Done():")
	fmt.Println("             return // Clean exit on interrupt")
	fmt.Println("         default:")
	fmt.Println("             // Process message")
	fmt.Println("         }")
	fmt.Println("     }")
	fmt.Println()
	fmt.Println("  4. Use timeouts for long operations:")
	fmt.Println("     ctx, cancel := context.WithTimeout(ctx, 30*time.Second)")
	fmt.Println("     defer cancel()")
	fmt.Println()
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Simulated functions that would be used with actual CLI

func saveSessionID(id string) error {
	return os.WriteFile("/tmp/claude-session-id", []byte(id), 0600)
}

func loadSessionID() (string, error) {
	data, err := os.ReadFile("/tmp/claude-session-id")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
