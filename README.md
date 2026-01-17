# Claude Code Agent SDK for Go

Go SDK for building agents with Claude Code CLI. Provides programmatic control over Claude Code sessions via subprocess communication.

## What This Is

This SDK wraps the [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code), enabling Go applications to:
- Run Claude Code programmatically
- Build custom agent loops
- Stream responses in real-time
- Execute multi-turn conversations

```
┌─────────────┐      stdin/stdout       ┌────────────┐
│   Your Go   │ ◄──── stream-json ────► │ Claude CLI │
│   Program   │                         │  (claude)  │
└─────────────┘                         └────────────┘
```

**No API key required** - uses your authenticated Claude Code CLI.

## Installation

```bash
# Requires Claude Code CLI installed and authenticated
# https://docs.anthropic.com/en/docs/claude-code

go get github.com/dotcommander/agent-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/dotcommander/agent-sdk-go/claude"
)

func main() {
    client, err := claude.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    response, err := client.Query(ctx, "What is 2+2?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response)
}
```

## Streaming Responses

```go
client, _ := claude.NewClient()

msgChan, errChan := client.QueryStream(ctx, "Tell me a story")

for {
    select {
    case msg, ok := <-msgChan:
        if !ok {
            return // Done
        }
        // Process message
        fmt.Printf("%T: %+v\n", msg, msg)
    case err := <-errChan:
        log.Fatal(err)
    }
}
```

## Interactive Sessions

```go
client, _ := claude.NewClient()

// Connect for multi-turn conversation
if err := client.Connect(ctx); err != nil {
    log.Fatal(err)
}
defer client.Disconnect()

// Send messages and receive responses
msgChan, errChan := client.ReceiveMessages(ctx)
// ... handle messages
```

## Configuration

```go
client, err := claude.NewClient(
    claude.WithModel("claude-sonnet-4-20250514"),
    claude.WithTimeout("60s"),
    claude.WithSystemPrompt("You are a helpful assistant"),
)
```

## Using Subprocess Transport Directly

For lower-level control:

```go
import "github.com/dotcommander/agent-sdk-go/claude/subprocess"

config := &subprocess.TransportConfig{
    Model:   "claude-sonnet-4-20250514",
    Timeout: 60 * time.Second,
}

// One-shot mode
transport, _ := subprocess.NewTransportWithPrompt(config, "Hello!")
transport.Connect(ctx)
defer transport.Close()

msgs, errs := transport.ReceiveMessages(ctx)
for msg := range msgs {
    fmt.Printf("%+v\n", msg)
}
```

## Documentation

| Guide | Description |
|-------|-------------|
| [Getting Started](docs/getting-started.md) | Installation and first steps |
| [Core Concepts](docs/concepts.md) | Architecture and design |
| [Client API](docs/client.md) | Complete API reference |
| [Streaming](docs/streaming.md) | Real-time response handling |
| [Sessions](docs/sessions.md) | Multi-turn conversations |
| [Configuration](docs/configuration.md) | All options and settings |
| [Advanced Usage](docs/advanced.md) | Low-level transport, custom handlers |
| [Troubleshooting](docs/troubleshooting.md) | Common issues and solutions |

## Project Structure

```
agent-sdk-go/
├── claude/                 # Public SDK package
│   ├── client.go           # High-level client
│   ├── subprocess/         # CLI subprocess transport
│   ├── parser/             # JSON message parsing
│   ├── shared/             # Shared types
│   ├── v2/                 # V2 session API
│   ├── cli/                # CLI discovery
│   └── mcp/                # MCP server support
├── docs/                   # Documentation
└── examples/               # Usage examples
```

## Requirements

- Go 1.21+
- Claude Code CLI installed and authenticated
- macOS, Linux, or Windows

## License

MIT
