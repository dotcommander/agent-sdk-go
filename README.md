# agent-sdk-go

A Go implementation of the Anthropic Claude Agent SDK, ported from TypeScript to provide native Go interfaces for building Claude-powered agents and tools.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [SDK Reference](#sdk-reference)
  - [Core Client](#core-client)
  - [Tool System](#tool-system)
  - [Streaming](#streaming)
  - [Session Management (v2-style)](#session-management-v2-style)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Development](#development)
  - [Porting Status](#porting-status)
  - [Dependencies](#dependencies)
- [Configuration](#configuration)
- [License](#license)

## Features

- **Native Go SDK**: No Node.js runtime required
- **Full API Coverage**: Messages, tools, streaming, error handling
- **Type Safety**: Go structs matching Anthropic API contracts
- **Concurrency Ready**: Leverages goroutines and channels
- **CLI Integration**: Built-in commands for agent interaction

## Installation

### From Source

```bash
# Build simple demo (from cmd/main.go)
go build -o agent-sdk-go ./cmd

# Build agent CLI with subcommands (from cmd/agent/main.go)
go build -o agent-sdk-go-agent ./cmd/agent

# Install both
go install ./cmd                          # Simple demo
go install ./cmd/agent                    # Agent CLI with subcommands

# Or create symlinks
ln -sf "$(pwd)/agent-sdk-go" ~/go/bin/
ln -sf "$(pwd)/agent-sdk-go-agent" ~/go/bin/

# Release build with version info
go build -ldflags "-X main.version=$(git describe --tags 2>/dev/null || echo dev)" -o agent-sdk-go ./cmd
go build -ldflags "-X main.version=$(git describe --tags 2>/dev/null || echo dev)" -o agent-sdk-go-agent ./cmd/agent
```

### As Library

```bash
go get agent-sdk-go/sdk
```

## Quick Start

### 1. Set API Key

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

### 2. Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "os"
    "agent-sdk-go/sdk"
)

func main() {
    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        fmt.Fprintln(os.Stderr, "ANTHROPIC_API_KEY environment variable not set")
        os.Exit(1)
    }

    client, err := sdk.NewClient(apiKey)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to create client: %v\n", err)
        os.Exit(1)
    }

    resp, err := client.SendMessage(context.Background(), sdk.MessageRequest{
        Model:     "claude-3-5-sonnet-20241022",
        MaxTokens: 100,
        Messages: []sdk.Message{{Role: "user", Content: "Hello, Claude!"}},
    })
    if err != nil {
        fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
        os.Exit(1)
    }

    for _, block := range resp.Content {
        if block.Type == "text" {
            fmt.Println(block.Text)
        }
    }
}
```

### 3. CLI Usage

The project includes two CLI binaries:

1. **Simple demo** (`agent-sdk-go`) - Basic message sending example
2. **Agent CLI** (`agent-sdk-go-agent`) - Full-featured CLI with subcommands

#### Simple Demo (agent-sdk-go)
```bash
# Build and run simple demo
go build -o agent-sdk-go ./cmd
./agent-sdk-go

# Output: Claude's response to "Hello, Claude!"
```

#### Agent CLI with Subcommands (agent-sdk-go-agent)
```bash
# Build agent CLI
go build -o agent-sdk-go-agent ./cmd/agent

# Interactive chat
agent-sdk-go-agent run

# Stream responses
agent-sdk-go-agent stream

# Test tool execution
agent-sdk-go-agent tool

# Show help
agent-sdk-go-agent help
```

## SDK Reference

### Core Client

```go
client, err := sdk.NewClient(apiKey,
    sdk.WithBaseURL("https://api.custom.com"),
    sdk.WithTimeout(30*time.Second),
)
```

### Tool System

```go
// Define tool
calculator := sdk.Tool{
    Name:        "calculator",
    Description: "Performs arithmetic",
    InputSchema: sdk.InputSchema{
        Type: "object",
        Properties: map[string]sdk.Property{
            "operation": {Type: "string", Enum: []string{"add", "subtract", "multiply", "divide"}},
            "a": {Type: "number"},
            "b": {Type: "number"},
        },
        Required: []string{"operation", "a", "b"},
    },
}

// Implement ToolExecutor interface
type calculatorExecutor struct{}

func (e *calculatorExecutor) Execute(ctx context.Context, toolName string, args map[string]any) (any, error) {
    operation, _ := args["operation"].(string)
    a, _ := args["a"].(float64)
    b, _ := args["b"].(float64)

    switch operation {
    case "add":
        return a + b, nil
    case "subtract":
        return a - b, nil
    case "multiply":
        return a * b, nil
    case "divide":
        if b == 0 {
            return nil, fmt.Errorf("division by zero")
        }
        return a / b, nil
    default:
        return nil, fmt.Errorf("unknown operation: %s", operation)
    }
}

// Register executor
err := client.RegisterTool(calculator, &calculatorExecutor{})
if err != nil {
    fmt.Fprintf(os.Stderr, "Failed to register tool: %v\n", err)
    return
}

// Tool responses are automatically executed
resp, err := client.SendMessage(ctx, sdk.MessageRequest{
    Model: "claude-3-5-sonnet-20241022",
    Messages: []sdk.Message{{Role: "user", Content: "What's 2 + 2?"}},
    Tools: []sdk.Tool{calculator},
})
if err != nil {
    fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
    return
}

// Process response with tool results
for _, block := range resp.Content {
    if block.Type == "text" {
        fmt.Println(block.Text)
    }
}
```

### Streaming

```go
// Create a request
req := sdk.MessageRequest{
    Model:     "claude-3-5-sonnet-20241022",
    MaxTokens: 1024,
    Messages: []sdk.Message{{Role: "user", Content: "Tell me a story about robots."}},
}

// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Start streaming
events, err := client.StreamMessage(ctx, req)
if err != nil {
    // Handle initial error (network, authentication, etc.)
    fmt.Fprintf(os.Stderr, "Failed to start stream: %v\n", err)
    return
}

// Process stream events
var responseText strings.Builder
for event := range events {
    if event.Error != nil {
        // Handle stream error (JSON parsing, etc.)
        fmt.Fprintf(os.Stderr, "Stream error: %v\n", event.Error)
        continue
    }

    if event.Delta.Text != "" {
        // Print real-time output
        fmt.Print(event.Delta.Text)
        responseText.WriteString(event.Delta.Text)
    }
}
fmt.Println() // New line after stream

// Use accumulated response if needed
fmt.Printf("Full response (%d chars):\n%s\n",
    responseText.Len(), responseText.String())
```


### Session Management (v2-style)

```go
// Create context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Create a new session with options
session := client.CreateSession(
    sdk.WithSessionModel("claude-3-5-sonnet-20241022"),
    sdk.WithSessionMaxTokens(1024),
    sdk.WithSessionSystem("You are a helpful assistant"),
    sdk.WithSessionTemperature(0.7),
)

// Send initial message
session.Send("Hello, Claude!") // Send() doesn't return errors

// Option 1: Get streaming response
events, err := session.Stream(ctx)
if err != nil {
    fmt.Fprintf(os.Stderr, "Failed to start stream: %v\n", err)
    return
}

var responseText strings.Builder
for event := range events {
    if event.Error != nil {
        fmt.Fprintf(os.Stderr, "Stream error: %v\n", event.Error)
        continue
    }
    if event.Delta.Text != "" {
        fmt.Print(event.Delta.Text)
        responseText.WriteString(event.Delta.Text)
    }
}
fmt.Println()

// Option 2: Get complete response (non-streaming)
resp, err := session.SendMessage(ctx)
if err != nil {
    fmt.Fprintf(os.Stderr, "Failed to send message: %v\n", err)
    return
}

for _, block := range resp.Content {
    if block.Type == "text" {
        fmt.Println(block.Text)
    }
}

// Session ID is automatically captured from responses
fmt.Printf("Session ID: %s\n", session.ID())

// Multi-turn conversation
session.Send("What's 2 + 2?")
resp2, err := session.SendMessage(ctx)
if err != nil {
    fmt.Fprintf(os.Stderr, "Failed second message: %v\n", err)
    return
}
// Process response (similar to above)...

// Resume a session later (in a different part of your application)
resumedSession := client.ResumeSession(session.ID())
resumedSession.Send("Continue our conversation...")
// Session maintains conversation history automatically
```

## Project Structure

```
agent-sdk-go/
├── cmd/                    # CLI entry points
│   ├── main.go            # Simple demo (basic message sending)
│   └── agent/             # Agent CLI with subcommands
│       └── main.go        # Subcommands: run, tool, stream
├── internal/
│   ├── sdk/               # Go SDK implementation
│   │   ├── client.go      # Core client
│   │   ├── types.go       # Type definitions
│   │   ├── tools.go       # Tool system
│   │   ├── stream.go      # SSE streaming
│   │   ├── session.go     # Session management (v2-style)
│   │   ├── errors.go      # Error handling
│   │   ├── client_test.go # Core tests
│   │   └── session_test.go # Session tests
│   ├── commands/          # CLI commands (Cobra)
│   └── actions/           # Business logic
└── SPEC-ANTHROPIC-SDK-PORT.md  # Implementation specification
```

## Testing

```bash
# Run SDK tests
go test ./internal/sdk/...

# With coverage
go test ./internal/sdk/... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Development

### Porting Status

The SDK implements the complete TypeScript API surface:

- ✅ Client initialization and configuration
- ✅ Message sending and receiving
- ✅ Tool definition and execution
- ✅ Streaming responses (SSE)
- ✅ Error handling and types
- ✅ CLI integration
- ✅ Session management (v2-style)
- ✅ Comprehensive test coverage (94.1%)

### Dependencies

- Standard library: `net/http`, `encoding/json`, `context`
- External: None for core functionality
- Hand-rolled SSE parser (no external dependencies)

## Configuration

Configuration is loaded from multiple sources with the following precedence (highest to lowest):

1. **Environment variable**: `ANTHROPIC_API_KEY`
2. **Local config file**: `./config.yaml` (current directory)
3. **User config file**: `~/.config/agent-sdk-go/config.yaml`
4. **System config file**: `/etc/agent-sdk-go/config.yaml`

### Configuration File Format

Create a YAML configuration file in one of the locations above:

```yaml
# Required: Your Anthropic API key
api_key: "sk-ant-..."

# Optional: Custom base URL (default: https://api.anthropic.com/v1)
base_url: "https://api.anthropic.com/v1"

# Optional: HTTP timeout in seconds (default: 60)
timeout: 30

# Optional: Default model to use
model: "claude-3-5-sonnet-20241022"
```

### Using Configuration in Code

```go
import "agent-sdk-go/internal/app"

// Load configuration
cfg, err := app.Load()
if err != nil {
    // Handle config loading error
}

// Create client from configuration
client, err := app.NewClientFromConfig(cfg)
if err != nil {
    // Handle client creation error
}

// Get model from config with fallback
model := cfg.GetModel("claude-3-5-sonnet-20241022")

// Convenience function that loads config and creates client
client2, err := app.NewClient()
if err != nil {
    // Handles both config loading and client creation
}
```

### CLI Usage

Both CLI binaries automatically load configuration using the precedence order above. Set your API key via environment variable or config file:

```bash
# Using environment variable
export ANTHROPIC_API_KEY="sk-ant-..."

# Simple demo
agent-sdk-go

# Agent CLI with subcommands
agent-sdk-go-agent run
agent-sdk-go-agent stream
agent-sdk-go-agent tool

# Using config file
echo 'api_key: "sk-ant-..."' > ~/.config/agent-sdk-go/config.yaml
agent-sdk-go
agent-sdk-go-agent run
```

## License

See LICENSE file for details.
