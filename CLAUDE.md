# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`agent-sdk-go` is a **public Go SDK** for building agents with Claude Code:
- **Public packages** (`claude/`) - importable as `github.com/dotcommander/agent-sdk-go/claude`
- **Subprocess transport** - communicates with Claude CLI via stdin/stdout streaming
- **Examples** (`examples/`) - usage demonstrations

**Note:** This SDK spawns the Claude CLI as a subprocess (not HTTP API). Requires Claude CLI installed.

## Build & Development

### Installation
```bash
# Install the SDK
go get github.com/dotcommander/agent-sdk-go

# Or add to go.mod
require github.com/dotcommander/agent-sdk-go latest
```

### Testing
```bash
# Run all tests
go test ./...

# With coverage report
go test ./... -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# Run integration tests (requires Claude CLI)
go test -tags=integration ./...
```

**Coverage by package:**
| Package | Coverage |
|---------|----------|
| claude/parser | 83% |
| claude/cli | 56% |
| claude/subprocess | 44% |
| internal/shared | 26% |
| claude/v2 | 5% (integration tests excluded) |

### Single Test Execution
```bash
# Run specific test
go test ./claude -run TestClient

# Run with verbose output
go test ./claude/parser -v -run TestParser

# Run with race detector
go test ./... -race
```

## Architecture

### Structure
```
agent-sdk-go/
├── claude/                 # Public SDK package (import this)
│   ├── client.go           # High-level Client interface
│   ├── types.go            # Message types, interfaces
│   ├── options.go          # Client configuration options
│   ├── exports.go          # Re-exports from internal/shared
│   ├── cli/                # CLI discovery and availability
│   ├── parser/             # JSON message parsing
│   ├── subprocess/         # Process management, transport
│   ├── mcp/                # MCP server support
│   └── v2/                 # V2 session/prompt API
├── internal/               # Internal packages (not imported by users)
│   └── shared/             # Shared types, options, factory (private)
│       ├── agents.go       # Agent definitions
│       ├── hooks.go        # Hook system (12 event types)
│       ├── mcp.go          # MCP server configurations
│       ├── message.go      # Message types
│       ├── options.go      # Base options (25+ fields)
│       ├── permissions.go  # Permission system
│       ├── sandbox.go      # Sandbox settings
│       └── tools/          # Tool input schemas
├── examples/               # Usage examples
└── docs/                   # Documentation
```

### Key Patterns

1. **Functional Options**: Client and session configuration via options pattern (`WithBaseURL`, `WithTimeout`)
2. **Interface-based Tool System**: `ToolExecutor` interface for custom tool implementations
3. **Separation of Concerns**: SDK library vs. CLI commands vs. business logic
4. **Concurrency Ready**: Uses goroutines and channels for streaming (hand-rolled SSE parser)

## Common Development Tasks

### Extending the SDK
1. Add types to `internal/shared/types.go`
2. Implement functionality in appropriate `claude/*` or `internal/shared/*` package
3. Re-export public types/functions via `claude/exports.go` if user-facing
4. Add corresponding tests in `*_test.go` files
5. Maintain test coverage (target 80%+)

### Tool Development Pattern
```go
// 1. Define tool in types.go
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema InputSchema `json:"input_schema"`
}

// 2. Implement ToolExecutor interface
type MyToolExecutor struct{}

func (e *MyToolExecutor) Execute(ctx context.Context, toolName string, args map[string]any) (any, error) {
    // Implementation
}

// 3. Register with client
err := client.RegisterTool(toolDefinition, &MyToolExecutor{})
```

## SDK Usage Examples

```go
import "github.com/dotcommander/agent-sdk-go/claude"

// Simple one-shot query
client, _ := claude.NewClient()
response, err := client.Query(ctx, "What is 2+2?")

// Streaming responses
msgChan, errChan := client.QueryStream(ctx, "Tell me a story")
for msg := range msgChan {
    fmt.Printf("%T: %+v\n", msg, msg)
}

// Interactive session
client.Connect(ctx)
defer client.Disconnect()
msgChan, errChan := client.ReceiveMessages(ctx)

// V2 Session API
import "github.com/dotcommander/agent-sdk-go/claude/v2"

session, err := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-20250514"),
    v2.WithTimeout(30*time.Second),
)
defer session.Close()

session.Send("Hello!")
resp, err := session.SendMessage(ctx)
```

## SDK Feature Status

Full TypeScript SDK port completed. All types from `@anthropic-ai/claude-agent-sdk` are now available:

**Implemented:**
- **Permissions system** - 6 permission modes, behaviors, updates
- **Hooks framework** - 12 hook events with typed inputs/outputs
- **MCP support** - Stdio, SSE, HTTP, SDK server configs
- **Agent definitions** - Custom subagents with model/tool config
- **Sandbox settings** - Network config, ripgrep, command exclusions
- **Session control** - Resume, fork, persist options
- **Limits** - Max turns, budget, thinking tokens
- **Tool schemas** - 20+ typed tool input definitions
- **Message types** - All SDK message variants with parsers

**Query interface methods** (stubs, transport integration pending):
- `Interrupt()`, `SetPermissionMode()`, `SetModel()`
- `SupportedCommands()`, `SupportedModels()`, `McpServerStatus()`
- `AccountInfo()`, `RewindFiles()`, `SetMcpServers()`

## Implementation Notes

This SDK uses **subprocess transport** to communicate with Claude CLI:
- Spawns `claude` CLI as a subprocess with `--output-format stream-json`
- Bidirectional communication via stdin/stdout pipes
- Parses JSON streaming output from CLI
- Requires Claude CLI to be installed and authenticated
- Uses Go idioms (interfaces, error wrapping, functional options)
- Table-driven tests with testify assertions

### Critical Gotchas

**One-shot stdin pipe:** Do NOT create stdin pipe for one-shot prompts - causes indefinite hang. The subprocess waits for stdin EOF before processing. See `subprocess/transport.go`.

**Channel buffer sizing:** Message channel uses buffer of 100. This is unbenchmarked - consider profiling under load if backpressure occurs.

**Coverage reality:** The v2 package shows 5% coverage because integration tests require Claude CLI. Unit test coverage is higher but requires the subprocess to actually run.

## Scale Limitations

This SDK's subprocess transport has several scale limitations to consider for production workloads:

### 1. Message Channel Buffer = 100
**Trigger:** High-throughput applications (>10 messages/second)
**Impact:** Unbenchmarked buffer could overflow under load, causing dropped messages
**Estimated Limit:** ~100 concurrent messages before blocking
**Mitigation:**
- Implement adaptive backpressure in `subprocess/transport.go`
- Add message batching for high-frequency scenarios
- Consider channel pooling in future versions

### 2. One Subprocess Per Session
**Trigger:** Massively concurrent sessions (>50 simultaneous)
**Impact:** No process pooling, OS-level process creation overhead
**Estimated Limit:** ~1000 sessions (OS-dependent, constrained by file descriptors)
**Mitigation:**
- Implement process reuse/session pooling
- Add session affinity for recurring interactions
- Use connection multiplexing for adjacent requests

### 3. O(n²) JSON Parsing for Large Messages
**Trigger:** Messages >100KB (e.g., code analysis, document processing)
**Impact:** Quadratic complexity in message framing and validation
**Estimated Limit:** Degradation above 500KB messages
**Mitigation:**
- Implement streaming JSON parser for large payloads
- Add message size limits and warnings
- Use `json.Decoder` with buffered reader instead of unmarshal

### 4. No Backpressure Handling
**Trigger:** Fast producer, slow consumer scenarios
**Impact:** Memory leaks and potential goroutine leaks
**Estimated Limit:** Unbounded memory growth under load
**Mitigation:**
- Implement circuit breaker pattern for message consumption
- Add request timeout and retry mechanisms
- Monitor memory usage and scale limits

### 5. Registry RWMutex Contention Under High Concurrency
**Trigger:** >100 concurrent operations on shared registries
**Impact:** Performance degradation due to lock contention
**Estimated Limit:** ~200 concurrent operations before significant slowdown
**Mitigation:**
- Implement sharded registries (e.g., per-model or per-region)
- Use RWMutex with reader/writer priority
- Consider lock-free data structures for high-frequency updates

### Scale Recommendations
- **Small Scale:** (<50 sessions) Current implementation is sufficient
- **Medium Scale:** (50-500 sessions) Add monitoring and consider pooling
- **Large Scale:** (>500 sessions) Architectural refactor required

## Project References

- `docs/` - Comprehensive documentation (getting-started, configuration, streaming, etc.)
- `examples/` - 21 usage examples with READMEs
- `README.md` - Project overview and quick start