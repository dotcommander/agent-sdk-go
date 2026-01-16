//go:build integration

package v2

import (
	"context"
	"testing"
	"time"

	"agent-sdk-go/internal/claude/cli"
)

func TestV2Prompt(t *testing.T) {
	// Skip test if Claude CLI is not available
	if !cli.IsCLIAvailable() {
		t.Skip("Claude CLI not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test a simple prompt
	result, err := Prompt(ctx, "Hello, this is a test prompt")
	if err != nil {
		t.Fatalf("Prompt failed: %v", err)
	}

	// Verify result has expected fields
	if result.Result == "" {
		t.Error("Result should not be empty")
	}

	if result.SessionID == "" {
		t.Error("SessionID should not be empty")
	}

	if result.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}

	if result.EndTime.IsZero() {
		t.Error("EndTime should not be zero")
	}

	if result.Duration <= 0 {
		t.Error("Duration should be positive")
	}

	t.Logf("Prompt result: %s", result.Result)
}

func TestV2Session(t *testing.T) {
	// Skip test if Claude CLI is not available
	if !cli.IsCLIAvailable() {
		t.Skip("Claude CLI not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a session
	session, err := CreateSession(ctx,
		WithModel("claude-3-5-sonnet-20241022"),
		WithTimeout(30*time.Second))
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}
	defer session.Close()

	// Test session ID
	if session.SessionID() == "" {
		t.Error("Session ID should not be empty")
	}

	// Send a message
	err = session.Send(ctx, "Hello, this is a test message")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Receive response
	msgChan := session.Receive(ctx)
	msgCount := 0

	for msg := range msgChan {
		msgCount++

		// Check message type
		switch msg.Type() {
		case V2EventTypeAssistant:
			// Assistant message
			text := ExtractAssistantText(msg)
			if text == "" {
				t.Error("Assistant message should have text content")
			}
			t.Logf("Assistant: %s", text)
		case V2EventTypeResult:
			// Result message
			text := ExtractResultText(msg)
			if text == "" {
				t.Error("Result message should have text content")
			}
			t.Logf("Result: %s", text)
		case V2EventTypeError:
			t.Errorf("Error message: %s", ExtractErrorMessage(msg))
		}

		// Break after receiving a few messages
		if msgCount >= 3 {
			break
		}
	}

	if msgCount == 0 {
		t.Error("Should have received at least one message")
	}
}