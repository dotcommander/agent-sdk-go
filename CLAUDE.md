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
│       ├── cli/            # CLI discovery and availability
│       ├── parser/         # JSON message parsing
│       ├── subprocess/     # Process management, transport
│       └── v2/             # V2 session/prompt API
└── docs/                   # Audit and analysis documents
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

## Enhancement Plan Status

See `SDK-ENHANCEMENT-PLAN.md` for comprehensive roadmap. Current implementation provides **core functionality** but lacks advanced features:

**Missing (P0-P3 priority)**:
- Permissions system and hooks framework
- User input workflows with timeouts
- Session forking capability
- File checkpointing for state restoration
- Structured outputs with JSON Schema validation
- System prompt presets and configuration
- MCP (Model Context Protocol) compatibility
- Subagent hierarchy support
- Slash command ecosystem integration

## Implementation Notes

This is a **CLI subprocess wrapper**, not an HTTP client SDK:
- Spawns `claude` CLI as a subprocess
- Parses JSON output from CLI stdout
- Requires Claude CLI to be installed and authenticated
- Uses Go idioms (interfaces, error wrapping, functional options)
- Table-driven tests with testify assertions

## Project References

- `SPEC-ANTHROPIC-SDK-PORT.md` - Original implementation specification
- `SDK-ENHANCEMENT-PLAN.md` - Detailed roadmap for missing features
- `README.md` - Comprehensive documentation and examples