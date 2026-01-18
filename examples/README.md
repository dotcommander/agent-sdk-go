# agent-sdk-go Examples

This directory contains examples demonstrating how to use the Claude Agent SDK for Go.
Ported from the official [claude-agent-sdk-demos](https://github.com/anthropics/claude-agent-sdk-demos).

## Prerequisites

All examples require:
- **Claude Code CLI** installed and authenticated
- **Go 1.21+**

Install Claude CLI:
```bash
# See https://docs.anthropic.com/claude/docs/quickstart
```

## Quick Start

```bash
# Run a simple one-shot query
go run ./examples/basic-query

# Stream a response in real-time
go run ./examples/streaming

# Interactive multi-turn conversation
go run ./examples/multi-turn

# Interactive chat application
go run ./examples/chatapp

# Multi-agent research system
go run ./examples/research-agent "artificial intelligence in healthcare"
```

## Examples Overview

| Example | Description | TypeScript Source | Complexity |
|---------|-------------|-------------------|------------|
| [basic-query](./basic-query/) | One-shot prompt/response | `hello-world` | Beginner |
| [streaming](./streaming/) | Real-time streaming responses | `hello-world-v2` | Beginner |
| [multi-turn](./multi-turn/) | Multi-turn conversations with context | `hello-world-v2` | Intermediate |
| [chatapp](./chatapp/) | Interactive chat application | `simple-chatapp` | Intermediate |
| [research-agent](./research-agent/) | Multi-agent research system | `research-agent` | Advanced |
| [custom-tools](./custom-tools/) | Custom tool implementation | - | Intermediate |
| [error-handling](./error-handling/) | Graceful error handling | - | Intermediate |
| [configuration](./configuration/) | SDK configuration options | - | Beginner |
| [hooks-lifecycle](./hooks-lifecycle/) | Hook events and lifecycle | `hello-world` | Intermediate |
| [permission-modes](./permission-modes/) | Permission system | - | Intermediate |
| [session-control](./session-control/) | Session resume and limits | `hello-world-v2` | Intermediate |
| [mcp_tools](./mcp_tools/) | MCP tool integration | - | Advanced |
| [demo](./demo/) | Full CLI demo tool | - | Intermediate |

## Example Categories

### Tier 1: Getting Started

These examples cover the fundamentals:

1. **[basic-query](./basic-query/)** - Simplest way to query Claude
   - `v2.Prompt()` for one-shot queries
   - Basic error handling
   - Timing and result extraction

2. **[streaming](./streaming/)** - Real-time response handling
   - Session-based streaming
   - Channel-based message processing
   - Different message types (assistant, delta, result, error)

3. **[multi-turn](./multi-turn/)** - Conversational context
   - Session lifecycle management
   - Context preservation across turns
   - Interactive chat interface

4. **[configuration](./configuration/)** - SDK configuration
   - Model selection and timeout options
   - Permission mode configuration
   - MCP server setup

### Tier 2: Core Features

5. **[chatapp](./chatapp/)** - Interactive chat application
   - Message queue pattern
   - Streaming responses
   - Graceful shutdown handling

6. **[custom-tools](./custom-tools/)** - Tool implementation
   - ToolExecutor interface
   - Tool registration
   - Error handling in tools

7. **[error-handling](./error-handling/)** - Error patterns
   - Error type discrimination
   - Context cancellation
   - Retry with backoff
   - Circuit breaker pattern

8. **[hooks-lifecycle](./hooks-lifecycle/)** - Hook system
   - All 12 hook event types
   - PreToolUse/PostToolUse hooks
   - Permission request hooks

9. **[permission-modes](./permission-modes/)** - Permissions
   - 6 permission modes
   - Permission behaviors
   - Rule configuration

10. **[session-control](./session-control/)** - Session management
    - Session resume and persistence
    - Session limits (turns, budget, tokens)
    - Graceful shutdown

### Tier 3: Advanced

11. **[research-agent](./research-agent/)** - Multi-agent system
    - Agent orchestration patterns
    - Parallel research execution
    - Result synthesis

12. **[mcp_tools](./mcp_tools/)** - MCP integration
    - SDK MCP server configuration
    - Tool registration and execution

## TypeScript SDK Comparison

These examples port patterns from the official [claude-agent-sdk-demos](https://github.com/anthropics/claude-agent-sdk-demos):

| TypeScript | Go | Notes |
|------------|-----|-------|
| `query()` | `v2.CreateSession()` + `Send/Receive` | Streaming query |
| `unstable_v2_prompt()` | `v2.Prompt()` | One-shot queries |
| `unstable_v2_createSession()` | `v2.CreateSession()` | Session creation |
| `unstable_v2_resumeSession()` | `v2.ResumeSession()` | Session resume |
| `await session.send()` | `session.Send()` | Send message |
| `for await (msg of session.stream())` | `for msg := range session.Receive()` | Stream responses |
| `msg.type === 'assistant'` | `msg.Type() == v2.V2EventTypeAssistant` | Type checking |

### Key Pattern Differences

**TypeScript (async/await):**
```typescript
const result = await unstable_v2_prompt('Hello', { model: 'sonnet' });
console.log(result.result);
```

**Go (context + error handling):**
```go
result, err := v2.Prompt(ctx, "Hello",
    v2.WithPromptModel("claude-sonnet-4-20250514"),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Result)
```

**TypeScript (async iteration):**
```typescript
for await (const msg of session.stream()) {
    if (msg.type === 'assistant') {
        console.log(msg.message.content);
    }
}
```

**Go (channel-based):**
```go
for msg := range session.Receive(ctx) {
    if msg.Type() == v2.V2EventTypeAssistant {
        fmt.Println(v2.ExtractAssistantText(msg))
    }
}
```

**TypeScript (event handlers):**
```typescript
client.on('message', (msg) => console.log(msg));
client.on('error', (err) => console.error(err));
```

**Go (channel select):**
```go
for {
    select {
    case msg, ok := <-msgChan:
        if !ok { return }
        fmt.Printf("%+v\n", msg)
    case err := <-errChan:
        log.Printf("Error: %v", err)
    }
}
```

## Running Examples

Each example can be run directly:

```bash
# From repository root
go run ./examples/basic-query
go run ./examples/streaming "Tell me a joke"
go run ./examples/multi-turn demo
go run ./examples/chatapp
go run ./examples/research-agent "quantum computing"
go run ./examples/hooks-lifecycle
go run ./examples/permission-modes
go run ./examples/session-control
```

Or build and install:

```bash
# Build all examples
go build ./examples/...

# Run specific examples
./examples/basic-query/basic-query "What is 2+2?"
./examples/streaming/streaming "Write a haiku"
./examples/chatapp/chatapp
```

## Troubleshooting

### "Claude CLI not found"

The SDK requires the Claude CLI to be installed and available in PATH:

```bash
# Check if CLI is available
go run ./examples/demo check

# Install Claude CLI
# See: https://docs.anthropic.com/claude/docs/quickstart
```

### "Connection timeout"

Increase the timeout for longer operations:

```go
v2.WithTimeout(120 * time.Second)
```

### "Context canceled"

The operation was interrupted. This happens on:
- SIGINT (Ctrl+C)
- Context timeout
- Explicit cancellation

Use `context.Background()` for operations that shouldn't be canceled.

## Contributing

To add a new example:

1. Create a new directory under `examples/`
2. Add `main.go` with the example code
3. Add `README.md` documenting the example
4. Update this README's table

Follow the existing patterns:
- Check CLI availability first
- Use context for cancellation
- Handle errors explicitly
- Document TypeScript equivalents

## Related Documentation

- [SDK Usage Guide](../docs/usage.md)
- [V2 API Reference](../claude/v2/)
- [TypeScript SDK Demos](https://github.com/anthropics/claude-agent-sdk-demos)
- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code)
