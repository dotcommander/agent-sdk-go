// Package main demonstrates streaming responses from Claude using the Agent SDK.
//
// This example shows:
// - Creating a V2 session for streaming
// - Processing messages via channels
// - Handling different message types (assistant, result, delta, error)
// - Real-time output as Claude generates the response
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world-v2
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/claude/v2"
)

func main() {
	// Check CLI availability first
	if !cli.IsCLIAvailable() {
		log.Fatal("Claude CLI not found. Please install it first: https://docs.anthropic.com/claude/docs/quickstart")
	}

	// Create a context that cancels on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Get the prompt from command line or use default
	prompt := "Tell me a short story about a robot learning to paint. Keep it under 200 words."
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	}

	fmt.Println("=== Streaming Example ===")
	fmt.Printf("Prompt: %s\n\n", prompt)
	fmt.Println("Response (streaming):")
	fmt.Println("----------------------------------------")

	// Create a session with streaming enabled
	session, err := v2.CreateSession(ctx,
		v2.WithModel("claude-sonnet-4-20250514"),
		v2.WithTimeout(120*time.Second),
		v2.WithEnablePartialMessages(true),
	)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer func() { _ = session.Close() }()

	// Send the prompt
	if err := session.Send(ctx, prompt); err != nil {
		log.Fatalf("Failed to send prompt: %v", err)
	}

	// Track metrics
	startTime := time.Now()
	var totalChars int

	// Stream the response using channels
	for msg := range session.Receive(ctx) {
		switch msg.Type() {
		case v2.V2EventTypeAssistant:
			// Full assistant message (may contain complete text blocks)
			text := v2.ExtractAssistantText(msg)
			if text != "" {
				fmt.Print(text)
				totalChars += len(text)
			}

		case v2.V2EventTypeStreamDelta:
			// Incremental text delta (real-time streaming)
			text := v2.ExtractDeltaText(msg)
			if text != "" {
				fmt.Print(text)
				totalChars += len(text)
			}

		case v2.V2EventTypeResult:
			// Final result message
			text := v2.ExtractResultText(msg)
			if text != "" && totalChars == 0 {
				// Only print if we haven't received any streaming content
				fmt.Print(text)
				totalChars += len(text)
			}

		case v2.V2EventTypeError:
			// Error occurred
			errMsg := v2.ExtractErrorMessage(msg)
			fmt.Fprintf(os.Stderr, "\nError: %s\n", errMsg)
			os.Exit(1)
		}
	}

	// Print final statistics
	duration := time.Since(startTime)
	fmt.Println("\n----------------------------------------")
	fmt.Printf("\nSession ID: %s\n", session.SessionID())
	fmt.Printf("Characters: %d\n", totalChars)
	fmt.Printf("Duration: %v\n", duration.Round(time.Millisecond))
	if totalChars > 0 && duration > 0 {
		charsPerSec := float64(totalChars) / duration.Seconds()
		fmt.Printf("Speed: %.1f chars/sec\n", charsPerSec)
	}
}
