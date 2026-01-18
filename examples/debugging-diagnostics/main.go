// Package main demonstrates debugging and diagnostics with the Claude Agent SDK.
//
// Debugging features allow you to:
// - Enable debug logging
// - Track stream statistics
// - Detect and diagnose issues
// - Monitor performance
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos
package main

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

func main() {
	fmt.Println("=== Debugging & Diagnostics Example ===")
	fmt.Println("This demonstrates debugging and diagnostic features.")
	fmt.Println()

	// Show debug logging
	demonstrateDebugLogging()

	// Show stream validation
	demonstrateStreamValidation()

	// Show stream statistics
	demonstrateStreamStats()

	// Show diagnostic patterns
	demonstrateDiagnosticPatterns()

	fmt.Println()
	fmt.Println("=== Debugging & Diagnostics Example Complete ===")
}

// demonstrateDebugLogging shows how to enable debug logging.
func demonstrateDebugLogging() {
	fmt.Println("--- Debug Logging ---")
	fmt.Println()

	fmt.Println("1. Enable debug output to stderr:")
	fmt.Println(`
  client, err := claude.NewClient(
      claude.WithModel("claude-sonnet-4-20250514"),
      claude.WithDebugWriter(os.Stderr),
  )`)

	fmt.Println("2. Log to file:")
	fmt.Println(`
  debugFile, _ := os.Create("debug.log")
  defer debugFile.Close()

  client, err := claude.NewClient(
      claude.WithDebugWriter(debugFile),
  )`)

	fmt.Println("3. Custom logger:")
	fmt.Println(`
  type DebugLogger struct {
      prefix string
  }

  func (l *DebugLogger) Write(p []byte) (n int, err error) {
      log.Printf("[%s] %s", l.prefix, string(p))
      return len(p), nil
  }

  client, err := claude.NewClient(
      claude.WithDebugWriter(&DebugLogger{prefix: "CLAUDE"}),
  )`)

	fmt.Println("4. Stderr callback for CLI output:")
	fmt.Println(`
  client, err := claude.NewClient(
      claude.WithStderrCallback(func(line string) {
          log.Printf("CLI stderr: %s", line)
      }),
  )`)
}

// demonstrateStreamValidation shows stream validation features.
func demonstrateStreamValidation() {
	fmt.Println("--- Stream Validation ---")
	fmt.Println()

	// Create a sample validator and show its stats
	validator := shared.NewStreamValidator()
	stats := validator.GetStats()

	fmt.Println("Stream validation detects issues like:")
	fmt.Println("  - Missing tool results")
	fmt.Println("  - Incomplete message sequences")
	fmt.Println("  - Unexpected message ordering")
	fmt.Println()

	printJSON("Initial Stats", stats)

	fmt.Println("Tracking messages with the validator:")
	fmt.Println(`
  validator := shared.NewStreamValidator()

  // Track each message from the stream
  iter := client.ReceiveResponseIterator(ctx)
  for {
      msg, err := iter.Next(ctx)
      if errors.Is(err, claude.ErrNoMoreMessages) {
          break
      }

      // Track the message for validation
      validator.TrackMessage(msg)

      // Process messages...
  }

  // Mark stream end and check for issues
  validator.MarkStreamEnd()
  issues := validator.GetIssues()

  if len(issues) > 0 {
      for _, issue := range issues {
          log.Printf("Stream issue: %s - %s", issue.Type, issue.Detail)
      }
  }`)
}

// demonstrateStreamStats shows stream statistics features.
func demonstrateStreamStats() {
	fmt.Println("--- Stream Statistics ---")
	fmt.Println()

	// Get stats from a validator
	validator := shared.NewStreamValidator()
	stats := validator.GetStats()

	printJSON("Stream Stats", stats)

	fmt.Println("Collecting statistics:")
	fmt.Println(`
  iter := client.ReceiveResponseIterator(ctx)
  for {
      msg, err := iter.Next(ctx)
      if errors.Is(err, claude.ErrNoMoreMessages) {
          break
      }
      // Process messages...
  }

  // Get stream statistics
  stats := client.GetStreamStats()

  // Format for display
  fmt.Println(claude.FormatStats(stats))`)
}

// demonstrateDiagnosticPatterns shows common diagnostic patterns.
func demonstrateDiagnosticPatterns() {
	fmt.Println("--- Diagnostic Patterns ---")
	fmt.Println()

	fmt.Println("1. Connection diagnostics:")
	fmt.Println(`
  info, err := client.GetServerInfo(ctx)
  if err != nil {
      log.Printf("Not connected: %v", err)
      return
  }

  fmt.Printf("Connected: %v\n", info["connected"])
  fmt.Printf("Transport: %s\n", info["transport_type"])
  fmt.Printf("Protocol Active: %v\n", info["protocol_active"])
  fmt.Printf("Initialized: %v\n", info["protocol_initialized"])`)

	fmt.Println("2. Error classification:")
	fmt.Println(`
  response, err := client.Query(ctx, prompt)
  if err != nil {
      // Classify the error type
      switch {
      case claude.IsCLINotFound(err):
          fmt.Println("Claude CLI not installed")
          fmt.Println("Install: npm install -g @anthropic-ai/claude")

      case claude.IsConnectionError(err):
          fmt.Println("Connection failed - check network/auth")
          if connErr, ok := claude.AsConnectionError(err); ok {
              fmt.Printf("Reason: %s\n", connErr.Reason)
          }

      case claude.IsTimeoutError(err):
          fmt.Println("Operation timed out")
          if timeoutErr, ok := claude.AsTimeoutError(err); ok {
              fmt.Printf("After: %s\n", timeoutErr.Timeout)
          }

      case claude.IsProtocolError(err):
          fmt.Println("Protocol error")
          if protoErr, ok := claude.AsProtocolError(err); ok {
              fmt.Printf("Message type: %s\n", protoErr.MessageType)
          }

      default:
          fmt.Printf("Unknown error: %v\n", err)
      }
  }`)

	fmt.Println("3. Performance monitoring:")
	fmt.Println(`
  startTime := time.Now()

  iter := client.ReceiveResponseIterator(ctx)
  var firstTokenTime time.Time
  tokenCount := 0

  for {
      msg, err := iter.Next(ctx)
      if errors.Is(err, claude.ErrNoMoreMessages) {
          break
      }

      if event, ok := msg.(*shared.StreamEvent); ok {
          if event.Event["type"] == shared.StreamEventTypeContentBlockDelta {
              if tokenCount == 0 {
                  firstTokenTime = time.Now()
                  fmt.Printf("Time to first token: %v\n", firstTokenTime.Sub(startTime))
              }
              tokenCount++
          }
      }
  }

  totalTime := time.Since(startTime)
  fmt.Printf("Total time: %v\n", totalTime)
  fmt.Printf("Tokens: %d\n", tokenCount)
  fmt.Printf("Tokens/sec: %.2f\n", float64(tokenCount)/totalTime.Seconds())`)

	fmt.Println("4. Comprehensive diagnostics function:")
	fmt.Println(`
  func runDiagnostics(client claude.Client, ctx context.Context) {
      fmt.Println("=== Claude SDK Diagnostics ===")

      // 1. Connection check
      info, err := client.GetServerInfo(ctx)
      if err != nil {
          fmt.Printf("X Connection: %v\n", err)
      } else {
          fmt.Printf("OK Connected: %v\n", info["connected"])
      }

      // 2. Protocol status
      if client.IsProtocolActive() {
          fmt.Println("OK Control protocol: Active")
      } else {
          fmt.Println("WARN Control protocol: Inactive")
      }

      // 3. Stream health
      issues := client.GetStreamIssues()
      if len(issues) == 0 {
          fmt.Println("OK Stream: Healthy")
      } else {
          fmt.Printf("WARN Stream issues: %d\n", len(issues))
          for _, issue := range issues {
              fmt.Printf("  - %s\n", issue.Detail)
          }
      }

      // 4. Statistics
      stats := client.GetStreamStats()
      fmt.Println(claude.FormatStats(stats))

      fmt.Println("==============================")
  }`)
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
