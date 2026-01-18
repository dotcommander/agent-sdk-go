// Package main demonstrates a basic one-shot query using the Claude Agent SDK.
//
// This example shows:
// - Using v2.Prompt() for one-shot queries
// - Handling the result with timing information
// - Proper error handling
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world-v2
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/claude/v2"
)

func main() {
	// Check CLI availability first
	if !cli.IsCLIAvailable() {
		log.Fatal("Claude CLI not found. Please install it first: https://docs.anthropic.com/claude/docs/quickstart")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get the prompt from command line or use default
	prompt := "What is the capital of France? Answer in one word."
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	}

	fmt.Println("=== Basic Query Example ===")
	fmt.Printf("Prompt: %s\n\n", prompt)

	// Send the one-shot prompt
	result, err := v2.Prompt(ctx, prompt,
		v2.WithPromptModel("claude-sonnet-4-20250514"),
		v2.WithPromptTimeout(60*time.Second),
	)
	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}

	// Display the result
	fmt.Println("Response:")
	fmt.Println("----------------------------------------")
	fmt.Println(result.Result)
	fmt.Println("----------------------------------------")
	fmt.Printf("\nSession ID: %s\n", result.SessionID)
	fmt.Printf("Duration: %v\n", result.Duration.Round(time.Millisecond))
}
