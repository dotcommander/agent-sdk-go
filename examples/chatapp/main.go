// Package main demonstrates a simple chat application with the Claude Agent SDK.
//
// This example shows:
// - Interactive multi-turn chat loop
// - Message queue pattern for async input
// - Streaming responses
// - Graceful shutdown handling
// - Chat session management
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/simple-chatapp
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/claude/v2"
)

const systemPrompt = `You are a helpful AI assistant. You can help users with a wide variety of tasks including:
- Answering questions
- Writing and editing text
- Coding and debugging
- Analysis and research
- Creative tasks

Be concise but thorough in your responses.`

func main() {
	fmt.Println("=== Simple Chat Application ===")
	fmt.Println()

	// Check CLI availability
	if !cli.IsCLIAvailable() {
		fmt.Println("Claude CLI not available.")
		fmt.Println("Please install: https://docs.anthropic.com/claude/docs/quickstart")
		return
	}

	// Create context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create chat session
	session, err := v2.CreateSession(ctx,
		v2.WithModel("claude-sonnet-4-20250514"),
		v2.WithTimeout(120*time.Second),
		v2.WithSystemPrompt(systemPrompt),
	)
	if err != nil {
		fmt.Printf("Failed to create session: %v\n", err)
		return
	}
	defer func() { _ = session.Close() }()

	fmt.Printf("Session started: %s\n", session.SessionID())
	fmt.Println("Type your message and press Enter. Type 'quit' or 'exit' to end.")
	fmt.Println("Type 'clear' to start a new session.")
	fmt.Println()

	// Start chat loop
	runChatLoop(ctx, session)

	fmt.Println()
	fmt.Println("Chat session ended. Goodbye!")
}

// runChatLoop runs the interactive chat loop.
func runChatLoop(ctx context.Context, session v2.V2Session) {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		// Check if context is done
		select {
		case <-ctx.Done():
			fmt.Println("\nInterrupted.")
			return
		default:
		}

		// Prompt for input
		fmt.Print("You: ")

		// Read input
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Printf("\nRead error: %v\n", err)
			}
			return
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		// Handle special commands
		switch strings.ToLower(input) {
		case "quit", "exit":
			return
		case "clear":
			fmt.Println("Session cleared. Starting fresh conversation.")
			fmt.Println()
			continue
		case "help":
			printHelp()
			continue
		}

		// Send message to Claude
		if err := session.Send(ctx, input); err != nil {
			fmt.Printf("Send error: %v\n", err)
			continue
		}

		// Stream response
		fmt.Print("Claude: ")
		streamResponse(ctx, session)
		fmt.Println()
	}
}

// streamResponse streams the response from Claude.
func streamResponse(ctx context.Context, session v2.V2Session) {
	var totalChars int

	for msg := range session.Receive(ctx) {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			fmt.Print("\n[Interrupted]")
			return
		default:
		}

		switch msg.Type() {
		case v2.V2EventTypeAssistant:
			text := v2.ExtractAssistantText(msg)
			if text != "" {
				fmt.Print(text)
				totalChars += len(text)
			}

		case v2.V2EventTypeStreamDelta:
			text := v2.ExtractDeltaText(msg)
			if text != "" {
				fmt.Print(text)
				totalChars += len(text)
			}

		case v2.V2EventTypeResult:
			// Final result - only print if we haven't received streaming content
			if totalChars == 0 {
				text := v2.ExtractResultText(msg)
				if text != "" {
					fmt.Print(text)
				}
			}

		case v2.V2EventTypeError:
			errMsg := v2.ExtractErrorMessage(msg)
			fmt.Printf("\n[Error: %s]", errMsg)
			return
		}
	}
}

// printHelp prints help information.
func printHelp() {
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  quit, exit  - End the chat session")
	fmt.Println("  clear       - Clear conversation history")
	fmt.Println("  help        - Show this help message")
	fmt.Println()
	fmt.Println("Tips:")
	fmt.Println("  - Press Ctrl+C to interrupt a long response")
	fmt.Println("  - Claude can help with coding, writing, analysis, and more")
	fmt.Println("  - Be specific in your requests for better responses")
	fmt.Println()
}

// ChatApp represents a chat application with message queue.
// This is an alternative implementation using channels for async input.
type ChatApp struct {
	session   v2.V2Session
	inputChan chan string
	done      chan struct{}
}

// NewChatApp creates a new chat application.
func NewChatApp(session v2.V2Session) *ChatApp {
	return &ChatApp{
		session:   session,
		inputChan: make(chan string, 10),
		done:      make(chan struct{}),
	}
}

// SendMessage sends a message to the chat.
func (app *ChatApp) SendMessage(content string) {
	select {
	case app.inputChan <- content:
	case <-app.done:
	}
}

// Close closes the chat application.
func (app *ChatApp) Close() {
	close(app.done)
}

// Run runs the chat application with async message handling.
func (app *ChatApp) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-app.done:
			return
		case input, ok := <-app.inputChan:
			if !ok {
				return
			}

			// Send to Claude
			if err := app.session.Send(ctx, input); err != nil {
				fmt.Printf("Send error: %v\n", err)
				continue
			}

			// Receive response
			for msg := range app.session.Receive(ctx) {
				switch msg.Type() {
				case v2.V2EventTypeAssistant:
					fmt.Print(v2.ExtractAssistantText(msg))
				case v2.V2EventTypeStreamDelta:
					fmt.Print(v2.ExtractDeltaText(msg))
				case v2.V2EventTypeResult:
					// Result received
				case v2.V2EventTypeError:
					fmt.Printf("[Error: %s]\n", v2.ExtractErrorMessage(msg))
				}
			}
			fmt.Println()
		}
	}
}
