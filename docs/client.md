# Client API

The `Client` interface is the primary way to interact with Claude Code.

## Creating a Client

### Basic Creation

```go
import "github.com/dotcommander/agent-sdk-go/claude"

client, err := claude.NewClient()
if err != nil {
    log.Fatal(err)
}
```

### With Options

```go
client, err := claude.NewClient(
    claude.WithModel("claude-sonnet-4-20250514"),
    claude.WithTimeout("60s"),
    claude.WithSystemPrompt("You are a helpful coding assistant"),
)
```

See [Configuration](configuration.md) for all available options.

## Client Interface

```go
type Client interface {
    // One-shot queries
    Query(ctx context.Context, prompt string) (string, error)
    QueryStream(ctx context.Context, prompt string) (<-chan Message, <-chan error)

    // Session management
    Connect(ctx context.Context) error
    Disconnect() error
    ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)

    // Configuration
    SetModel(model string)
    SetSessionID(sessionID string)
    GetSessionID() string

    // Control
    Interrupt() error
}
```

## One-Shot Queries

### Query

Sends a prompt and returns the complete response as a string.

```go
response, err := client.Query(ctx, "What is 2 + 2?")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response) // "4"
```

**Parameters:**
- `ctx` - Context for cancellation and timeout
- `prompt` - The prompt to send to Claude

**Returns:**
- `string` - The complete text response
- `error` - Any error that occurred

**Behavior:**
- Creates a new subprocess for each call
- Waits for the complete response
- Automatically handles message parsing
- Extracts text content from response

### QueryStream

Sends a prompt and returns channels for streaming responses.

```go
msgChan, errChan := client.QueryStream(ctx, "Tell me a story")

for {
    select {
    case msg, ok := <-msgChan:
        if !ok {
            return // Done
        }
        processMessage(msg)
    case err := <-errChan:
        if err != nil {
            log.Fatal(err)
        }
        return
    }
}
```

**Parameters:**
- `ctx` - Context for cancellation
- `prompt` - The prompt to send

**Returns:**
- `<-chan Message` - Channel of messages (closed when done)
- `<-chan error` - Channel of errors

**Behavior:**
- Creates a new subprocess
- Streams messages as they arrive
- Message channel closes when response is complete
- Error channel receives any errors

## Processing Messages

### Extracting Text

```go
import "github.com/dotcommander/agent-sdk-go/claude"

msgChan, errChan := client.QueryStream(ctx, prompt)

for msg := range msgChan {
    text := claude.GetContentText(msg)
    if text != "" {
        fmt.Print(text)
    }
}
```

### Type Switching

```go
for msg := range msgChan {
    switch m := msg.(type) {
    case *claude.AssistantMessage:
        fmt.Println("Assistant:", m.Content)
    case *claude.ResultMessage:
        fmt.Println("Done! Cost:", m.Usage)
    case *claude.ErrorMessage:
        fmt.Println("Error:", m.Error)
    }
}
```

## Session Management

For multi-turn conversations, use session methods.

### Connect

Establishes a persistent connection to Claude CLI.

```go
err := client.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Disconnect()
```

### ReceiveMessages

Receives messages from an active session.

```go
// After Connect()
msgChan, errChan := client.ReceiveMessages(ctx)
```

### Disconnect

Closes the session.

```go
err := client.Disconnect()
```

See [Sessions](sessions.md) for detailed session management.

## Configuration Methods

### SetModel

Changes the model for subsequent queries.

```go
client.SetModel("claude-opus-4-5-20251101")
```

### SetSessionID

Sets a session ID for resuming conversations.

```go
client.SetSessionID("session-abc-123")
```

### GetSessionID

Returns the current session ID.

```go
id := client.GetSessionID()
```

## Control Methods

### Interrupt

Forcibly interrupts the current operation.

```go
// In another goroutine
go func() {
    time.Sleep(5 * time.Second)
    client.Interrupt()
}()
```

## Error Handling

### Common Errors

```go
response, err := client.Query(ctx, prompt)
if err != nil {
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        log.Println("Request timed out")
    case errors.Is(err, context.Canceled):
        log.Println("Request was cancelled")
    default:
        log.Printf("Error: %v", err)
    }
}
```

### Retrying Failed Requests

```go
func queryWithRetry(client claude.Client, ctx context.Context, prompt string, maxRetries int) (string, error) {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        response, err := client.Query(ctx, prompt)
        if err == nil {
            return response, nil
        }
        lastErr = err
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

## Complete Example

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/dotcommander/agent-sdk-go/claude"
)

func main() {
    // Create client with configuration
    client, err := claude.NewClient(
        claude.WithModel("claude-sonnet-4-20250514"),
        claude.WithTimeout("60s"),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Stream a response
    fmt.Println("Asking Claude...")
    msgChan, errChan := client.QueryStream(ctx, "Explain goroutines in Go")

    for {
        select {
        case msg, ok := <-msgChan:
            if !ok {
                fmt.Println("\nDone!")
                return
            }
            if text := claude.GetContentText(msg); text != "" {
                fmt.Print(text)
            }
        case err := <-errChan:
            if err != nil {
                log.Fatal(err)
            }
            return
        case <-ctx.Done():
            log.Fatal("Timeout")
        }
    }
}
```

## Next Steps

- [Streaming](streaming.md) - Advanced streaming patterns
- [Sessions](sessions.md) - Multi-turn conversations
- [Configuration](configuration.md) - All available options
