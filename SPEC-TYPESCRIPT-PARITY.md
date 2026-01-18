# Feature: TypeScript SDK Full Feature Parity

**Author**: Spec Creator Agent
**Date**: 2026-01-18
**Status**: Draft

---

## TL;DR

| Aspect | Detail |
|--------|--------|
| **What** | Add missing 6 features from TypeScript SDK: `tool()` helper, `createSdkMcpServer()`, `cwd` option, `tools` preset, `stderr` callback, and `canUseTool` runtime integration |
| **Why** | Achieve 100% API compatibility with `@anthropic-ai/claude-agent-sdk` for Go developers |
| **Who** | Go developers using agent-sdk-go who need feature parity with TypeScript SDK |
| **When** | When developers need MCP tool creation, in-process MCP servers, working directory control, stderr handling, or runtime permission callbacks |

---

## Problem Statement

**Current state**: Go SDK is at ~95% feature parity with TypeScript SDK. Types exist for most features, but 6 functional gaps remain:

1. **No `tool()` helper** - Developers must manually construct MCP tool definitions without type safety
2. **No `createSdkMcpServer()`** - Cannot create in-process MCP servers (only external stdio/SSE/HTTP)
3. **Missing `cwd` option** - Cannot control subprocess working directory (uses parent process cwd)
4. **Missing `tools` preset** - Cannot use `{ type: 'preset', preset: 'claude_code' }` tool configuration
5. **Missing `stderr` callback** - Cannot intercept Claude CLI stderr output for logging/debugging
6. **`canUseTool` stub only** - Types exist (`CanUseToolOptions`, `PermissionResult`) but no runtime integration with session

**Pain point**:
- Developers migrating from TypeScript cannot use these 6 features in Go
- MCP tool creation requires verbose JSON schema construction instead of type-safe helpers
- Debugging is harder without stderr access
- Permission callbacks cannot be used to implement custom authorization logic

**Impact**:
- **Critical**: Items 1-2 (tool creation UX, in-process servers)
- **High**: Items 3-4 (working directory, presets)
- **Medium**: Items 5-6 (stderr debugging, runtime callbacks)

---

## Proposed Solution

### Overview

Add the 6 missing features using Go idioms (functional options, interfaces, generics where applicable) while maintaining API compatibility with TypeScript SDK behavior. Prioritize developer experience with type safety and clear documentation.

### Architecture

```
┌──────────────────────────────────────────────────────────┐
│                     agent-sdk-go                         │
├──────────────────────────────────────────────────────────┤
│                                                          │
│  ┌─────────────┐    ┌─────────────────┐                │
│  │  v2.Session │───►│ shared.BaseOpts │                │
│  │             │    │                 │                │
│  │  +Send()    │    │ +Cwd           │◄──(NEW: #3)   │
│  │  +Receive() │    │ +Tools         │◄──(NEW: #4)   │
│  │             │    │ +Stderr        │◄──(NEW: #5)   │
│  │             │    │ +CanUseTool    │◄──(NEW: #6)   │
│  └─────────────┘    └─────────────────┘                │
│         │                    │                          │
│         ▼                    ▼                          │
│  ┌─────────────────────────────────────┐               │
│  │     subprocess.Transport             │               │
│  │                                      │               │
│  │  - Spawn with custom Cwd            │◄──(#3)       │
│  │  - Capture stderr to callback       │◄──(#5)       │
│  │  - Send canUseTool responses        │◄──(#6)       │
│  └─────────────────────────────────────┘               │
│                                                          │
│  ┌─────────────────────────────────────┐               │
│  │     shared.McpHelpers (NEW)         │               │
│  │                                      │               │
│  │  +Tool[T any](...)                  │◄──(NEW: #1)  │
│  │  +CreateSdkMcpServer(...)           │◄──(NEW: #2)  │
│  └─────────────────────────────────────┘               │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

**Components affected:**

| Component | Change Type | Description |
|-----------|-------------|-------------|
| `claude/shared/options.go` | Modified | Add `Cwd`, `Tools`, `Stderr`, `CanUseTool` fields to `BaseOptions` |
| `claude/shared/mcp_helpers.go` | New | Add `Tool()` and `CreateSdkMcpServer()` helper functions |
| `claude/subprocess/transport.go` | Modified | Apply `Cwd`, capture `Stderr`, handle `canUseTool` protocol messages |
| `claude/v2/options.go` | Modified | Add `WithCwd()`, `WithTools()`, `WithStderr()`, `WithCanUseTool()` option functions |
| `claude/shared/tools/inputs.go` | Modified | Add `ToolsConfig` discriminated union type |
| `claude/shared/permissions.go` | Modified | Add `CanUseToolCallback` function signature |

---

## User Stories

### US-1: Type-Safe MCP Tool Creation

**As a** Go developer
**I want** a type-safe `Tool()` helper function
**So that** I can define MCP tools without manually constructing JSON schemas

**Acceptance Criteria:**
- [ ] Given tool name, description, and input struct, when I call `shared.Tool[MyInput]("name", "desc", handler)`, then I get a typed `SdkMcpToolDefinition`
- [ ] Given invalid input type, when I call `Tool()`, then I get a compile-time error (not runtime)
- [ ] Given handler function with correct signature, when tool is invoked, then handler receives typed struct (not `map[string]any`)

### US-2: In-Process MCP Server

**As a** Go developer
**I want** to create an in-process MCP server
**So that** I can provide tools to Claude without spawning external processes

**Acceptance Criteria:**
- [ ] Given array of tool definitions, when I call `CreateSdkMcpServer(name, tools)`, then I get a `McpSdkServerConfig` with server instance
- [ ] Given MCP server config, when I pass it to session options, then Claude CLI can invoke my tools
- [ ] Given server instance, when session closes, then server is properly shut down

### US-3: Subprocess Working Directory Control

**As a** Go developer
**I want** to set the working directory for Claude CLI subprocess
**So that** file operations resolve relative to a specific directory

**Acceptance Criteria:**
- [ ] Given `WithCwd("/custom/dir")` option, when session starts, then subprocess uses that working directory
- [ ] Given invalid directory, when session starts, then I get a clear error (not subprocess failure)
- [ ] Given no `cwd` option, when session starts, then subprocess inherits parent process working directory (current behavior)

### US-4: Tools Preset Configuration

**As a** Go developer
**I want** to use `tools: "claude_code"` preset
**So that** I can enable Claude Code tool suite without listing every tool

**Acceptance Criteria:**
- [ ] Given `WithTools(shared.ToolsPreset("claude_code"))`, when session starts, then all Claude Code tools are available
- [ ] Given mix of preset + individual tools, when session starts, then both preset and individual tools are available
- [ ] Given invalid preset name, when session starts, then I get a clear error

### US-5: Stderr Callback

**As a** Go developer
**I want** to receive Claude CLI stderr output
**So that** I can log warnings, debug subprocess issues, or detect errors

**Acceptance Criteria:**
- [ ] Given `WithStderr(func(line string) { log.Warn(line) })` option, when Claude CLI writes to stderr, then my callback receives each line
- [ ] Given stderr callback panics, when called, then session continues (isolated failure)
- [ ] Given no stderr callback, when CLI writes to stderr, then output goes to parent process stderr (current behavior)

### US-6: Runtime Permission Callback

**As a** Go developer
**I want** to implement a `canUseTool` callback
**So that** I can approve/deny tool usage with custom logic at runtime

**Acceptance Criteria:**
- [ ] Given `WithCanUseTool(callback)` option, when Claude requests tool usage, then my callback is invoked with `ToolInput` and `CanUseToolOptions`
- [ ] Given callback returns `PermissionResult{Behavior: "allow"}`, when tool use requested, then tool executes
- [ ] Given callback returns `PermissionResult{Behavior: "deny", Message: "reason"}`, when tool use requested, then Claude receives denial message
- [ ] Given callback modifies `UpdatedInput`, when tool executes, then modified input is used
- [ ] Given callback returns suggestions, when permission denied, then Claude receives permission update suggestions

---

## Interface Design

### API Changes

#### 1. Tool Helper (`claude/shared/mcp_helpers.go`)

```go
package shared

import "context"

// ToolHandler is a typed handler for MCP tool execution.
// T is the input type (must be a struct with json tags).
// Returns tool result and optional error.
type ToolHandler[T any] func(ctx context.Context, input T, extra any) (any, error)

// SdkMcpToolDefinition represents a type-safe MCP tool definition.
type SdkMcpToolDefinition[T any] struct {
    Name        string
    Description string
    InputSchema map[string]any  // Generated from T
    Handler     ToolHandler[T]
}

// Tool creates a type-safe MCP tool definition.
// Uses reflection to generate JSON schema from T's struct tags.
//
// Example:
//   type GreetInput struct {
//       Name string `json:"name" jsonschema:"required,description=User name"`
//   }
//
//   tool := shared.Tool("greet", "Greet a user",
//       func(ctx context.Context, input GreetInput, extra any) (any, error) {
//           return map[string]string{"greeting": "Hello " + input.Name}, nil
//       })
func Tool[T any](name, description string, handler ToolHandler[T]) SdkMcpToolDefinition[T] {
    // Implementation: Use jsonschema generation library
}
```

#### 2. MCP Server Factory (`claude/shared/mcp_helpers.go`)

```go
// McpSdkServerConfigWithInstance is the result of createSdkMcpServer.
// Embeds McpSdkServerConfig and adds the server instance.
type McpSdkServerConfigWithInstance struct {
    McpSdkServerConfig
    server any // Internal MCP server instance
}

// CreateSdkMcpServer creates an in-process MCP server from tool definitions.
//
// Example:
//   server := shared.CreateSdkMcpServer("my-tools", shared.McpServerOptions{
//       Version: "1.0.0",
//       Tools: []shared.SdkMcpToolDefinition[any]{greetTool, calculateTool},
//   })
//
//   session, _ := v2.CreateSession(ctx,
//       v2.WithMcpServers(map[string]shared.McpServerConfig{
//           "my-tools": server,
//       }))
func CreateSdkMcpServer(name string, opts McpServerOptions) McpSdkServerConfigWithInstance {
    // Implementation: Create in-process MCP server using SDK protocol
}

// McpServerOptions configures an in-process MCP server.
type McpServerOptions struct {
    Version string
    Tools   []SdkMcpToolDefinition[any]
}
```

#### 3. Options Extensions (`claude/shared/options.go`)

```go
// Add to BaseOptions struct:

type BaseOptions struct {
    // ... existing fields ...

    // Cwd sets the working directory for the subprocess.
    // If empty, inherits parent process working directory.
    Cwd string

    // Tools configures tool availability.
    // Can be a preset ("claude_code") or explicit tool list.
    Tools *ToolsConfig

    // Stderr is a callback invoked for each stderr line from subprocess.
    // If nil, stderr goes to parent process stderr.
    Stderr func(line string)

    // CanUseTool is a callback for runtime permission checks.
    // Invoked before each tool use when permission mode requires it.
    CanUseTool CanUseToolCallback
}

// ToolsConfig is a discriminated union for tool configuration.
type ToolsConfig struct {
    Type   string   // "preset" | "explicit"
    Preset string   // "claude_code" (when Type == "preset")
    Tools  []string // Tool names (when Type == "explicit")
}

// ToolsPreset creates a tools config for a preset.
func ToolsPreset(preset string) *ToolsConfig {
    return &ToolsConfig{Type: "preset", Preset: preset}
}

// ToolsExplicit creates a tools config for explicit tool list.
func ToolsExplicit(tools ...string) *ToolsConfig {
    return &ToolsConfig{Type: "explicit", Tools: tools}
}
```

#### 4. Permission Callback (`claude/shared/permissions.go`)

```go
// CanUseToolCallback is invoked before tool execution for permission checks.
// Returns PermissionResult indicating whether to allow/deny the tool use.
//
// The callback can:
// - Approve: Return PermissionResult{Behavior: "allow"}
// - Deny: Return PermissionResult{Behavior: "deny", Message: "reason"}
// - Modify input: Set UpdatedInput field
// - Suggest permission updates: Set UpdatedPermissions field
//
// Context may be cancelled if session times out.
type CanUseToolCallback func(
    ctx context.Context,
    toolName string,
    toolInput map[string]any,
    opts CanUseToolOptions,
) (PermissionResult, error)
```

#### 5. Session Option Functions (`claude/v2/options.go`)

```go
// WithCwd sets the working directory for the Claude CLI subprocess.
func WithCwd(cwd string) SessionOption {
    return func(opts *V2SessionOptions) {
        opts.Cwd = cwd
    }
}

// WithTools sets the tools configuration (preset or explicit list).
func WithTools(tools *shared.ToolsConfig) SessionOption {
    return func(opts *V2SessionOptions) {
        opts.Tools = tools
    }
}

// WithStderr sets a callback for subprocess stderr output.
func WithStderr(callback func(line string)) SessionOption {
    return func(opts *V2SessionOptions) {
        opts.Stderr = callback
    }
}

// WithCanUseTool sets a permission callback for runtime tool approval.
func WithCanUseTool(callback shared.CanUseToolCallback) SessionOption {
    return func(opts *V2SessionOptions) {
        opts.CanUseTool = callback
    }
}

// Prompt option variants:
// WithPromptCwd, WithPromptTools, WithPromptStderr, WithPromptCanUseTool
```

#### 6. Transport Integration (`claude/subprocess/transport.go`)

```go
// Modify startProcess() to:
// 1. Set cmd.Dir = options.Cwd (if non-empty)
// 2. Start goroutine to read stderr and invoke options.Stderr callback
// 3. Listen for "canUseTool" request messages from CLI
// 4. Invoke options.CanUseTool callback and send response back to CLI

// Add to Transport struct:
type Transport struct {
    // ... existing fields ...

    stderrReader *bufio.Scanner
    canUseToolCallback shared.CanUseToolCallback
}

// Protocol message types to add:
type CanUseToolRequest struct {
    Type      string         `json:"type"` // "canUseTool"
    ToolName  string         `json:"toolName"`
    ToolInput map[string]any `json:"toolInput"`
    Options   shared.CanUseToolOptions `json:"options"`
}

type CanUseToolResponse struct {
    Type   string                 `json:"type"` // "canUseToolResponse"
    Result shared.PermissionResult `json:"result"`
}
```

---

## Implementation Notes

### Approach

**Phase 1: Low-hanging fruit (cwd, stderr, tools preset)**
1. Add fields to `BaseOptions`
2. Add option functions to `v2/options.go`
3. Modify `subprocess/transport.go` to apply cwd and capture stderr
4. Add `ToolsConfig` type and validation

**Phase 2: MCP helpers (tool, createSdkMcpServer)**
1. Add `jsonschema` dependency for schema generation from structs
2. Implement `Tool[T]()` with reflection-based schema generation
3. Implement `CreateSdkMcpServer()` using in-process MCP protocol
4. Add tests for tool invocation and type safety

**Phase 3: Runtime integration (canUseTool)**
1. Add protocol message types for `canUseTool` request/response
2. Modify transport to listen for `canUseTool` messages
3. Invoke callback and send response back to CLI
4. Add timeout handling (default 30s for permission callbacks)

### Dependencies

| Dependency | Type | Notes |
|------------|------|-------|
| `github.com/invopop/jsonschema` | External | For generating JSON schemas from Go structs (already used in similar projects) |
| `bufio.Scanner` | Stdlib | For line-buffered stderr reading |
| Existing subprocess transport | Internal | Modify to handle new protocol messages |

### Alternatives Considered

| Approach | Pros | Cons | Why Not |
|----------|------|------|---------|
| **Tool(): Code generation instead of reflection** | No runtime overhead, compile-time schema errors | Requires build step, more setup | Reflection is Go-idiomatic for struct tags, acceptable perf cost |
| **Stderr: Channel instead of callback** | More Go-idiomatic concurrency | Requires caller to consume channel or risk blocking | Callback matches TypeScript API, simpler for users |
| **CanUseTool: Hooks-based (reuse existing hook system)** | Reuses existing hook infrastructure | Hooks are session-level, canUseTool needs transport-level integration | Permission callbacks need lower-level access than hooks provide |
| **Tools preset: CLI flag `--tools=preset:claude_code`** | Simpler Go implementation | Less type-safe, harder to compose with explicit tools | Discriminated union provides better type safety |

---

## Test Plan

### Unit Tests

| Test | Input | Expected Output |
|------|-------|-----------------|
| `Tool[T]()` with valid struct | Struct with json tags | Valid `SdkMcpToolDefinition` with correct schema |
| `Tool[T]()` with missing json tags | Struct without tags | Schema uses field names as-is |
| `ToolsPreset()` | "claude_code" | `ToolsConfig{Type: "preset", Preset: "claude_code"}` |
| `ToolsExplicit()` | []string{"Read", "Write"} | `ToolsConfig{Type: "explicit", Tools: [...]}` |
| `WithCwd()` option | "/custom/dir" | Options.Cwd == "/custom/dir" |
| `WithStderr()` option | callback function | Options.Stderr == callback |
| Transport with cwd | cmd.Dir set | Subprocess starts in custom directory |
| Transport stderr capture | CLI writes to stderr | Callback invoked with each line |
| canUseTool allow | Callback returns allow | Tool executes |
| canUseTool deny | Callback returns deny | Tool blocked, message sent to Claude |

### Integration Tests

| Scenario | Steps | Expected |
|----------|-------|----------|
| **Happy path: tool() helper** | 1. Define tool with `Tool[MyInput]()`, 2. Register with MCP server, 3. Invoke from session | Handler receives typed input, returns result |
| **Happy path: createSdkMcpServer** | 1. Create server with tools, 2. Attach to session, 3. Invoke tools | All tools execute successfully |
| **Happy path: cwd option** | 1. Create session with `WithCwd("/tmp")`, 2. Use Read tool on relative path | File resolved from `/tmp` |
| **Happy path: stderr callback** | 1. Create session with stderr callback, 2. Trigger CLI warning | Callback receives warning line |
| **Happy path: canUseTool allow** | 1. Set callback that allows all, 2. Request tool use | Tool executes |
| **Happy path: canUseTool deny** | 1. Set callback that denies Write tool, 2. Request Write | Claude receives denial, doesn't execute |
| **Edge: cwd nonexistent** | Create session with invalid cwd | Session creation fails with clear error |
| **Edge: stderr callback panics** | Callback panics on invoke | Session continues, error logged |
| **Edge: canUseTool timeout** | Callback takes >30s | Request times out, tool denied |
| **Error: invalid tool input type** | Pass non-struct to `Tool[T]()` | Compile-time error (generics constraint) |
| **Error: invalid preset name** | `ToolsPreset("invalid")` | Session start fails with error |

### Performance Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| Tool schema generation | < 1ms per tool | Benchmark `Tool[T]()` with 100 tools |
| Stderr callback overhead | < 100μs per line | Benchmark stderr capture with high-frequency output |
| canUseTool callback latency | < 10ms (not counting user logic) | Benchmark transport round-trip for canUseTool request/response |

---

## Security Considerations

| Risk | Mitigation |
|------|------------|
| **Arbitrary cwd allows path traversal** | Validate cwd exists and is absolute before starting subprocess. Document that cwd affects all file operations. |
| **Stderr leaks sensitive data** | Document that stderr may contain API keys, file paths, or other sensitive data. User is responsible for filtering. |
| **canUseTool callback blocks session** | Enforce 30s timeout on callback execution. Log warning if callback takes >5s. |
| **MCP server instance lifetime** | Ensure server is shut down when session closes to prevent resource leaks. |
| **Tool handler panics crash session** | Wrap handler calls in recover() and return error to Claude instead of crashing. |

---

## Rollback Plan

**If feature causes issues in production:**

1. **Feature flags**: Not applicable (library, not service)
2. **Version rollback**: Users can pin to previous version in `go.mod`
3. **Graceful degradation**: All 6 features are opt-in via options, existing code continues to work

**Data implications:**
- No persistent data changes
- Sessions using new features will fail if rolled back to old version

**Mitigation:**
- Comprehensive test coverage before release
- Document minimum supported CLI version for new features
- Add version compatibility matrix to README

---

## Open Questions

| Question | Owner | Status |
|----------|-------|--------|
| Which jsonschema library to use? (`invopop/jsonschema` vs `alecthomas/jsonschema`) | Implementation team | **Resolved**: Use `invopop/jsonschema` (more active, better struct tag support) |
| Should `canUseTool` callbacks run in separate goroutines? | Architecture team | **Open** - Could improve responsiveness but adds complexity |
| Does Claude CLI support in-process MCP servers via stdin/stdout protocol? | Claude SDK team | **Open** - May require CLI flag or beta feature |
| Should `Tool[T]()` cache generated schemas? | Performance team | **Resolved**: Yes, use sync.Map for schema cache per type |

---

## Implementation Tasks

Tasks sized for 10-20 minute completion by an LLM agent or developer.

### Milestone 1: Foundation (Options & Types)

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M1.1 | Add `Cwd`, `Tools`, `Stderr`, `CanUseTool` fields to `BaseOptions` struct | `claude/shared/options.go` | `claude/shared/options.go` (modified) |
| M1.2 | Add `ToolsConfig` type with `ToolsPreset()` and `ToolsExplicit()` constructors | - | `claude/shared/tools/inputs.go` |
| M1.3 | Add `CanUseToolCallback` function type to `permissions.go` | `claude/shared/permissions.go` | `claude/shared/permissions.go` (modified) |
| M1.4 | Add `WithCwd()`, `WithTools()`, `WithStderr()`, `WithCanUseTool()` to `v2/options.go` | M1.1, M1.2, M1.3 | `claude/v2/options.go` (modified) |
| M1.5 | Add corresponding `WithPrompt*` variants for one-shot prompts | M1.4 | `claude/v2/options.go` (modified) |
| M1.6 | Write unit tests for option functions | M1.4, M1.5 | `claude/v2/options_test.go` |

### Milestone 2: Subprocess Integration (Cwd, Stderr)

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M2.1 | Modify `transport.startProcess()` to set `cmd.Dir = options.Cwd` | M1.1, `claude/subprocess/transport.go` | `claude/subprocess/transport.go` (modified) |
| M2.2 | Add cwd validation (exists, absolute path) before subprocess start | M2.1 | `claude/subprocess/transport.go` (modified) |
| M2.3 | Create stderr capture goroutine in `startProcess()` | M2.1 | `claude/subprocess/transport.go` (modified) |
| M2.4 | Invoke `options.Stderr` callback for each stderr line | M2.3 | `claude/subprocess/transport.go` (modified) |
| M2.5 | Add panic recovery for stderr callback | M2.4 | `claude/subprocess/transport.go` (modified) |
| M2.6 | Write unit tests for cwd application | M2.1, M2.2 | `claude/subprocess/transport_test.go` |
| M2.7 | Write unit tests for stderr capture | M2.3, M2.4, M2.5 | `claude/subprocess/transport_test.go` |

### Milestone 3: Tools Preset Support

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M3.1 | Add `ToolsConfig` to CLI args generation in transport | M1.2, `claude/subprocess/transport.go` | `claude/subprocess/transport.go` (modified) |
| M3.2 | Map `ToolsConfig{Type: "preset"}` to `--tools=preset:<name>` CLI flag | M3.1 | `claude/subprocess/transport.go` (modified) |
| M3.3 | Map `ToolsConfig{Type: "explicit"}` to `--tools=<tool1>,<tool2>` CLI flag | M3.1 | `claude/subprocess/transport.go` (modified) |
| M3.4 | Add validation for preset names ("claude_code" currently only valid value) | M3.2 | `claude/shared/options.go` |
| M3.5 | Write unit tests for tools preset CLI arg generation | M3.2, M3.3, M3.4 | `claude/subprocess/transport_test.go` |

### Milestone 4: MCP Tool Helper

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M4.1 | Add `invopop/jsonschema` dependency to `go.mod` | - | `go.mod` |
| M4.2 | Create `claude/shared/mcp_helpers.go` | - | `claude/shared/mcp_helpers.go` (new) |
| M4.3 | Define `ToolHandler[T]` and `SdkMcpToolDefinition[T]` types | M4.2 | `claude/shared/mcp_helpers.go` |
| M4.4 | Implement `Tool[T]()` function with schema generation | M4.1, M4.3 | `claude/shared/mcp_helpers.go` |
| M4.5 | Add schema caching using `sync.Map` keyed by type | M4.4 | `claude/shared/mcp_helpers.go` |
| M4.6 | Add handler wrapper to convert `map[string]any` to typed `T` | M4.4 | `claude/shared/mcp_helpers.go` |
| M4.7 | Add panic recovery for tool handlers | M4.6 | `claude/shared/mcp_helpers.go` |
| M4.8 | Write unit tests for `Tool[T]()` with various input types | M4.4-M4.7 | `claude/shared/mcp_helpers_test.go` |

### Milestone 5: In-Process MCP Server

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M5.1 | Define `McpServerOptions` struct | M4.3 | `claude/shared/mcp_helpers.go` |
| M5.2 | Define `McpSdkServerConfigWithInstance` struct | `claude/shared/mcp.go` | `claude/shared/mcp_helpers.go` |
| M5.3 | Implement `CreateSdkMcpServer()` function skeleton | M5.1, M5.2 | `claude/shared/mcp_helpers.go` |
| M5.4 | Create in-process MCP server instance using stdio protocol | M5.3 | `claude/shared/mcp_helpers.go` |
| M5.5 | Register tools with MCP server instance | M5.4 | `claude/shared/mcp_helpers.go` |
| M5.6 | Add server lifecycle management (start on session create, stop on close) | M5.4 | `claude/v2/session.go` |
| M5.7 | Write integration tests for in-process MCP server | M5.3-M5.6 | `claude/shared/mcp_helpers_integration_test.go` |

### Milestone 6: CanUseTool Runtime Integration

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M6.1 | Add `CanUseToolRequest` and `CanUseToolResponse` message types | `claude/shared/message.go` | `claude/shared/message.go` (modified) |
| M6.2 | Add message parser cases for canUseTool request/response | M6.1, `claude/parser/parser.go` | `claude/parser/parser.go` (modified) |
| M6.3 | Add canUseTool request handler to transport message loop | M6.1, `claude/subprocess/transport.go` | `claude/subprocess/transport.go` (modified) |
| M6.4 | Invoke `options.CanUseTool` callback with timeout (30s default) | M6.3 | `claude/subprocess/transport.go` (modified) |
| M6.5 | Send `CanUseToolResponse` back to CLI via stdin | M6.4 | `claude/subprocess/transport.go` (modified) |
| M6.6 | Add error handling for callback errors (deny with error message) | M6.4 | `claude/subprocess/transport.go` (modified) |
| M6.7 | Write unit tests for canUseTool protocol integration | M6.3-M6.6 | `claude/subprocess/transport_test.go` |
| M6.8 | Write integration tests for canUseTool allow/deny scenarios | M6.7 | `claude/v2/session_integration_test.go` |

### Milestone 7: Documentation & Examples

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M7.1 | Add `Tool[T]()` example to `examples/mcp-tools/` | M4.8 | `examples/mcp-tools/typed-tool.go` |
| M7.2 | Add `CreateSdkMcpServer()` example to `examples/mcp-server/` | M5.7 | `examples/mcp-server/in-process-server.go` |
| M7.3 | Add `WithCwd()` usage to `examples/basic/` | M2.6 | `examples/basic/custom-working-dir.go` |
| M7.4 | Add `WithStderr()` usage to `examples/debugging/` | M2.7 | `examples/debugging/stderr-logging.go` |
| M7.5 | Add `WithCanUseTool()` example to `examples/permissions/` | M6.8 | `examples/permissions/custom-approval.go` |
| M7.6 | Update `README.md` with feature parity checklist | All above | `README.md` (modified) |
| M7.7 | Update `CLAUDE.md` SDK Feature Status section | All above | `CLAUDE.md` (modified) |
| M7.8 | Create `docs/typescript-parity.md` migration guide | All above | `docs/typescript-parity.md` (new) |

### Summary

| Milestone | Tasks | Focus |
|-----------|-------|-------|
| M1 | 6 | Add option fields and types |
| M2 | 7 | Subprocess cwd and stderr handling |
| M3 | 5 | Tools preset CLI integration |
| M4 | 8 | Type-safe tool helper |
| M5 | 7 | In-process MCP server |
| M6 | 8 | canUseTool runtime callbacks |
| M7 | 8 | Documentation and examples |

**Total: 7 milestones, 49 tasks**

### MVP Cutoff

For minimal working feature parity (most critical items):
- **Complete**: M1, M2, M3, M4 (26 tasks) - Covers items #1, #3, #4, #5
- **Deferred**: M5, M6, M7 - Can be added in follow-up releases

This provides:
- ✅ `tool()` helper (item #1)
- ✅ `cwd` option (item #3)
- ✅ `tools` preset (item #4)
- ✅ `stderr` callback (item #5)
- ⏳ `createSdkMcpServer()` (item #2) - deferred to M5
- ⏳ `canUseTool` integration (item #6) - deferred to M6

Achieves ~85% of value with 53% of effort.

---

## References

- TypeScript SDK: `@anthropic-ai/claude-agent-sdk` [NPM](https://www.npmjs.com/package/@anthropic-ai/claude-agent-sdk)
- MCP Protocol: [Model Context Protocol](https://modelcontextprotocol.io/)
- JSON Schema from Go structs: [invopop/jsonschema](https://github.com/invopop/jsonschema)
- Existing SDK port spec: `SPEC-ANTHROPIC-SDK-PORT.md` in repository
- Enhancement plan: `SDK-ENHANCEMENT-PLAN.md` in repository
