package v2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"agent-sdk-go/claude"
	"agent-sdk-go/claude/cli"
	"agent-sdk-go/claude/shared"
)

// Prompt sends a one-shot prompt to Claude and returns the result.
// This is equivalent to unstable_v2_prompt() in the TypeScript SDK.
//
// The prompt function creates a temporary session, sends the message,
// receives the complete response, extracts the text result, and returns it.
// The session is automatically closed after the response is received.
//
// Example:
//
//	result, err := v2.Prompt(ctx, "What is 2 + 2?",
//	    v2.WithPromptModel("claude-sonnet-4-5-20250929"),
//	    v2.WithPromptTimeout(30*time.Second))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(result.Result)
func Prompt(ctx context.Context, prompt string, opts ...PromptOption) (*V2Result, error) {
	// Apply options
	options := DefaultPromptOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Validate options
	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Check if Claude CLI is available (using injected checker for testability)
	cliChecker := options.cliChecker
	if cliChecker == nil {
		cliChecker = shared.CLICheckerFunc(cli.IsCLIAvailable)
	}
	if !cliChecker.IsCLIAvailable() {
		return nil, fmt.Errorf("claude CLI not found. Please install it first")
	}

	// Get the client factory (DIP: depend on abstraction, not concrete NewClient)
	factory := options.clientFactory
	if factory == nil {
		factory = DefaultClientFactory()
	}

	// Create a temporary client using the factory
	client, err := factory.NewClient(
		claude.WithModel(options.Model),
		claude.WithTimeout(options.Timeout.String()),
	)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	// Connect to Claude CLI
	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}
	defer func() { _ = client.Disconnect() }()

	// Generate a session ID for this one-shot query
	sessionID := fmt.Sprintf("prompt-%d", time.Now().UnixNano())

	// Send the prompt
	startTime := time.Now()

	msgChan, errChan := client.QueryStream(ctx, prompt)

	// Collect the response
	var resultText strings.Builder

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				// Channel closed - response complete
				goto Done
			}

			// Extract text from the message
			if text := ExtractText(msg); text != "" {
				resultText.WriteString(text)
			}

		case err, ok := <-errChan:
			if !ok {
				goto Done
			}
			// Return error on first error
			return nil, fmt.Errorf("receive error: %w", err)

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

Done:
	endTime := time.Now()

	// Return the result
	return &V2Result{
		Result:    resultText.String(),
		SessionID: sessionID,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  endTime.Sub(startTime),
	}, nil
}
