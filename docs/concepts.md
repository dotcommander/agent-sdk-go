# Core Concepts

Understanding how the SDK works will help you build more effective agents.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        Your Application                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────────────┐  │
│  │   Client    │───►│  Transport  │───►│  Claude Code CLI    │  │
│  │             │◄───│             │◄───│                     │  │
│  └─────────────┘    └─────────────┘    └─────────────────────┘  │
│        │                  │                      │               │
│        ▼                  ▼                      ▼               │
│   High-level API    Subprocess I/O         Anthropic API        │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Components

| Component | Package | Purpose |
|-----------|---------|---------|
| Client | `claude` | High-level API for queries and sessions |
| Transport | `claude/subprocess` | Manages CLI subprocess communication |
| Parser | `claude/parser` | Parses JSON messages from CLI |
| Shared | `claude` | Common types and utilities (re-exported via `claude` package) |

## Message Flow

### One-Shot Query

```
Your App                    Transport                   Claude CLI
   │                           │                            │
   │  Query("Hello")           │                            │
   │──────────────────────────►│                            │
   │                           │  spawn process             │
   │                           │───────────────────────────►│
   │                           │  stdin: prompt             │
   │                           │───────────────────────────►│
   │                           │                            │
   │                           │  stdout: JSON messages     │
   │                           │◄───────────────────────────│
   │                           │                            │
   │  response string          │  process exits             │
   │◄──────────────────────────│◄───────────────────────────│
   │                           │                            │
```

### Streaming Query

```
Your App                    Transport                   Claude CLI
   │                           │                            │
   │  QueryStream("Hello")     │                            │
   │──────────────────────────►│                            │
   │                           │  spawn + connect           │
   │  msgChan, errChan         │───────────────────────────►│
   │◄──────────────────────────│                            │
   │                           │                            │
   │  <-msgChan (message 1)    │  stdout: {"type":"..."}   │
   │◄──────────────────────────│◄───────────────────────────│
   │                           │                            │
   │  <-msgChan (message 2)    │  stdout: {"type":"..."}   │
   │◄──────────────────────────│◄───────────────────────────│
   │                           │                            │
   │  msgChan closed           │  process exits             │
   │◄──────────────────────────│◄───────────────────────────│
```

## Message Types

The SDK handles various message types from Claude Code CLI:

### Content Messages

| Type | Description |
|------|-------------|
| `AssistantMessage` | Claude's response with content blocks |
| `UserMessage` | Echo of user input |
| `SystemMessage` | System-level information |

### Control Messages

| Type | Description |
|------|-------------|
| `ResultMessage` | Final result with metadata |
| `ErrorMessage` | Error information |
| `ToolUseMessage` | Tool invocation request |
| `ToolResultMessage` | Tool execution result |

### System Messages

| Type | Description |
|------|-------------|
| `InitMessage` | Session initialization |
| `PingMessage` | Keep-alive ping |
| `ConfigMessage` | Configuration updates |

## Content Blocks

Messages contain content blocks of various types:

```go
// Text content
type TextBlock struct {
    Type string `json:"type"` // "text"
    Text string `json:"text"`
}

// Tool use request
type ToolUseBlock struct {
    Type   string         `json:"type"` // "tool_use"
    ID     string         `json:"id"`
    Name   string         `json:"name"`
    Input  map[string]any `json:"input"`
}

// Tool result
type ToolResultBlock struct {
    Type      string `json:"type"` // "tool_result"
    ToolUseID string `json:"tool_use_id"`
    Content   string `json:"content"`
    IsError   bool   `json:"is_error"`
}
```

## Extracting Content

Use helper functions to extract content from messages:

```go
import "github.com/dotcommander/agent-sdk-go/claude"

// Get text content from any message
text := claude.GetContentText(msg)

// Check message type
switch m := msg.(type) {
case *claude.AssistantMessage:
    for _, block := range m.Content {
        // Process each content block
    }
case *claude.ResultMessage:
    // Handle completion
}
```

## Contexts and Cancellation

All SDK operations accept a `context.Context` for cancellation:

```go
// Create a context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// The query will be cancelled if it exceeds 30 seconds
response, err := client.Query(ctx, "Complex question...")
if err == context.DeadlineExceeded {
    log.Println("Query timed out")
}
```

## Concurrency

The client is safe for concurrent use from multiple goroutines. Each `Query` or `QueryStream` call creates its own subprocess.

```go
var wg sync.WaitGroup
client, _ := claude.NewClient()

for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        resp, _ := client.Query(ctx, fmt.Sprintf("Question %d", id))
        fmt.Printf("Response %d: %s\n", id, resp)
    }(i)
}

wg.Wait()
```

## Error Handling

The SDK returns descriptive errors:

```go
response, err := client.Query(ctx, prompt)
if err != nil {
    switch {
    case errors.Is(err, context.DeadlineExceeded):
        // Timeout
    case errors.Is(err, context.Canceled):
        // Cancelled
    default:
        // Check for specific error types
        var cliErr *claude.CLIError
        if errors.As(err, &cliErr) {
            log.Printf("CLI error: %s (exit code: %d)", cliErr.Message, cliErr.ExitCode)
        }
    }
}
```

## Next Steps

- [Client API](client.md) - Detailed API reference
- [Streaming](streaming.md) - Real-time response handling
- [Sessions](sessions.md) - Multi-turn conversations
