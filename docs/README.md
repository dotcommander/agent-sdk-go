# Claude Code Agent SDK Documentation

Welcome to the Claude Code Agent SDK for Go. This SDK enables you to build powerful AI agents by programmatically controlling Claude Code CLI from your Go applications.

## Documentation

| Guide | Description |
|-------|-------------|
| [Getting Started](getting-started.md) | Installation, prerequisites, and your first agent |
| [Core Concepts](concepts.md) | Architecture, message flow, and design principles |
| [Client API](client.md) | Complete client reference with examples |
| [Streaming](streaming.md) | Real-time response handling |
| [Sessions](sessions.md) | Multi-turn conversations and state management |
| [Configuration](configuration.md) | All available options and settings |
| [Advanced Usage](advanced.md) | Low-level transport, custom message handling |
| [Troubleshooting](troubleshooting.md) | Common issues and solutions |

## Quick Example

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/dotcommander/agent-sdk-go/claude"
)

func main() {
    // Create a client
    client, err := claude.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    // Send a query
    response, err := client.Query(context.Background(), "Explain Go interfaces in one paragraph")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response)
}
```

## How It Works

```
┌──────────────────┐                    ┌──────────────────┐
│                  │   stdin (JSON)     │                  │
│   Your Go App    │ ─────────────────► │   Claude Code    │
│                  │                    │      CLI         │
│   claude.Client  │ ◄───────────────── │                  │
│                  │   stdout (JSON)    │                  │
└──────────────────┘                    └──────────────────┘
```

The SDK spawns Claude Code CLI as a subprocess and communicates via JSON streaming over stdin/stdout. This means:

- **No API keys needed** - Uses your authenticated Claude Code CLI
- **Full Claude Code features** - Tools, MCP servers, file access
- **Real-time streaming** - Process responses as they arrive

## Requirements

- Go 1.21 or later
- Claude Code CLI installed and authenticated
- macOS, Linux, or Windows

## Getting Help

- [GitHub Issues](https://github.com/dotcommander/agent-sdk-go/issues)
- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code)
