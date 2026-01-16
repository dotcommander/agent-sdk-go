# agent-sdk-go

Go wrapper for the Claude CLI, enabling programmatic control of Claude Code sessions from Go applications.

## What This Is

This is **not** an HTTP client for the Anthropic API. It's a Go library that wraps the [Claude CLI](https://docs.anthropic.com/en/docs/claude-code) (`claude` command), communicating via subprocess stdin/stdout with JSON streaming.

```
┌─────────────┐      stdin/stdout       ┌────────────┐
│   Your Go   │ ◄──── stream-json ────► │ Claude CLI │
│   Program   │                         │ (claude)   │
└─────────────┘                         └────────────┘
```

## Use Cases

- Embed Claude Code capabilities in Go tools
- Build automation around Claude CLI
- Create custom agent loops (like [looper](https://github.com/dotcommander/looper))
- Programmatic multi-turn conversations

## Installation

```bash
# Requires Claude CLI installed
# https://docs.anthropic.com/en/docs/claude-code

go get github.com/dotcommander/agent-sdk-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/dotcommander/agent-sdk-go/internal/claude/subprocess"
)

func main() {
    // One-shot mode: prompt passed as CLI argument
    transport, _ := subprocess.NewTransportWithPrompt(
        &subprocess.TransportConfig{
            Model: "claude-sonnet-4-5-20250929",
        },
        "What is 2+2?",
    )

    ctx := context.Background()
    transport.Connect(ctx)
    defer transport.Close()

    msgs, errs := transport.ReceiveMessages(ctx)
    for msg := range msgs {
        fmt.Printf("%T: %+v\n", msg, msg)
    }
    for err := range errs {
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Modes

### One-Shot Mode

For single prompts - prompt passed as CLI argument:

```go
transport, _ := subprocess.NewTransportWithPrompt(config, "Your prompt here")
```

Equivalent to: `claude -p --output-format stream-json "Your prompt here"`

### Interactive Mode

For multi-turn conversations via stdin/stdout:

```go
transport, _ := subprocess.NewTransport(config)
transport.Connect(ctx)
transport.SendMessage(ctx, "Hello")
// ... receive response ...
transport.SendMessage(ctx, "Follow-up question")
```

## Project Structure

```
agent-sdk-go/
├── cmd/
│   └── demo/           # Demo CLI showing SDK usage
├── internal/
│   └── claude/
│       ├── subprocess/ # Core transport (stdin/stdout communication)
│       ├── parser/     # JSON message parsing
│       ├── cli/        # CLI discovery
│       ├── shared/     # Shared types and utilities
│       └── v2/         # V2-style session API
```

## Configuration

```go
config := &subprocess.TransportConfig{
    CLIPath:      "",                           // Auto-discover
    Model:        "claude-sonnet-4-5-20250929", // Model to use
    SystemPrompt: "You are a helpful assistant",
    Timeout:      60 * time.Second,
    Env:          map[string]string{},          // Extra env vars
}
```

## Demo CLI

```bash
# Build
go build -o agent-demo ./cmd/demo

# Commands
./agent-demo check              # Verify Claude CLI is available
./agent-demo prompt "question"  # One-shot prompt
./agent-demo chat               # Interactive chat
./agent-demo stream "question"  # Stream response tokens
```

## Requirements

- Go 1.21+
- Claude CLI installed and authenticated
- macOS, Linux, or Windows with Claude CLI support

## Related

- [looper](https://github.com/dotcommander/looper) - Ralph Wiggum loop built on this SDK
- [Claude CLI Docs](https://docs.anthropic.com/en/docs/claude-code)

## License

MIT
