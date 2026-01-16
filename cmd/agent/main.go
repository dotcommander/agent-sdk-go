package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"agent-sdk-go/internal/claude"
	"agent-sdk-go/internal/claude/cli"
	"agent-sdk-go/internal/claude/shared"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCmd(os.Args[2:])
	case "tool":
		toolCmd(os.Args[2:])
	case "stream":
		streamCmd(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: agent-sdk-go-agent <subcommand> [flags]

Subcommands:
  run     Send a simple message and print the response
  tool    Test tool execution with a calculator tool
  stream  Stream a response and print deltas in real-time

Flags for all commands:
  -h, --help  Show help for the subcommand

Examples:
  agent-sdk-go-agent run
  agent-sdk-go-agent tool
  agent-sdk-go-agent stream

`)
}

func runCmd(args []string) {
	cmd := flag.NewFlagSet("run", flag.ExitOnError)
	model := cmd.String("model", "claude-3-5-sonnet-20241022", "Model to use")
	message := cmd.String("message", "Hello, Claude!", "Message to send")
	timeout := cmd.Int("timeout", 60, "Timeout in seconds")
	if err := cmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		os.Exit(1)
	}

	// Check if Claude CLI is available
	if !cli.IsCLIAvailable() {
		fmt.Fprintf(os.Stderr, "Claude CLI not found. Please install it first.\n")
		fmt.Fprintf(os.Stderr, "Installation: https://github.com/anthropics/claude-cli\n")
		os.Exit(1)
	}

	// Create client with options
	client, err := claude.NewClient(
		claude.WithModel(*model),
		claude.WithTimeout(fmt.Sprintf("%ds", *timeout)),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create client: %v\n", err)
		os.Exit(1)
	}

	// Connect to Claude CLI
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect() }()

	// Send query
	response, err := client.Query(ctx, *message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "query: %v\n", err)
		os.Exit(1)
	}

	// Print response
	fmt.Println(response)
}

func toolCmd(args []string) {
	cmd := flag.NewFlagSet("tool", flag.ExitOnError)
	model := cmd.String("model", "claude-3-5-sonnet-20241022", "Model to use")
	message := cmd.String("message", "What's 2 + 2?", "Message to send")
	timeout := cmd.Int("timeout", 60, "Timeout in seconds")
	if err := cmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		os.Exit(1)
	}

	// Check if Claude CLI is available
	if !cli.IsCLIAvailable() {
		fmt.Fprintf(os.Stderr, "Claude CLI not found. Please install it first.\n")
		fmt.Fprintf(os.Stderr, "Installation: https://github.com/anthropics/claude-cli\n")
		os.Exit(1)
	}

	// Create client with options
	client, err := claude.NewClient(
		claude.WithModel(*model),
		claude.WithTimeout(fmt.Sprintf("%ds", *timeout)),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create client: %v\n", err)
		os.Exit(1)
	}

	// Connect to Claude CLI
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect() }()

	// Send query and receive response
	msgChan, errChan := client.QueryStream(ctx, *message)

	fmt.Println("Response (watching for tool uses):")

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				goto Done
			}

			// Check for tool uses in the message
			if shared.HasToolUses(msg) {
				toolUses := shared.ExtractToolUses(msg)
				fmt.Printf("\n[Detected %d tool use(s)]\n", len(toolUses))
				for i, toolUse := range toolUses {
					fmt.Printf("  Tool %d: %s\n", i+1, toolUse.Name)
					fmt.Printf("    ID: %s\n", toolUse.ToolUseID)
					fmt.Printf("    Input: %v\n", toolUse.Input)
				}
			}

			// Print text content
			if assistantMsg, ok := msg.(*claude.AssistantMessage); ok {
				for _, block := range assistantMsg.Content {
					if textBlock, ok := block.(*claude.TextBlock); ok {
						fmt.Print(textBlock.Text)
					}
				}
			}

		case err, ok := <-errChan:
			if !ok {
				goto Done
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nerror: %v\n", err)
			}
			goto Done
		}
	}

Done:
	fmt.Println("\n\nNote: In the subprocess SDK, tools are executed by the Claude CLI itself.")
	fmt.Println("The SDK can detect tool uses in messages for monitoring purposes.")
}

func streamCmd(args []string) {
	cmd := flag.NewFlagSet("stream", flag.ExitOnError)
	model := cmd.String("model", "claude-3-5-sonnet-20241022", "Model to use")
	message := cmd.String("message", "Tell me a short story about robots.", "Message to send")
	timeout := cmd.Int("timeout", 60, "Stream timeout in seconds")
	if err := cmd.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "parse flags: %v\n", err)
		os.Exit(1)
	}

	// Check if Claude CLI is available
	if !cli.IsCLIAvailable() {
		fmt.Fprintf(os.Stderr, "Claude CLI not found. Please install it first.\n")
		fmt.Fprintf(os.Stderr, "Installation: https://github.com/anthropics/claude-cli\n")
		os.Exit(1)
	}

	// Create client with options
	client, err := claude.NewClient(
		claude.WithModel(*model),
		claude.WithTimeout(fmt.Sprintf("%ds", *timeout)),
		claude.WithIncludePartialMessages(true),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create client: %v\n", err)
		os.Exit(1)
	}

	// Connect to Claude CLI
	ctx := context.Background()
	if err := client.Connect(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = client.Disconnect() }()

	// Stream query
	msgChan, errChan := client.QueryStream(ctx, *message)

	var responseText strings.Builder
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				// Channel closed
				goto Done
			}

			// Handle different message types
			switch m := msg.(type) {
			case *claude.AssistantMessage:
				// Extract text from content blocks
				for _, block := range m.Content {
					if textBlock, ok := block.(*claude.TextBlock); ok {
						fmt.Print(textBlock.Text)
						responseText.WriteString(textBlock.Text)
					}
				}
			case *claude.StreamEvent:
				// Handle stream events
				if textDelta, err := shared.ExtractDelta(m.Event); err == nil && textDelta != "" {
					fmt.Print(textDelta)
					responseText.WriteString(textDelta)
				}
			}

		case err, ok := <-errChan:
			if !ok {
				// Channel closed
				goto Done
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "\nstream error: %v\n", err)
			}
			goto Done
		}
	}

Done:
	fmt.Println() // New line after stream

	if responseText.Len() > 0 {
		fmt.Printf("\nTotal response length: %d characters\n", responseText.Len())
	}
}