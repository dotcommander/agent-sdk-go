// Package main provides a demo CLI tool showcasing the agent-sdk-go internal packages.
//
// This demo demonstrates key features of the V2 SDK:
//   - One-shot prompts via v2.Prompt()
//   - Interactive sessions via v2.CreateSession()
//   - Streaming responses with real-time output
//   - Proper error handling and CLI availability checks
//
// Usage:
//
//	demo prompt "What is 2 + 2?"              # One-shot prompt
//	demo chat                                  # Interactive chat session
//	demo stream "Tell me a story"             # Streaming response
//	demo check                                 # Check CLI availability
//
// Build:
//
//	go build -o demo ./cmd/demo
//
// Install:
//
//	ln -sf $(pwd)/demo ~/go/bin/demo
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

const (
	defaultModel   = "claude-sonnet-4-5-20250929"
	defaultTimeout = 60 * time.Second
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var err error
	switch os.Args[1] {
	case "prompt":
		err = runPrompt(ctx, os.Args[2:])
	case "chat":
		err = runChat(ctx, os.Args[2:])
	case "stream":
		err = runStream(ctx, os.Args[2:])
	case "check":
		err = runCheck()
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(`demo - showcase agent-sdk-go internal packages

USAGE:
    demo <command> [arguments]

COMMANDS:
    prompt <question>    Send a one-shot prompt and get a response
    chat                 Start an interactive chat session
    stream <question>    Stream a response with real-time output
    check                Check if Claude CLI is available

EXAMPLES:
    demo prompt "What is the capital of France?"
    demo chat
    demo stream "Tell me a short story about a robot"
    demo check

FLAGS:
    --model <model>      Model to use (default: claude-sonnet-4-5-20250929)
    --timeout <seconds>  Timeout in seconds (default: 60)

NOTES:
    This demo requires the Claude CLI to be installed and available in PATH.
    Install: https://docs.anthropic.com/claude/docs/quickstart#installing-claude

`)
}

// runPrompt demonstrates one-shot prompts via v2.Prompt()
func runPrompt(ctx context.Context, args []string) error {
	model := defaultModel
	timeout := defaultTimeout
	var prompt string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--model":
			if i+1 >= len(args) {
				return fmt.Errorf("--model requires a value")
			}
			i++
			model = args[i]
		case "--timeout":
			if i+1 >= len(args) {
				return fmt.Errorf("--timeout requires a value")
			}
			i++
			var seconds int
			if _, err := fmt.Sscanf(args[i], "%d", &seconds); err != nil {
				return fmt.Errorf("invalid timeout value: %s", args[i])
			}
			timeout = time.Duration(seconds) * time.Second
		default:
			if strings.HasPrefix(args[i], "-") {
				return fmt.Errorf("unknown flag: %s", args[i])
			}
			prompt = args[i]
		}
	}

	if prompt == "" {
		return fmt.Errorf("prompt is required\n\nUsage: demo prompt <question>")
	}

	// Check CLI availability first
	if !cli.IsCLIAvailable() {
		return cliNotAvailableError()
	}

	fmt.Printf("Sending prompt to %s...\n\n", model)

	result, err := v2.Prompt(ctx, prompt,
		v2.WithPromptModel(model),
		v2.WithPromptTimeout(timeout),
	)
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	fmt.Println("Response:")
	fmt.Println(strings.Repeat("-", 40))
	fmt.Println(result.Result)
	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("\nSession: %s\n", result.SessionID)
	fmt.Printf("Duration: %v\n", result.Duration.Round(time.Millisecond))

	return nil
}

// runChat demonstrates interactive sessions via v2.CreateSession()
func runChat(ctx context.Context, args []string) error {
	model := defaultModel
	timeout := defaultTimeout

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--model":
			if i+1 >= len(args) {
				return fmt.Errorf("--model requires a value")
			}
			i++
			model = args[i]
		case "--timeout":
			if i+1 >= len(args) {
				return fmt.Errorf("--timeout requires a value")
			}
			i++
			var seconds int
			if _, err := fmt.Sscanf(args[i], "%d", &seconds); err != nil {
				return fmt.Errorf("invalid timeout value: %s", args[i])
			}
			timeout = time.Duration(seconds) * time.Second
		default:
			if strings.HasPrefix(args[i], "-") {
				return fmt.Errorf("unknown flag: %s", args[i])
			}
		}
	}

	// Check CLI availability first
	if !cli.IsCLIAvailable() {
		return cliNotAvailableError()
	}

	fmt.Printf("Starting interactive chat session with %s\n", model)
	fmt.Println("Type 'quit' or 'exit' to end the session")
	fmt.Println("Type 'clear' to start a new session")
	fmt.Println(strings.Repeat("-", 40))

	// Create session
	session, err := v2.CreateSession(ctx,
		v2.WithModel(model),
		v2.WithTimeout(timeout),
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	defer func() { _ = session.Close() }()

	fmt.Printf("Session ID: %s\n\n", session.SessionID())

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
			return nil
		case "clear":
			// Close current session and create a new one
			_ = session.Close()
			session, err = v2.CreateSession(ctx,
				v2.WithModel(model),
				v2.WithTimeout(timeout),
			)
			if err != nil {
				return fmt.Errorf("create new session: %w", err)
			}
			fmt.Printf("New session started: %s\n\n", session.SessionID())
			continue
		}

		// Send message
		if err := session.Send(ctx, input); err != nil {
			fmt.Fprintf(os.Stderr, "send error: %v\n", err)
			continue
		}

		// Receive and display response
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
			case v2.V2EventTypeResult:
				text := v2.ExtractResultText(msg)
				if text != "" && responseText.Len() == 0 {
					fmt.Print(text)
				}
			case v2.V2EventTypeStreamDelta:
				text := v2.ExtractDeltaText(msg)
				if text != "" {
					fmt.Print(text)
					responseText.WriteString(text)
				}
			case v2.V2EventTypeError:
				errMsg := v2.ExtractErrorMessage(msg)
				fmt.Fprintf(os.Stderr, "\nerror: %s\n", errMsg)
			}
		}

		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	return nil
}

// runStream demonstrates streaming responses with real-time output
func runStream(ctx context.Context, args []string) error {
	model := defaultModel
	timeout := defaultTimeout
	var prompt string

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--model":
			if i+1 >= len(args) {
				return fmt.Errorf("--model requires a value")
			}
			i++
			model = args[i]
		case "--timeout":
			if i+1 >= len(args) {
				return fmt.Errorf("--timeout requires a value")
			}
			i++
			var seconds int
			if _, err := fmt.Sscanf(args[i], "%d", &seconds); err != nil {
				return fmt.Errorf("invalid timeout value: %s", args[i])
			}
			timeout = time.Duration(seconds) * time.Second
		default:
			if strings.HasPrefix(args[i], "-") {
				return fmt.Errorf("unknown flag: %s", args[i])
			}
			prompt = args[i]
		}
	}

	if prompt == "" {
		return fmt.Errorf("prompt is required\n\nUsage: demo stream <question>")
	}

	// Check CLI availability first
	if !cli.IsCLIAvailable() {
		return cliNotAvailableError()
	}

	fmt.Printf("Streaming response from %s...\n\n", model)

	// Create session for streaming
	session, err := v2.CreateSession(ctx,
		v2.WithModel(model),
		v2.WithTimeout(timeout),
		v2.WithEnablePartialMessages(true),
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	defer func() { _ = session.Close() }()

	// Send the prompt
	if err := session.Send(ctx, prompt); err != nil {
		return fmt.Errorf("send: %w", err)
	}

	// Stream the response
	var totalChars int
	startTime := time.Now()

	for msg := range session.Receive(ctx) {
		switch msg.Type() {
		case v2.V2EventTypeAssistant:
			text := v2.ExtractAssistantText(msg)
			if text != "" {
				fmt.Print(text)
				totalChars += len(text)
			}
		case v2.V2EventTypeResult:
			text := v2.ExtractResultText(msg)
			if text != "" && totalChars == 0 {
				fmt.Print(text)
				totalChars += len(text)
			}
		case v2.V2EventTypeStreamDelta:
			text := v2.ExtractDeltaText(msg)
			if text != "" {
				fmt.Print(text)
				totalChars += len(text)
			}
		case v2.V2EventTypeError:
			errMsg := v2.ExtractErrorMessage(msg)
			return fmt.Errorf("stream error: %s", errMsg)
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("\n\n%s\n", strings.Repeat("-", 40))
	fmt.Printf("Session: %s\n", session.SessionID())
	fmt.Printf("Characters: %d\n", totalChars)
	fmt.Printf("Duration: %v\n", duration.Round(time.Millisecond))
	if totalChars > 0 && duration > 0 {
		charsPerSec := float64(totalChars) / duration.Seconds()
		fmt.Printf("Speed: %.1f chars/sec\n", charsPerSec)
	}

	return nil
}

// runCheck verifies Claude CLI availability and shows discovery info
func runCheck() error {
	fmt.Println("Checking Claude CLI availability...")
	fmt.Println()

	result, err := cli.DiscoverCLI("", cli.GetDefaultCommand())
	if err != nil {
		fmt.Println("Status: NOT FOUND")
		fmt.Println()
		fmt.Println("Claude CLI is not available in your PATH.")
		fmt.Println()
		fmt.Println("Installation suggestions:")
		for _, cmd := range cli.GetSuggestedCommands() {
			fmt.Printf("  %s\n", cmd)
		}
		return nil
	}

	cli.PrintDiscoveryInfo(result)

	// Validate the CLI
	fmt.Println()
	fmt.Print("Validating CLI... ")
	if err := cli.ValidateCLI(result); err != nil {
		fmt.Printf("FAILED: %v\n", err)
		return nil
	}
	fmt.Println("OK")

	fmt.Println()
	fmt.Println("Claude CLI is ready to use!")

	return nil
}

// cliNotAvailableError returns a user-friendly error for missing CLI
func cliNotAvailableError() error {
	return fmt.Errorf(`Claude CLI not found

The Claude CLI must be installed and available in PATH to use this demo.

Installation:
  https://docs.anthropic.com/claude/docs/quickstart#installing-claude

Check availability:
  demo check`)
}
