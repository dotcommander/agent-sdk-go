// Package main demonstrates error handling patterns with the Claude Agent SDK.
//
// This example shows:
// - Graceful failure modes and recovery
// - Error type discrimination (CLINotFound, Connection, Timeout, etc.)
// - Context cancellation handling
// - Error wrapping and unwrapping
// - Circuit breaker pattern for resilience
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/error-handling
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/claude/shared"
	"github.com/dotcommander/agent-sdk-go/claude/v2"
)

func main() {
	fmt.Println("=== Error Handling Example ===")
	fmt.Println("This demonstrates graceful error handling patterns in the SDK.")
	fmt.Println()

	// Run all error handling demonstrations
	demonstrateCLIAvailabilityCheck()
	demonstrateErrorTypeDiscrimination()
	demonstrateContextCancellation()
	demonstrateTimeoutHandling()
	demonstrateErrorWrapping()
	demonstrateCircuitBreaker()
	demonstrateGracefulDegradation()

	fmt.Println()
	fmt.Println("=== Error Handling Example Complete ===")
}

// demonstrateCLIAvailabilityCheck shows how to check CLI availability before operations.
func demonstrateCLIAvailabilityCheck() {
	fmt.Println("--- CLI Availability Check ---")

	// Check if CLI is available
	if !cli.IsCLIAvailable() {
		fmt.Println("Claude CLI not found.")
		fmt.Println("This is expected if Claude CLI is not installed.")
		fmt.Println("Suggestions:")
		fmt.Println("  1. Install Claude CLI: https://docs.anthropic.com/claude/docs/quickstart")
		fmt.Println("  2. Add Claude CLI to your PATH")
		fmt.Println()
		return
	}

	fmt.Println("Claude CLI is available")

	// Get CLI path for diagnostics
	result, err := cli.DiscoverCLI("", cli.GetDefaultCommand())
	if err == nil && result.Found {
		fmt.Printf("CLI path: %s\n", result.Path)
		fmt.Printf("CLI version: %s\n", result.Version)
	}
	fmt.Println()
}

// demonstrateErrorTypeDiscrimination shows how to handle different error types.
func demonstrateErrorTypeDiscrimination() {
	fmt.Println("--- Error Type Discrimination ---")

	// Create various error types for demonstration
	errors := []error{
		shared.NewCLINotFoundError("/usr/local/bin/claude", "claude"),
		shared.NewConnectionError("failed to establish connection", nil),
		shared.NewTimeoutError("query", "30s"),
		shared.NewParserError(42, 10, `{"invalid": json}`, "unexpected character"),
		shared.NewProtocolError("unknown_message", "received unexpected message type"),
		shared.NewConfigurationError("model", "invalid-model", "model name must start with 'claude-'"),
		shared.NewProcessError(12345, "claude", "process terminated unexpectedly", "SIGKILL"),
	}

	for _, err := range errors {
		handleError(err)
	}
	fmt.Println()
}

// handleError demonstrates type-specific error handling.
func handleError(err error) {
	fmt.Printf("  Error: %v\n", err)

	switch {
	case shared.IsCLINotFound(err):
		fmt.Println("    Type: CLINotFoundError")
		fmt.Println("    Recovery: Install Claude CLI or check PATH")

	case shared.IsConnectionError(err):
		fmt.Println("    Type: ConnectionError")
		fmt.Println("    Recovery: Retry with backoff or check network")

	case shared.IsTimeoutError(err):
		fmt.Println("    Type: TimeoutError")
		fmt.Println("    Recovery: Increase timeout or reduce query complexity")

	case shared.IsParserError(err):
		fmt.Println("    Type: ParserError")
		fmt.Println("    Recovery: Check CLI output format or report bug")

	case shared.IsProtocolError(err):
		fmt.Println("    Type: ProtocolError")
		fmt.Println("    Recovery: Check SDK version compatibility")

	default:
		fmt.Println("    Type: Unknown error")
		fmt.Println("    Recovery: Log and report")
	}
	fmt.Println()
}

// demonstrateContextCancellation shows proper context cancellation handling.
func demonstrateContextCancellation() {
	fmt.Println("--- Context Cancellation ---")

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Simulate cancellation after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	// Simulate a long-running operation
	err := simulateLongOperation(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Println("  Operation was cancelled (expected)")
			fmt.Println("  Recovery: Clean up resources, notify user")
		} else {
			fmt.Printf("  Unexpected error: %v\n", err)
		}
	}
	fmt.Println()
}

// simulateLongOperation simulates an operation that respects context cancellation.
func simulateLongOperation(ctx context.Context) error {
	select {
	case <-time.After(5 * time.Second):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// demonstrateTimeoutHandling shows timeout configuration and handling.
func demonstrateTimeoutHandling() {
	fmt.Println("--- Timeout Handling ---")

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Simulate operation that exceeds timeout
	err := simulateLongOperation(ctx)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("  Operation timed out (expected)")
			fmt.Println("  Recovery options:")
			fmt.Println("    1. Increase timeout duration")
			fmt.Println("    2. Reduce query complexity")
			fmt.Println("    3. Use streaming for long responses")
		} else {
			fmt.Printf("  Unexpected error: %v\n", err)
		}
	}
	fmt.Println()
}

// demonstrateErrorWrapping shows proper error wrapping for debugging.
func demonstrateErrorWrapping() {
	fmt.Println("--- Error Wrapping ---")

	// Simulate layered error wrapping
	baseErr := shared.NewConnectionError("network unreachable", nil)
	wrappedErr := fmt.Errorf("session creation failed: %w", baseErr)
	topErr := fmt.Errorf("query failed: %w", wrappedErr)

	fmt.Printf("  Top-level error: %v\n", topErr)
	fmt.Println()

	// Demonstrate error unwrapping
	fmt.Println("  Unwrapping chain:")
	var connErr *shared.ConnectionError
	if errors.As(topErr, &connErr) {
		fmt.Println("    Found ConnectionError in chain")
		fmt.Printf("    Original error: %v\n", connErr)
	}

	// Check error type
	if shared.IsConnectionError(errors.Unwrap(errors.Unwrap(topErr))) {
		fmt.Println("    Verified: root cause is ConnectionError")
	}
	fmt.Println()
}

// demonstrateCircuitBreaker shows the circuit breaker pattern for resilience.
func demonstrateCircuitBreaker() {
	fmt.Println("--- Circuit Breaker Pattern ---")

	// Create circuit breaker with custom config
	cb := shared.NewStubCircuitBreaker(shared.CircuitBreakerConfig{
		FailureThreshold:    3,
		RecoveryTimeout:     100 * time.Millisecond,
		HalfOpenMaxRequests: 1,
	})

	ctx := context.Background()

	// Simulate failures to trip the circuit
	fmt.Println("  Simulating failures to trip circuit:")
	for i := 0; i < 5; i++ {
		err := cb.Execute(ctx, func() error {
			return errors.New("simulated failure")
		})
		fmt.Printf("    Attempt %d: State=%s, Error=%v\n", i+1, cb.State(), err != nil)
	}
	fmt.Println()

	// Wait for recovery timeout
	fmt.Println("  Waiting for recovery timeout...")
	time.Sleep(150 * time.Millisecond)

	// Circuit should be half-open now
	fmt.Printf("  State after recovery: %s\n", cb.State())

	// Successful request should close the circuit
	err := cb.Execute(ctx, func() error {
		return nil // Success
	})
	fmt.Printf("  After success: State=%s, Error=%v\n", cb.State(), err)
	fmt.Println()
}

// demonstrateGracefulDegradation shows how to handle errors gracefully in production.
func demonstrateGracefulDegradation() {
	fmt.Println("--- Graceful Degradation ---")

	// Skip if CLI is not available
	if !cli.IsCLIAvailable() {
		fmt.Println("  Skipping (Claude CLI not available)")
		fmt.Println("  In production, you would provide fallback behavior here")
		fmt.Println()
		return
	}

	// Create context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create session with error handling
	session, err := createSessionWithFallback(ctx)
	if err != nil {
		fmt.Printf("  Session creation failed: %v\n", err)
		fmt.Println("  Graceful degradation: providing cached/default response")
		fmt.Println()
		return
	}
	defer session.Close()

	fmt.Printf("  Session created: %s\n", session.SessionID())

	// Send query with error handling
	response, err := queryWithRetry(ctx, session, "Hello!", 3)
	if err != nil {
		fmt.Printf("  Query failed after retries: %v\n", err)
		fmt.Println("  Graceful degradation: notify user of temporary unavailability")
	} else {
		fmt.Printf("  Response: %s\n", truncate(response, 100))
	}
	fmt.Println()
}

// createSessionWithFallback attempts to create a session with fallback options.
func createSessionWithFallback(ctx context.Context) (v2.V2Session, error) {
	// Try primary model first
	session, err := v2.CreateSession(ctx,
		v2.WithModel("claude-sonnet-4-5-20250929"),
		v2.WithTimeout(30*time.Second),
	)
	if err == nil {
		return session, nil
	}

	log.Printf("Primary model failed: %v, trying fallback", err)

	// Try fallback model
	session, err = v2.CreateSession(ctx,
		v2.WithModel("claude-3-5-haiku-20241022"),
		v2.WithTimeout(30*time.Second),
	)
	if err == nil {
		return session, nil
	}

	return nil, fmt.Errorf("all models failed: %w", err)
}

// queryWithRetry attempts a query with exponential backoff retries.
func queryWithRetry(ctx context.Context, session v2.V2Session, prompt string, maxRetries int) (string, error) {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
			log.Printf("Retry %d after %v", attempt, backoff)

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		response, err := executeQuery(ctx, session, prompt)
		if err == nil {
			return response, nil
		}

		lastErr = err

		// Don't retry on certain errors
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}

		log.Printf("Attempt %d failed: %v", attempt+1, err)
	}

	return "", fmt.Errorf("query failed after %d attempts: %w", maxRetries, lastErr)
}

// executeQuery sends a prompt and collects the response.
func executeQuery(ctx context.Context, session v2.V2Session, prompt string) (string, error) {
	if err := session.Send(ctx, prompt); err != nil {
		return "", fmt.Errorf("send failed: %w", err)
	}

	var result string
	for msg := range session.Receive(ctx) {
		switch msg.Type() {
		case v2.V2EventTypeAssistant:
			result += v2.ExtractAssistantText(msg)
		case v2.V2EventTypeStreamDelta:
			result += v2.ExtractDeltaText(msg)
		case v2.V2EventTypeResult:
			if result == "" {
				result = v2.ExtractResultText(msg)
			}
		case v2.V2EventTypeError:
			return "", errors.New(v2.ExtractErrorMessage(msg))
		}
	}

	if result == "" {
		return "", errors.New("empty response")
	}

	return result, nil
}

// truncate truncates a string to maxLen characters.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Example of production-ready error handler
func productionErrorHandler(err error) {
	// Ignore if running in test
	if os.Getenv("GO_TEST") != "" {
		return
	}

	switch {
	case shared.IsCLINotFound(err):
		// Log and alert - infrastructure issue
		log.Printf("CRITICAL: Claude CLI not found: %v", err)

	case shared.IsConnectionError(err):
		// Log and retry with backoff
		log.Printf("WARN: Connection error, will retry: %v", err)

	case shared.IsTimeoutError(err):
		// Log and potentially increase timeout
		log.Printf("WARN: Operation timed out: %v", err)

	case shared.IsParserError(err):
		// Log with full context for debugging
		log.Printf("ERROR: Parser error, may indicate CLI version mismatch: %v", err)

	case errors.Is(err, context.Canceled):
		// User-initiated cancellation, clean exit
		log.Printf("INFO: Operation cancelled by user")

	case errors.Is(err, context.DeadlineExceeded):
		// Timeout, may need to adjust configuration
		log.Printf("WARN: Operation deadline exceeded")

	default:
		// Unknown error, log and report
		log.Printf("ERROR: Unexpected error: %v", err)
	}
}
