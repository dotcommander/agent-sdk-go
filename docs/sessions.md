# Sessions

Sessions enable multi-turn conversations where Claude remembers previous context.

## One-Shot vs Sessions

| Aspect | One-Shot (`Query`) | Session (`Connect`) |
|--------|-------------------|---------------------|
| Process | New process per query | Persistent process |
| Context | No memory | Remembers conversation |
| Use case | Independent questions | Conversations, agents |
| Overhead | Higher (process spawn) | Lower (reuse process) |

## Creating a Session

```go
client, err := claude.NewClient()
if err != nil {
    log.Fatal(err)
}

// Establish session
err = client.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Disconnect()
```

## Multi-Turn Conversation

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "log"
    "os"

    "github.com/dotcommander/agent-sdk-go/claude"
    "github.com/dotcommander/agent-sdk-go/claude/shared"
)

func main() {
    client, _ := claude.NewClient()
    ctx := context.Background()

    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect()

    scanner := bufio.NewScanner(os.Stdin)

    for {
        fmt.Print("You: ")
        if !scanner.Scan() {
            break
        }
        input := scanner.Text()
        if input == "exit" {
            break
        }

        // Send message and receive response
        msgChan, errChan := client.ReceiveMessages(ctx)
        // Note: You'd need to send the message via the transport

        fmt.Print("Claude: ")
        for {
            select {
            case msg, ok := <-msgChan:
                if !ok {
                    fmt.Println()
                    goto nextTurn
                }
                fmt.Print(shared.GetContentText(msg))
            case err := <-errChan:
                if err != nil {
                    log.Printf("Error: %v\n", err)
                }
                goto nextTurn
            }
        }
    nextTurn:
    }
}
```

## Session IDs

Session IDs allow resuming conversations:

```go
// Set a session ID
client.SetSessionID("conversation-123")

// Get current session ID
id := client.GetSessionID()

// Resume a previous session
client.SetSessionID(previousSessionID)
err := client.Connect(ctx)
```

## V2 Session API

For more control, use the V2 session API:

```go
import "github.com/dotcommander/agent-sdk-go/claude/v2"

// Create a session
session, err := v2.NewSession(ctx,
    v2.WithModel("claude-sonnet-4-20250514"),
    v2.WithSystemPrompt("You are a coding assistant"),
)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// Send messages
response, err := session.Send(ctx, "Hello!")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response)

// Continue conversation
response, err = session.Send(ctx, "Can you elaborate?")
```

## Session Options

Configure session behavior:

```go
session, err := v2.NewSession(ctx,
    v2.WithModel("claude-sonnet-4-20250514"),
    v2.WithSystemPrompt("You are helpful"),
    v2.WithTimeout(60*time.Second),
    v2.WithResume("previous-session-id"),
)
```

## Handling Disconnections

```go
func runSession(ctx context.Context) error {
    client, _ := claude.NewClient()

    for {
        err := client.Connect(ctx)
        if err != nil {
            log.Printf("Connection failed: %v, retrying...", err)
            time.Sleep(time.Second)
            continue
        }

        // Session loop
        for {
            msgChan, errChan := client.ReceiveMessages(ctx)

            select {
            case msg, ok := <-msgChan:
                if !ok {
                    // Session ended normally
                    break
                }
                processMessage(msg)
            case err := <-errChan:
                if err != nil {
                    log.Printf("Session error: %v", err)
                    client.Disconnect()
                    break // Reconnect
                }
            case <-ctx.Done():
                client.Disconnect()
                return ctx.Err()
            }
        }
    }
}
```

## Context Management

Add context files to the session:

```go
// Add files to context
err := client.RewindFiles(ctx, []string{
    "main.go",
    "utils.go",
})

// Now queries can reference these files
response, _ := client.Query(ctx, "What does the main function do?")
```

## Session State

Track session state in your application:

```go
type SessionState struct {
    ID        string
    StartTime time.Time
    Messages  int
    LastError error
}

func (s *SessionState) RecordMessage() {
    s.Messages++
}

func (s *SessionState) RecordError(err error) {
    s.LastError = err
}
```

## Graceful Shutdown

Handle shutdown gracefully:

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())

    // Handle interrupt
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigChan
        fmt.Println("\nShutting down...")
        cancel()
    }()

    client, _ := claude.NewClient()
    if err := client.Connect(ctx); err != nil {
        log.Fatal(err)
    }

    // Session loop
    for {
        select {
        case <-ctx.Done():
            client.Disconnect()
            return
        default:
            // Process messages
        }
    }
}
```

## Best Practices

### 1. Always Disconnect

```go
client.Connect(ctx)
defer client.Disconnect() // Always clean up
```

### 2. Handle Context Cancellation

```go
select {
case <-ctx.Done():
    client.Disconnect()
    return ctx.Err()
case msg := <-msgChan:
    process(msg)
}
```

### 3. Use Timeouts

```go
ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
defer cancel()

client.Connect(ctx)
```

### 4. Persist Session IDs

```go
// Save session ID for later
sessionID := client.GetSessionID()
saveToDatabase(sessionID)

// Later, resume
savedID := loadFromDatabase()
client.SetSessionID(savedID)
client.Connect(ctx)
```

## Next Steps

- [Configuration](configuration.md) - All session options
- [Advanced Usage](advanced.md) - Custom transport
- [Troubleshooting](troubleshooting.md) - Session issues
