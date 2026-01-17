# Getting Started

This guide will help you install the SDK and build your first Claude Code agent.

## Prerequisites

### 1. Install Go

The SDK requires Go 1.21 or later. Download from [go.dev](https://go.dev/dl/) or use your package manager:

```bash
# macOS
brew install go

# Ubuntu/Debian
sudo apt install golang-go

# Verify installation
go version
```

### 2. Install Claude Code CLI

The SDK communicates with Claude through the Claude Code CLI. Install it following the [official instructions](https://docs.anthropic.com/en/docs/claude-code).

```bash
# Verify Claude Code is installed and authenticated
claude --version
claude "Hello, world"
```

If Claude responds, you're ready to proceed.

## Installation

Add the SDK to your Go project:

```bash
go get github.com/dotcommander/agent-sdk-go
```

## Your First Agent

Create a new file `main.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/dotcommander/agent-sdk-go/claude"
)

func main() {
    // Create a client with default settings
    client, err := claude.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    // Create a context (can be used for cancellation)
    ctx := context.Background()

    // Send a query and get the response
    response, err := client.Query(ctx, "What is the capital of France?")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(response)
}
```

Run your agent:

```bash
go run main.go
```

You should see Claude's response printed to the console.

## Understanding the Response

The `Query` method returns the complete text response as a string. For more control over the response, you can use streaming.

## Streaming Responses

To process responses as they arrive:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/dotcommander/agent-sdk-go/claude"
    "github.com/dotcommander/agent-sdk-go/claude/shared"
)

func main() {
    client, err := claude.NewClient()
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()
    msgChan, errChan := client.QueryStream(ctx, "Write a haiku about Go programming")

    for {
        select {
        case msg, ok := <-msgChan:
            if !ok {
                fmt.Println() // Final newline
                return
            }
            // Extract and print text content
            if text := shared.GetContentText(msg); text != "" {
                fmt.Print(text)
            }
        case err := <-errChan:
            if err != nil {
                log.Fatal(err)
            }
            return
        }
    }
}
```

## Configuring the Client

Customize the client with options:

```go
client, err := claude.NewClient(
    claude.WithModel("claude-sonnet-4-20250514"),
    claude.WithTimeout("120s"),
)
```

See [Configuration](configuration.md) for all available options.

## Next Steps

- [Core Concepts](concepts.md) - Understand how the SDK works
- [Client API](client.md) - Complete API reference
- [Streaming](streaming.md) - Advanced streaming patterns
- [Sessions](sessions.md) - Multi-turn conversations

## Common Issues

### "claude: command not found"

Claude Code CLI is not installed or not in your PATH. Install it following the [official instructions](https://docs.anthropic.com/en/docs/claude-code).

### "not authenticated"

Run `claude` manually and complete the authentication flow.

### Context deadline exceeded

The default timeout may be too short for complex queries. Increase it:

```go
client, err := claude.NewClient(
    claude.WithTimeout("300s"), // 5 minutes
)
```

See [Troubleshooting](troubleshooting.md) for more solutions.
