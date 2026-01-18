// Package main demonstrates partial message streaming with the Claude Agent SDK.
//
// Partial streaming allows you to:
// - Receive incremental content updates in real-time
// - Build responsive UIs with live content
// - Process large responses progressively
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos
package main

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

func main() {
	fmt.Println("=== Partial Streaming Example ===")
	fmt.Println("This demonstrates receiving incremental content updates.")
	fmt.Println()

	// Show enabling partial messages
	demonstrateEnablingPartialMessages()

	// Show stream event types
	demonstrateStreamEventTypes()

	// Show processing patterns
	demonstrateProcessingPatterns()

	// Show UI integration
	demonstrateUIIntegration()

	fmt.Println()
	fmt.Println("=== Partial Streaming Example Complete ===")
}

// demonstrateEnablingPartialMessages shows how to enable partial streaming.
func demonstrateEnablingPartialMessages() {
	fmt.Println("--- Enabling Partial Messages ---")
	fmt.Println()

	fmt.Println(`
  // Enable partial message streaming
  client, err := claude.NewClient(
      claude.WithModel("claude-sonnet-4-20250514"),
      claude.WithIncludePartialMessages(true),
  )

  // Or with V2 session API
  session, err := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-20250514"),
      v2.WithIncludePartialMessages(true),
  )
`)
}

// demonstrateStreamEventTypes shows the different stream event types.
func demonstrateStreamEventTypes() {
	fmt.Println("--- Stream Event Types ---")
	fmt.Println()

	fmt.Println("1. message_start - Beginning of a message:")
	messageStart := shared.StreamEvent{
		UUID:      "msg-123",
		SessionID: "session-abc",
		Event: map[string]any{
			"type": shared.StreamEventTypeMessageStart,
			"message": map[string]any{
				"id":    "msg-123",
				"role":  "assistant",
				"model": "claude-sonnet-4-20250514",
			},
		},
	}
	printJSON("message_start", messageStart)

	fmt.Println("2. content_block_start - Beginning of a content block:")
	contentStart := shared.StreamEvent{
		UUID:      "msg-123",
		SessionID: "session-abc",
		Event: map[string]any{
			"type":  shared.StreamEventTypeContentBlockStart,
			"index": 0,
			"content_block": map[string]any{
				"type": "text",
				"text": "",
			},
		},
	}
	printJSON("content_block_start", contentStart)

	fmt.Println("3. content_block_delta - Incremental content update:")
	contentDelta := shared.StreamEvent{
		UUID:      "msg-123",
		SessionID: "session-abc",
		Event: map[string]any{
			"type":  shared.StreamEventTypeContentBlockDelta,
			"index": 0,
			"delta": map[string]any{
				"type": "text_delta",
				"text": "Here is the ",
			},
		},
	}
	printJSON("content_block_delta", contentDelta)

	fmt.Println("4. content_block_stop - End of a content block:")
	contentStop := shared.StreamEvent{
		UUID:      "msg-123",
		SessionID: "session-abc",
		Event: map[string]any{
			"type":  shared.StreamEventTypeContentBlockStop,
			"index": 0,
		},
	}
	printJSON("content_block_stop", contentStop)

	fmt.Println("5. message_delta - Message metadata update:")
	messageDelta := shared.StreamEvent{
		UUID:      "msg-123",
		SessionID: "session-abc",
		Event: map[string]any{
			"type": shared.StreamEventTypeMessageDelta,
			"delta": map[string]any{
				"stop_reason": "end_turn",
			},
			"usage": map[string]any{
				"output_tokens": 150,
			},
		},
	}
	printJSON("message_delta", messageDelta)

	fmt.Println("6. message_stop - End of message:")
	messageStop := shared.StreamEvent{
		UUID:      "msg-123",
		SessionID: "session-abc",
		Event: map[string]any{
			"type": shared.StreamEventTypeMessageStop,
		},
	}
	printJSON("message_stop", messageStop)
}

// demonstrateProcessingPatterns shows patterns for processing stream events.
func demonstrateProcessingPatterns() {
	fmt.Println("--- Processing Patterns ---")
	fmt.Println()

	fmt.Println("1. Basic delta accumulation:")
	fmt.Println(`
  var fullText strings.Builder

  iter := client.ReceiveResponseIterator(ctx)
  for {
      msg, err := iter.Next(ctx)
      if errors.Is(err, claude.ErrNoMoreMessages) {
          break
      }

      // Check for stream events (partial messages)
      if event, ok := msg.(*shared.StreamEvent); ok {
          eventType := event.Event["type"].(string)

          switch eventType {
          case shared.StreamEventTypeContentBlockDelta:
              delta := event.Event["delta"].(map[string]any)
              if text, ok := delta["text"].(string); ok {
                  fullText.WriteString(text)
                  fmt.Print(text) // Print as it arrives
              }

          case shared.StreamEventTypeMessageStop:
              fmt.Println("\n---Message Complete---")
          }
      }
  }

  fmt.Println("Full response:", fullText.String())
`)

	fmt.Println("2. Tracking multiple content blocks:")
	fmt.Println(`
  type ContentAccumulator struct {
      blocks map[int]*strings.Builder
  }

  func (c *ContentAccumulator) ProcessEvent(event *shared.StreamEvent) {
      eventType := event.Event["type"].(string)

      switch eventType {
      case shared.StreamEventTypeContentBlockStart:
          index := int(event.Event["index"].(float64))
          c.blocks[index] = &strings.Builder{}

      case shared.StreamEventTypeContentBlockDelta:
          index := int(event.Event["index"].(float64))
          delta := event.Event["delta"].(map[string]any)
          if text, ok := delta["text"].(string); ok {
              c.blocks[index].WriteString(text)
          }
      }
  }
`)

	fmt.Println("3. Handling thinking blocks:")
	fmt.Println(`
  iter := client.ReceiveResponseIterator(ctx)
  for {
      msg, err := iter.Next(ctx)
      if errors.Is(err, claude.ErrNoMoreMessages) {
          break
      }

      if event, ok := msg.(*shared.StreamEvent); ok {
          if event.Event["type"] == shared.StreamEventTypeContentBlockStart {
              block := event.Event["content_block"].(map[string]any)
              blockType := block["type"].(string)

              if blockType == "thinking" {
                  fmt.Println("[Thinking...]")
              } else if blockType == "text" {
                  fmt.Println("[Response:]")
              }
          }
      }
  }
`)
}

// demonstrateUIIntegration shows patterns for UI integration.
func demonstrateUIIntegration() {
	fmt.Println("--- UI Integration Patterns ---")
	fmt.Println()

	fmt.Println("1. Real-time typing effect:")
	fmt.Println(`
  // Channel for UI updates
  uiUpdates := make(chan string, 100)

  go func() {
      iter := client.ReceiveResponseIterator(ctx)
      for {
          msg, err := iter.Next(ctx)
          if errors.Is(err, claude.ErrNoMoreMessages) {
              close(uiUpdates)
              return
          }

          if event, ok := msg.(*shared.StreamEvent); ok {
              if event.Event["type"] == shared.StreamEventTypeContentBlockDelta {
                  delta := event.Event["delta"].(map[string]any)
                  if text, ok := delta["text"].(string); ok {
                      uiUpdates <- text
                  }
              }
          }
      }
  }()

  // Render in UI
  for text := range uiUpdates {
      ui.AppendText(text)
      ui.ScrollToBottom()
  }
`)

	fmt.Println("2. Progress indicator:")
	fmt.Println(`
  type StreamProgress struct {
      Started    bool
      BlockCount int
      TokenCount int
      Complete   bool
  }

  progress := &StreamProgress{}

  iter := client.ReceiveResponseIterator(ctx)
  for {
      msg, err := iter.Next(ctx)
      if errors.Is(err, claude.ErrNoMoreMessages) {
          break
      }

      if event, ok := msg.(*shared.StreamEvent); ok {
          switch event.Event["type"].(string) {
          case shared.StreamEventTypeMessageStart:
              progress.Started = true
              ui.ShowSpinner("Thinking...")

          case shared.StreamEventTypeContentBlockStart:
              progress.BlockCount++
              ui.UpdateStatus(fmt.Sprintf("Block %d", progress.BlockCount))

          case shared.StreamEventTypeMessageDelta:
              if usage, ok := event.Event["usage"].(map[string]any); ok {
                  progress.TokenCount = int(usage["output_tokens"].(float64))
                  ui.UpdateStatus(fmt.Sprintf("%d tokens", progress.TokenCount))
              }

          case shared.StreamEventTypeMessageStop:
              progress.Complete = true
              ui.HideSpinner()
          }
      }
  }
`)

	fmt.Println("3. Cancelable streaming:")
	fmt.Println(`
  ctx, cancel := context.WithCancel(context.Background())

  // Cancel button handler
  ui.OnCancel(func() {
      cancel()
      client.Interrupt()
  })

  iter := client.ReceiveResponseIterator(ctx)
  for {
      msg, err := iter.Next(ctx)
      if err != nil {
          if errors.Is(err, context.Canceled) {
              ui.ShowMessage("Response cancelled")
          }
          break
      }
      // Process message...
  }
`)
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
