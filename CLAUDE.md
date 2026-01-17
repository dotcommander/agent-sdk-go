# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`agent-sdk-go` is a **CLI wrapper** around the Claude CLI:
- **CLI tool** (`cmd/`) for interacting with Claude via subprocess
- **Internal packages** (`internal/`) - not importable as a library

**Note:** This is NOT an HTTP client SDK. It spawns the Claude CLI as a subprocess.

## Build & Development

### Installation
```bash
# Build binary
go build -o agent-sdk-go ./cmd

# Install globally (option 1)
go install ./cmd

# Symlink (option 2 - keeps binary in project)
ln -sf "$(pwd)/agent-sdk-go" ~/go/bin/agent-sdk-go

# Release build with version info
go build -ldflags "-X main.version=$(git describe --tags 2>/dev/null || echo dev)" -o agent-sdk-go ./cmd
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
| internal/app | 93% |
| internal/claude/parser | 83% |
| internal/claude/cli | 56% |
| internal/claude/subprocess | 44% |
| internal/claude/shared | 26% |
| internal/claude/v2 | 5% (integration tests excluded) |

### Single Test Execution
```bash
# Run specific test
go test ./internal/claude -run TestClient

# Run with verbose output
go test ./internal/claude/parser -v -run TestParser

# Run with race detector
go test ./... -race
```

## Architecture

### Structure
```
agent-sdk-go/
├── cmd/                    # CLI entry point
│   └── agent/              # Main binary
├── internal/               # Internal packages (not importable)
│   ├── app/                # Application configuration
│   └── claude/             # Claude CLI wrapper
│       ├── shared/         # Shared types, options, factory
│       │   ├── agents.go       # Agent definitions
│       │   ├── hooks.go        # Hook system (12 event types)
│       │   ├── mcp.go          # MCP server configurations
│       │   ├── message.go      # Message types
│       │   ├── options.go      # Base options (25+ fields)
│       │   ├── permissions.go  # Permission system
│       │   ├── sandbox.go      # Sandbox settings
│       │   ├── types.go        # Account, model, command info
│       │   ├── usage.go        # Usage tracking types
│       │   └── tools/          # Tool input schemas
│       ├── cli/            # CLI discovery and availability
│       ├── parser/         # JSON message parsing
│       ├── subprocess/     # Process management, transport
│       └── v2/             # V2 session/prompt API
└── docs/                   # Documentation
    └── usage.md            # Comprehensive usage guide
```

### Key Patterns

1. **Functional Options**: Client and session configuration via options pattern (`WithBaseURL`, `WithTimeout`)
2. **Interface-based Tool System**: `ToolExecutor` interface for custom tool implementations
3. **Separation of Concerns**: SDK library vs. CLI commands vs. business logic
4. **Concurrency Ready**: Uses goroutines and channels for streaming (hand-rolled SSE parser)

### Configuration Loading Order
1. `./config.yaml` (current directory)
2. `~/.config/agent-sdk-go/config.yaml`
3. `/etc/agent-sdk-go/config.yaml`
4. Environment variables (prefix: `AGENT-SDK-GO_`)

### State Persistence
```go
// Save/load runtime state between executions
gokart.SaveState("agent-sdk-go", "state.json", myState)
state, _ := gokart.LoadState[MyState]("agent-sdk-go", "state.json")
```

## Common Development Tasks

### Adding New CLI Commands
1. Add command definition in `internal/commands/`
2. Register in `internal/commands/root.go`
3. Add business logic in `internal/actions/`
4. Add tests in `internal/actions/*_test.go`

### Extending the Internal Packages
1. Add types to `internal/claude/shared/types.go`
2. Implement functionality in appropriate `internal/claude/*` package
3. Add corresponding tests in `*_test.go` files
4. Maintain test coverage (target 80%+)

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

## CLI Usage Examples

```bash
# Set API key
export ANTHROPIC_API_KEY="sk-ant-..."

# Interactive chat
agent-sdk-go agent run --model "claude-3-5-sonnet-20241022"

# Stream responses
agent-sdk-go agent stream

# Test tool execution
agent-sdk-go agent tool

# Example greet command
agent-sdk-go greet --name "Alice"
```

## Go API Usage (Internal)

```go
// NOTE: These packages are internal and not importable from external projects.
// This is a CLI tool, not a library.

import "agent-sdk-go/internal/claude/v2"

// Create a session (requires Claude CLI installed)
session, err := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-20250514"),
    v2.WithTimeout(30*time.Second),
)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// Send a message
session.Send("Hello!")
resp, err := session.SendMessage(ctx)

// Or use one-shot prompt
result, err := v2.Prompt(ctx, "What is 2+2?",
    v2.WithPromptModel("claude-sonnet-4-20250514"),
)
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

This is a **CLI subprocess wrapper**, not an HTTP client SDK:
- Spawns `claude` CLI as a subprocess
- Parses JSON output from CLI stdout
- Requires Claude CLI to be installed and authenticated
- Uses Go idioms (interfaces, error wrapping, functional options)
- Table-driven tests with testify assertions

### Critical Gotchas

**One-shot stdin pipe:** Do NOT create stdin pipe for one-shot prompts - causes indefinite hang. The subprocess waits for stdin EOF before processing. See `subprocess/transport.go`.

**Channel buffer sizing:** Message channel uses buffer of 100. This is unbenchmarked - consider profiling under load if backpressure occurs.

**Coverage reality:** The v2 package shows 5% coverage because integration tests require Claude CLI. Unit test coverage is higher but requires the subprocess to actually run.

## Scale Limitations

This CLI subprocess wrapper has several scale limitations that should be considered for production workloads:

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

- `docs/usage.md` - Comprehensive usage guide with examples
- `SPEC-ANTHROPIC-SDK-PORT.md` - Original implementation specification
- `SDK-ENHANCEMENT-PLAN.md` - Detailed roadmap (now complete)
- `README.md` - Project overview and quick start