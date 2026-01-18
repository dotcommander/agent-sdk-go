// Package main demonstrates multi-turn conversations using the Claude Agent SDK.
//
// This example shows:
// - Creating a persistent session
// - Maintaining context across multiple turns
// - Session resume capability
// - Interactive conversation flow
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world-v2
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
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

	// Check for demo mode
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		runDemo(ctx)
		return
	}

	// Interactive mode
	runInteractive(ctx)
}

// runDemo demonstrates multi-turn conversation with context preservation
func runDemo(ctx context.Context) {
	fmt.Println("=== Multi-Turn Demo ===")
	fmt.Println("This demonstrates how Claude maintains context across conversation turns.")
	fmt.Println()

	// Create a session
	session, err := v2.CreateSession(ctx,
		v2.WithModel("claude-sonnet-4-20250514"),
		v2.WithTimeout(60*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer func() { _ = session.Close() }()

	fmt.Printf("Session ID: %s\n\n", session.SessionID())

	// Turn 1: Establish a fact
	fmt.Println("--- Turn 1 ---")
	response1 := sendAndReceive(ctx, session, "My name is Alex and my favorite number is 42. Please acknowledge.")
	fmt.Printf("User: My name is Alex and my favorite number is 42. Please acknowledge.\n")
	fmt.Printf("Claude: %s\n\n", response1)

	// Turn 2: Reference the fact
	fmt.Println("--- Turn 2 ---")
	response2 := sendAndReceive(ctx, session, "What is my name?")
	fmt.Printf("User: What is my name?\n")
	fmt.Printf("Claude: %s\n\n", response2)

	// Turn 3: Reference another fact
	fmt.Println("--- Turn 3 ---")
	response3 := sendAndReceive(ctx, session, "What is my favorite number multiplied by 2?")
	fmt.Printf("User: What is my favorite number multiplied by 2?\n")
	fmt.Printf("Claude: %s\n\n", response3)

	fmt.Println("=== Demo Complete ===")
	fmt.Println("Claude maintained context across all three turns!")
}

// runInteractive provides an interactive chat interface
func runInteractive(ctx context.Context) {
	fmt.Println("=== Interactive Multi-Turn Chat ===")
	fmt.Println("Commands:")
	fmt.Println("  'quit' or 'exit' - End the session")
	fmt.Println("  'new' - Start a new session")
	fmt.Println("  'id' - Show current session ID")
	fmt.Println("----------------------------------------")

	// Create initial session
	session, err := v2.CreateSession(ctx,
		v2.WithModel("claude-sonnet-4-20250514"),
		v2.WithTimeout(120*time.Second),
	)
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}
	defer func() { _ = session.Close() }()

	fmt.Printf("Session started: %s\n\n", session.SessionID())

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(input) {
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		case "new":
			// Close current session and create a new one
			_ = session.Close()
			session, err = v2.CreateSession(ctx,
				v2.WithModel("claude-sonnet-4-20250514"),
				v2.WithTimeout(120*time.Second),
			)
			if err != nil {
				log.Fatalf("Failed to create new session: %v", err)
			}
			fmt.Printf("\nNew session started: %s\n\n", session.SessionID())
			continue
		case "id":
			fmt.Printf("\nCurrent session: %s\n\n", session.SessionID())
			continue
		}

		// Send message and stream response
		if err := session.Send(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "Error sending message: %v\n", err)
			continue
		}

		fmt.Print("Claude: ")
		var responseText strings.Builder

		for msg := range session.Receive(ctx) {
			switch msg.Type() {
			case v2.V2EventTypeAssistant:
				text := v2.ExtractAssistantText(msg)
				if text != "" {
					fmt.Print(text)
					responseText.WriteString(text)
				}
			case v2.V2EventTypeStreamDelta:
				text := v2.ExtractDeltaText(msg)
				if text != "" {
					fmt.Print(text)
					responseText.WriteString(text)
				}
			case v2.V2EventTypeResult:
				text := v2.ExtractResultText(msg)
				if text != "" && responseText.Len() == 0 {
					fmt.Print(text)
				}
			case v2.V2EventTypeError:
				fmt.Fprintf(os.Stderr, "\nError: %s\n", v2.ExtractErrorMessage(msg))
			}
		}

		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Input error: %v", err)
	}
}

// sendAndReceive is a helper that sends a message and collects the full response
func sendAndReceive(ctx context.Context, session v2.V2Session, message string) string {
	if err := session.Send(ctx, message); err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	var result strings.Builder

	for msg := range session.Receive(ctx) {
		switch msg.Type() {
		case v2.V2EventTypeAssistant:
			result.WriteString(v2.ExtractAssistantText(msg))
		case v2.V2EventTypeStreamDelta:
			result.WriteString(v2.ExtractDeltaText(msg))
		case v2.V2EventTypeResult:
			if result.Len() == 0 {
				result.WriteString(v2.ExtractResultText(msg))
			}
		case v2.V2EventTypeError:
			return fmt.Sprintf("Error: %s", v2.ExtractErrorMessage(msg))
		}
	}

	return strings.TrimSpace(result.String())
}
