// Package main demonstrates file checkpointing with the Claude Agent SDK.
//
// File checkpointing allows you to:
// - Track file changes during a session
// - Rewind files to their state at any previous user message
// - Recover from unwanted changes
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos
package main

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude"
)

func main() {
	fmt.Println("=== File Checkpointing Example ===")
	fmt.Println("This demonstrates file checkpoint and rewind capabilities.")
	fmt.Println()

	// Show how to enable file checkpointing
	demonstrateEnablingCheckpoints()

	// Show capturing message UUIDs for rewind
	demonstrateCapturingUUIDs()

	// Show the rewind workflow
	demonstrateRewindWorkflow()

	// Show practical use cases
	demonstrateUseCases()

	fmt.Println()
	fmt.Println("=== File Checkpointing Example Complete ===")
}

// demonstrateEnablingCheckpoints shows how to enable file checkpointing.
func demonstrateEnablingCheckpoints() {
	fmt.Println("--- Enabling File Checkpointing ---")
	fmt.Println()
	fmt.Print(`
  // Enable file checkpointing when creating the client
  client, err := claude.NewClient(
      claude.WithModel("claude-sonnet-4-20250514"),
      claude.WithFileCheckpointing(true),  // Enable checkpointing
  )

  // Or use the v2 session API
  session, err := v2.CreateSession(ctx,
      v2.WithModel("claude-sonnet-4-20250514"),
      v2.WithFileCheckpointing(true),
  )
`)
	fmt.Println()
}

// demonstrateCapturingUUIDs shows how to capture message UUIDs.
func demonstrateCapturingUUIDs() {
	fmt.Println("--- Capturing Message UUIDs ---")
	fmt.Println("Every UserMessage includes a UUID that can be used for rewind.")
	fmt.Println()

	// Simulated user message with UUID
	userMsg := claude.UserMessage{
		MessageType: "user",
		Content:     "Please refactor the authentication module",
		UUID:        stringPtr("msg-uuid-abc123"),
	}

	printJSON("UserMessage with UUID", userMsg)

	fmt.Print(`
  // Capture UUID from received messages
  var checkpointUUID string

  iter := client.ReceiveResponseIterator(ctx)
  for {
      msg, err := iter.Next(ctx)
      if errors.Is(err, claude.ErrNoMoreMessages) {
          break
      }

      // Capture UUID from user messages
      if userMsg, ok := msg.(*claude.UserMessage); ok && userMsg.UUID != nil {
          checkpointUUID = *userMsg.UUID
          fmt.Printf("Checkpoint created: %s\n", checkpointUUID)
      }
  }
`)
	fmt.Println()
}

// demonstrateRewindWorkflow shows the rewind process.
func demonstrateRewindWorkflow() {
	fmt.Println("--- Rewind Workflow ---")
	fmt.Println()
	fmt.Print(`
  // Step 1: Store checkpoint UUIDs as you work
  checkpoints := make(map[string]string)

  // Step 2: Before a risky operation, note the checkpoint
  checkpoints["before_refactor"] = currentUUID

  // Step 3: If something goes wrong, rewind
  err := client.RewindFiles(ctx, checkpoints["before_refactor"])
  if err != nil {
      log.Printf("Rewind failed: %v", err)
      return
  }
  fmt.Println("Files reverted to checkpoint state")
`)
	fmt.Println()

	fmt.Println("Example: Interactive rewind prompt")
	fmt.Print(`
  // Offer rewind to user after a complex operation
  func maybeRewind(ctx context.Context, client claude.Client, checkpointUUID string) error {
      fmt.Print("Do you want to keep these changes? [y/n]: ")
      var response string
      fmt.Scanln(&response)

      if response == "n" {
          return client.RewindFiles(ctx, checkpointUUID)
      }
      return nil
  }
`)
	fmt.Println()
}

// demonstrateUseCases shows practical use cases.
func demonstrateUseCases() {
	fmt.Println("--- Practical Use Cases ---")
	fmt.Println()

	fmt.Println("1. Safe Refactoring")
	fmt.Print(`
  // Checkpoint before each refactoring phase
  phases := []string{"rename_variables", "extract_functions", "update_tests"}
  checkpoints := make(map[string]string)

  for _, phase := range phases {
      checkpoints[phase] = currentUUID
      // ... perform refactoring ...
      // If tests fail, rewind to last checkpoint
  }
`)
	fmt.Println()

	fmt.Println("2. Exploratory Changes")
	fmt.Print(`
  // Try different approaches, rewind if unsatisfactory
  approaches := []string{"approach_a", "approach_b", "approach_c"}
  var bestCheckpoint string

  for _, approach := range approaches {
      client.RewindFiles(ctx, baseCheckpoint)
      // ... implement approach ...
      // ... evaluate results ...
      if betterThanBest {
          bestCheckpoint = currentUUID
      }
  }
  // Keep the best approach
  client.RewindFiles(ctx, bestCheckpoint)
`)
	fmt.Println()

	fmt.Println("3. Undo Last Operation")
	fmt.Print(`
  // Maintain a stack of checkpoints for undo
  type CheckpointStack struct {
      uuids []string
  }

  func (s *CheckpointStack) Push(uuid string) {
      s.uuids = append(s.uuids, uuid)
  }

  func (s *CheckpointStack) Pop() string {
      if len(s.uuids) == 0 {
          return ""
      }
      uuid := s.uuids[len(s.uuids)-1]
      s.uuids = s.uuids[:len(s.uuids)-1]
      return uuid
  }

  // Undo last change
  if uuid := stack.Pop(); uuid != "" {
      client.RewindFiles(ctx, uuid)
  }
`)
	fmt.Println()
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

func stringPtr(s string) *string {
	return &s
}
