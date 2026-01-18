# Feature: Priority Enhancements from severity1/claude-agent-sdk-go

**Author**: Spec Creator Agent
**Date**: 2026-01-18
**Status**: Draft
**Reference**: Comparison between severity1/claude-agent-sdk-go and dotcommander/agent-sdk-go

---

## TL;DR

| Aspect | Detail |
|--------|--------|
| **What** | Add 7 feature groups from severity1 SDK: SDK MCP Server (P0), Lifecycle Hooks (P0), System Prompt Options (P1), Typed Permission Results (P1), As*Error Helpers (P1), Debug Writer Options (P2), Sandbox Network Config (P3) |
| **Why** | Achieve feature parity with severity1 fork, enable in-process tool hosting, lifecycle observability, and improved developer experience |
| **Who** | Go developers using agent-sdk-go who need MCP tool creation, hook integration, or enhanced debugging |
| **Impact** | P0 features (SDK MCP, Hooks) enable critical new workflows; P1-P3 improve DX and reduce friction |
| **Effort** | P0: 2-3 days, P1: 1 day, P2-P3: 4-6 hours |

---

## Problem Statement

**Current state**: The dotcommander/agent-sdk-go is missing 7 feature groups present in severity1/claude-agent-sdk-go:

| Feature | Priority | Current Gap | Severity1 Status |
|---------|----------|-------------|------------------|
| SDK MCP Server | P0 | No in-process MCP hosting | `CreateSDKMcpServer()`, `NewTool()`, `WithSdkMcpServer()` implemented |
| Lifecycle Hooks | P0 | Hook types exist, no invocation | 12 hook events, callback registration, control protocol integration |
| System Prompt Options | P1 | No `WithSystemPrompt()` | `WithSystemPrompt()`, `WithAppendSystemPrompt()` |
| Typed Permission Results | P1 | Generic `PermissionResult` struct | `PermissionResultAllow`, `PermissionResultDeny` constructors |
| As*Error Helpers | P1 | ✅ Already implemented | ✅ `AsCLINotFoundError()`, etc. (commit 2906253) |
| Debug Writer Options | P2 | No debug output control | `WithDebugWriter()`, `WithDebugStderr()`, `WithDebugDisabled()` |
| Sandbox Network Config | P3 | ✅ Already implemented | ✅ `SandboxNetworkConfig` exists in `shared/sandbox.go` |

**Pain points**:
1. **No in-process MCP tools** → Developers must run external MCP servers even for simple in-memory computations
2. **No lifecycle observability** → Cannot track tool usage, session events, or implement custom analytics
3. **No system prompt customization** → Cannot inject domain-specific instructions without custom CLI args
4. **Verbose permission handling** → Must manually construct `PermissionResult` structs instead of typed helpers
5. **No debug output** → CLI stderr goes to terminal, cannot capture for structured logging

**Impact**:
- **P0 (Critical)**: SDK MCP enables zero-subprocess tool hosting; Hooks enable observability and analytics
- **P1 (High)**: System prompts improve usability; Typed permissions reduce errors; As*Error already done
- **P2 (Medium)**: Debug writers improve troubleshooting
- **P3 (Low)**: Sandbox network already implemented

---

## Proposed Solution

### Architecture Overview

```
┌────────────────────────────────────────────────────────────────┐
│                    agent-sdk-go (Enhanced)                     │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  ┌──────────────────┐         ┌──────────────────┐            │
│  │  claude/         │         │  claude/shared/  │            │
│  │                  │         │                  │            │
│  │  +Client         │────────►│  +BaseOptions    │            │
│  │  +NewTool()      │◄─(P0)─  │    +SystemPrompt │◄─(P1)     │
│  │  +CreateSDK...() │◄─(P0)─  │    +AppendSysP.. │◄─(P1)     │
│  │                  │         │    +DebugWriter  │◄─(P2)     │
│  │                  │         │    +Hooks        │◄─(P0)     │
│  │                  │         │                  │            │
│  └──────────────────┘         └──────────────────┘            │
│         │                              │                       │
│         │                              ▼                       │
│         │                    ┌──────────────────┐             │
│         │                    │ shared/mcp.go    │             │
│         │                    │                  │             │
│         │                    │ +McpSdkServer... │◄─(P0)      │
│         │                    │ +McpTool         │◄─(P0)      │
│         │                    │ +McpToolHandler  │◄─(P0)      │
│         │                    └──────────────────┘             │
│         │                                                      │
│         ▼                                                      │
│  ┌──────────────────────────────────────┐                     │
│  │     subprocess/transport.go          │                     │
│  │                                       │                     │
│  │  - Register SDK MCP servers          │◄─(P0)              │
│  │  - Invoke hook callbacks             │◄─(P0)              │
│  │  - Pass system prompt CLI args       │◄─(P1)              │
│  │  - Pipe stderr to DebugWriter        │◄─(P2)              │
│  └──────────────────────────────────────┘                     │
│                                                                │
│  ┌──────────────────────────────────────┐                     │
│  │  claude/shared/permissions.go        │                     │
│  │                                       │                     │
│  │  +NewPermissionResultAllow()         │◄─(P1)              │
│  │  +NewPermissionResultDeny(msg)       │◄─(P1)              │
│  └──────────────────────────────────────┘                     │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

**Component Changes:**

| Component | Change Type | Description |
|-----------|-------------|-------------|
| `claude/mcp.go` | New | Public API for `NewTool()`, `CreateSDKMcpServer()`, `WithSdkMcpServer()` |
| `claude/shared/mcp.go` | Modified | Add `McpTool`, `McpToolHandler`, `McpSdkServerConfig.Instance` field |
| `claude/shared/hooks.go` | ✅ Exists | Already has 12 hook types, add registration logic |
| `claude/shared/options.go` | Modified | Add `SystemPrompt`, `AppendSystemPrompt`, `DebugWriter`, `Hooks` fields |
| `claude/shared/permissions.go` | Modified | Add `NewPermissionResultAllow()`, `NewPermissionResultDeny()` |
| `claude/subprocess/transport.go` | Modified | Handle SDK MCP, invoke hooks, apply debug writer, pass prompt args |
| `claude/v2/options.go` | Modified | Add `WithSystemPrompt()`, `WithAppendSystemPrompt()`, `WithDebugWriter()`, `WithHooks()` |

---

## Detailed Feature Specifications

### P0-1: SDK MCP Server (In-Process Tool Hosting)

**Goal**: Enable developers to create MCP tools that run in-process (no subprocess) for low-latency, in-memory computations.

**API Surface:**

```go
// claude/mcp.go (NEW FILE)

package claude

// McpToolHandler is invoked when Claude calls the tool.
// Context-first per Go idioms, returns result or error.
type McpToolHandler func(ctx context.Context, args map[string]any) (*McpToolResult, error)

// NewTool creates an MCP tool definition.
// This is the Go alternative to Python's @tool decorator.
//
// Example:
//   addTool := claude.NewTool(
//       "add",
//       "Add two numbers together",
//       map[string]any{
//           "type": "object",
//           "properties": map[string]any{
//               "a": map[string]any{"type": "number"},
//               "b": map[string]any{"type": "number"},
//           },
//           "required": []string{"a", "b"},
//       },
//       func(ctx context.Context, args map[string]any) (*claude.McpToolResult, error) {
//           a, _ := args["a"].(float64)
//           b, _ := args["b"].(float64)
//           return &claude.McpToolResult{
//               Content: []claude.McpContent{{Type: "text", Text: fmt.Sprintf("%f", a+b)}},
//           }, nil
//       },
//   )
func NewTool(name, description string, inputSchema map[string]any, handler McpToolHandler) *McpTool

// CreateSDKMcpServer creates an in-process MCP server with the given tools.
//
// Example:
//   calculator := claude.CreateSDKMcpServer("calculator", "1.0.0", addTool, sqrtTool)
//   client, _ := claude.NewClient(
//       claude.WithSdkMcpServer("calc", calculator),
//       claude.WithAllowedTools("mcp__calc__add", "mcp__calc__sqrt"),
//   )
func CreateSDKMcpServer(name, version string, tools ...*McpTool) *shared.McpSdkServerConfig

// WithSdkMcpServer adds an in-process SDK MCP server.
// Tool names follow format: mcp__<server_name>__<tool_name>
func WithSdkMcpServer(name string, server *shared.McpSdkServerConfig) Option
```

**Internal Implementation (claude/shared/mcp.go):**

```go
// McpTool represents a tool for SDK MCP servers.
type McpTool struct {
    name        string
    description string
    inputSchema map[string]any
    handler     McpToolHandler
}

// SdkMcpServer implements the MCP server interface for in-process tools.
type SdkMcpServer struct {
    name    string
    version string
    mu      sync.RWMutex
    tools   map[string]*McpTool
}

// McpSdkServerConfig (MODIFY EXISTING)
type McpSdkServerConfig struct {
    Type     string `json:"type"` // "sdk"
    Name     string `json:"name"`
    Instance *SdkMcpServer `json:"-"` // NEW: In-process server instance
}
```

**Transport Integration (subprocess/transport.go):**

When subprocess starts, register SDK MCP servers from `opts.McpServers` where `Type == "sdk"`. When CLI requests tool via control protocol, route to `Instance.CallTool()`.

**Success Criteria:**
- [ ] `NewTool()` creates tool with handler
- [ ] `CreateSDKMcpServer()` returns config with Instance
- [ ] `WithSdkMcpServer()` adds server to options
- [ ] Transport routes SDK tool calls to in-process handlers
- [ ] Example from severity1 (`examples/14_sdk_mcp_server/main.go`) ports successfully
- [ ] Tools accessible via `mcp__<name>__<tool>` format

---

### P0-2: Lifecycle Hooks (Observability & Analytics)

**Goal**: Enable developers to observe and react to lifecycle events (tool use, session start/stop, compaction, etc.).

**API Surface:**

```go
// claude/v2/options.go (ADD)

// WithPreToolUseHook registers a callback invoked before tool execution.
// Matcher is a regex pattern (empty = match all tools).
func WithPreToolUseHook(matcher string, callback HookHandler) Option

// WithPostToolUseHook registers a callback invoked after successful tool execution.
func WithPostToolUseHook(matcher string, callback HookHandler) Option

// WithPostToolUseFailureHook registers a callback invoked after failed tool execution.
func WithPostToolUseFailureHook(matcher string, callback HookHandler) Option

// WithSessionStartHook registers a callback invoked when session starts.
func WithSessionStartHook(callback HookHandler) Option

// WithSessionEndHook registers a callback invoked when session ends.
func WithSessionEndHook(callback HookHandler) Option

// WithStopHook registers a callback invoked when session is interrupted.
func WithStopHook(callback HookHandler) Option

// WithSubagentStartHook registers a callback invoked when subagent starts.
func WithSubagentStartHook(callback HookHandler) Option

// WithSubagentStopHook registers a callback invoked when subagent stops.
func WithSubagentStopHook(callback HookHandler) Option

// WithPreCompactHook registers a callback invoked before context compaction.
func WithPreCompactHook(callback HookHandler) Option

// WithPermissionRequestHook registers a callback invoked on permission requests.
func WithPermissionRequestHook(callback HookHandler) Option

// WithNotificationHook registers a callback invoked on CLI notifications.
func WithNotificationHook(callback HookHandler) Option

// WithUserPromptSubmitHook registers a callback invoked when user submits prompt.
func WithUserPromptSubmitHook(callback HookHandler) Option
```

**Hook Handler Signature (claude/shared/hooks.go - already exists):**

```go
// HookHandler is invoked when a hook event fires.
// Context has timeout (default 30s).
// Input is typed to hook event (PreToolUseHookInput, etc.).
type HookHandler func(ctx context.Context, input any) (*SyncHookOutput, error)

// SyncHookOutput controls how CLI proceeds after hook.
type SyncHookOutput struct {
    Continue           bool           `json:"continue,omitempty"`           // Continue execution
    SuppressOutput     bool           `json:"suppressOutput,omitempty"`     // Hide tool output from Claude
    StopReason         string         `json:"stopReason,omitempty"`         // Reason for stopping
    Decision           string         `json:"decision,omitempty"`           // "approve" | "block"
    SystemMessage      string         `json:"systemMessage,omitempty"`      // Message to inject
    Reason             string         `json:"reason,omitempty"`             // Reason for decision
    HookSpecificOutput map[string]any `json:"hookSpecificOutput,omitempty"` // Hook-specific data
}
```

**Transport Integration (subprocess/transport.go):**

Listen for `hook_event` messages from CLI. Match against registered hooks by event type and tool name regex. Invoke handler, send `hook_response` back to CLI.

**Success Criteria:**
- [ ] All 12 hook event types supported
- [ ] Matcher regex filters tool names (PreToolUse, PostToolUse, PostToolUseFailure)
- [ ] Timeout enforced (default 30s, configurable via HookConfig)
- [ ] Hook response sent to CLI via control protocol
- [ ] Example: PreToolUseHook logs every tool call with timestamp
- [ ] Example: PostToolUseFailureHook sends alert on Write tool failures

---

### P1-1: System Prompt Options

**Goal**: Allow developers to inject custom system prompts without manual CLI arg construction.

**API Surface:**

```go
// claude/v2/options.go (ADD)

// WithSystemPrompt sets the system prompt.
// Replaces any existing system prompt.
func WithSystemPrompt(prompt string) Option

// WithAppendSystemPrompt appends to the system prompt.
// Useful for adding domain-specific instructions.
func WithAppendSystemPrompt(prompt string) Option
```

**Internal (claude/shared/options.go - MODIFY):**

```go
type BaseOptions struct {
    // ... existing fields ...
    SystemPrompt       *string `json:"systemPrompt,omitempty"`
    AppendSystemPrompt *string `json:"appendSystemPrompt,omitempty"`
}
```

**Transport Integration (subprocess/transport.go):**

When building CLI args, add:
- `--system-prompt=<value>` if SystemPrompt is set
- `--append-system-prompt=<value>` if AppendSystemPrompt is set

**Success Criteria:**
- [ ] `WithSystemPrompt("Custom instructions")` passes to CLI
- [ ] `WithAppendSystemPrompt("Extra context")` appends to existing
- [ ] Both options work together (base + append)
- [ ] Example: Inject "You are a Go expert" system prompt

---

### P1-2: Typed Permission Results

**Goal**: Provide type-safe constructors for permission callback results instead of manual struct construction.

**API Surface:**

```go
// claude/shared/permissions.go (ADD)

// NewPermissionResultAllow creates a permission result that allows tool execution.
// Optionally modify input or suggest permission updates.
//
// Example:
//   return NewPermissionResultAllow(
//       WithUpdatedInput(map[string]any{"path": "/safe/path"}),
//       WithPermissionUpdates(addRule),
//   ), nil
func NewPermissionResultAllow(opts ...PermissionResultOption) PermissionResult

// NewPermissionResultDeny creates a permission result that denies tool execution.
//
// Example:
//   return NewPermissionResultDeny("Path outside allowed directories"), nil
func NewPermissionResultDeny(message string, opts ...PermissionResultOption) PermissionResult

// PermissionResultOption configures a PermissionResult.
type PermissionResultOption func(*PermissionResult)

// WithUpdatedInput modifies the tool input before execution.
func WithUpdatedInput(input map[string]any) PermissionResultOption

// WithPermissionUpdates suggests permission rule changes.
func WithPermissionUpdates(updates ...PermissionUpdate) PermissionResultOption

// WithInterrupt terminates the session on denial.
func WithInterrupt(interrupt bool) PermissionResultOption
```

**Success Criteria:**
- [ ] `NewPermissionResultAllow()` returns allow behavior
- [ ] `NewPermissionResultDeny("reason")` returns deny behavior with message
- [ ] `WithUpdatedInput()` modifies tool arguments
- [ ] `WithPermissionUpdates()` suggests rule changes
- [ ] Example: Deny Write outside project directory with custom message

---

### P1-3: As*Error Helpers

**Status**: ✅ **Already Implemented** (commit 2906253)

Existing implementation in `claude/shared/errors.go`:
- `AsCLINotFoundError(err error) (*CLINotFoundError, bool)`
- `AsConnectionError(err error) (*ConnectionError, bool)`
- `AsTimeoutError(err error) (*TimeoutError, bool)`
- `AsParserError(err error) (*ParserError, bool)`
- `AsProtocolError(err error) (*ProtocolError, bool)`
- `AsConfigurationError(err error) (*ConfigurationError, bool)`
- `AsProcessError(err error) (*ProcessError, bool)`
- `AsJSONDecodeError(err error) (*JSONDecodeError, bool)`
- `AsMessageParseError(err error) (*MessageParseError, bool)`

**No action required** - feature complete.

---

### P2-1: Debug Writer Options

**Goal**: Allow developers to capture CLI stderr output for structured logging instead of terminal pollution.

**API Surface:**

```go
// claude/v2/options.go (ADD)

// WithDebugWriter redirects CLI stderr to the given writer.
// Useful for structured logging or debug output capture.
//
// Example:
//   var buf bytes.Buffer
//   client, _ := claude.NewClient(claude.WithDebugWriter(&buf))
//   // buf contains all CLI stderr output
func WithDebugWriter(w io.Writer) Option

// WithDebugStderr redirects CLI stderr to os.Stderr (default behavior).
func WithDebugStderr() Option

// WithDebugDisabled suppresses CLI stderr output entirely.
func WithDebugDisabled() Option
```

**Internal (claude/shared/options.go - MODIFY):**

```go
type BaseOptions struct {
    // ... existing fields ...
    DebugWriter io.Writer `json:"-"` // io.Writer for CLI stderr
}
```

**Transport Integration (subprocess/transport.go):**

Set `cmd.Stderr` based on `opts.DebugWriter`:
- If `nil`, suppress (discard)
- If `os.Stderr`, default behavior
- Otherwise, pipe to custom writer

**Success Criteria:**
- [ ] `WithDebugWriter(&buf)` captures all CLI stderr
- [ ] `WithDebugStderr()` sends to terminal
- [ ] `WithDebugDisabled()` suppresses output
- [ ] Example: Capture debug output to structured logger (zerolog/zap)

---

### P3-1: Sandbox Network Config

**Status**: ✅ **Already Implemented**

Existing implementation in `claude/shared/sandbox.go`:
```go
type SandboxNetworkConfig struct {
    AllowedDomains      []string `json:"allowedDomains,omitempty"`
    AllowUnixSockets    []string `json:"allowUnixSockets,omitempty"`
    AllowAllUnixSockets bool     `json:"allowAllUnixSockets,omitempty"`
    AllowLocalBinding   bool     `json:"allowLocalBinding,omitempty"`
    HttpProxyPort       int      `json:"httpProxyPort,omitempty"`
    SocksProxyPort      int      `json:"socksProxyPort,omitempty"`
}

type SandboxConfig struct {
    Network *SandboxNetworkConfig `json:"network,omitempty"`
    // ... other fields ...
}
```

**No action required** - feature complete.

---

## Implementation Tasks

### Milestone 1: P0 - SDK MCP Server (Critical)

| ID | Task | Effort | Inputs | Outputs |
|----|------|--------|--------|---------|
| M1.1 | Create `claude/mcp.go` with public API | 1h | severity1 mcp.go | `claude/mcp.go` |
| M1.2 | Add `McpTool`, `McpToolHandler` to `shared/mcp.go` | 1h | M1.1 | `claude/shared/mcp.go` |
| M1.3 | Modify `McpSdkServerConfig` to include `Instance` field | 30m | M1.2 | `claude/shared/mcp.go` |
| M1.4 | Implement `NewTool()`, `CreateSDKMcpServer()` | 2h | M1.3 | `claude/mcp.go` |
| M1.5 | Add `WithSdkMcpServer()` option function | 30m | M1.4 | `claude/options.go` |
| M1.6 | Update transport to register SDK MCP servers | 3h | M1.5 | `subprocess/transport.go` |
| M1.7 | Implement SDK tool routing in control protocol | 2h | M1.6 | `subprocess/control.go` |
| M1.8 | Write tests for SDK MCP functionality | 2h | M1.7 | `claude/mcp_test.go` |
| M1.9 | Port `examples/14_sdk_mcp_server/main.go` | 1h | M1.8 | `examples/sdk_mcp_server/` |

**Milestone Total**: ~12 hours

---

### Milestone 2: P0 - Lifecycle Hooks (Critical)

| ID | Task | Effort | Inputs | Outputs |
|----|------|--------|--------|---------|
| M2.1 | Add hook registration to `BaseOptions` | 30m | - | `claude/shared/options.go` |
| M2.2 | Implement `WithPreToolUseHook()` and siblings | 1h | M2.1 | `claude/v2/options.go` |
| M2.3 | Add hook registry to transport | 1h | M2.2 | `subprocess/transport.go` |
| M2.4 | Implement hook event listener in transport | 2h | M2.3 | `subprocess/transport.go` |
| M2.5 | Implement hook matcher regex logic | 1h | M2.4 | `claude/shared/hooks.go` |
| M2.6 | Send `hook_response` to CLI via control protocol | 2h | M2.5 | `subprocess/control.go` |
| M2.7 | Write tests for all 12 hook types | 3h | M2.6 | `claude/shared/hooks_test.go` |
| M2.8 | Create example: PreToolUse analytics logger | 1h | M2.7 | `examples/hooks_analytics/` |

**Milestone Total**: ~11.5 hours

---

### Milestone 3: P1 - System Prompts & Permissions (High Priority)

| ID | Task | Effort | Inputs | Outputs |
|----|------|--------|--------|---------|
| M3.1 | Add `SystemPrompt`, `AppendSystemPrompt` to `BaseOptions` | 15m | - | `claude/shared/options.go` |
| M3.2 | Implement `WithSystemPrompt()`, `WithAppendSystemPrompt()` | 30m | M3.1 | `claude/v2/options.go` |
| M3.3 | Pass system prompt args to CLI in transport | 30m | M3.2 | `subprocess/transport.go` |
| M3.4 | Add `NewPermissionResultAllow()`, `NewPermissionResultDeny()` | 1h | - | `claude/shared/permissions.go` |
| M3.5 | Implement `PermissionResultOption` pattern | 1h | M3.4 | `claude/shared/permissions.go` |
| M3.6 | Write tests for system prompt options | 30m | M3.3 | `claude/v2/options_test.go` |
| M3.7 | Write tests for typed permission results | 1h | M3.5 | `claude/shared/permissions_test.go` |
| M3.8 | Create example: Custom system prompt | 30m | M3.6 | `examples/system_prompt/` |

**Milestone Total**: ~5.5 hours

---

### Milestone 4: P2 - Debug Writer (Medium Priority)

| ID | Task | Effort | Inputs | Outputs |
|----|------|--------|--------|---------|
| M4.1 | Add `DebugWriter` to `BaseOptions` | 15m | - | `claude/shared/options.go` |
| M4.2 | Implement `WithDebugWriter()`, `WithDebugStderr()`, `WithDebugDisabled()` | 30m | M4.1 | `claude/v2/options.go` |
| M4.3 | Apply debug writer to `cmd.Stderr` in transport | 30m | M4.2 | `subprocess/transport.go` |
| M4.4 | Write tests for debug writer | 45m | M4.3 | `claude/v2/options_test.go` |
| M4.5 | Create example: Capture debug to structured logger | 30m | M4.4 | `examples/debug_logging/` |

**Milestone Total**: ~2.5 hours

---

### Milestone 5: Documentation & Integration

| ID | Task | Effort | Inputs | Outputs |
|----|------|--------|--------|---------|
| M5.1 | Update README with SDK MCP section | 30m | M1.9 | `README.md` |
| M5.2 | Update README with hooks section | 30m | M2.8 | `README.md` |
| M5.3 | Document system prompt options | 15m | M3.8 | `docs/usage.md` |
| M5.4 | Document permission result helpers | 15m | M3.7 | `docs/usage.md` |
| M5.5 | Document debug writer options | 15m | M4.5 | `docs/usage.md` |
| M5.6 | Create migration guide from severity1 to dotcommander | 1h | All | `docs/MIGRATION-SEVERITY1.md` |
| M5.7 | Update CLAUDE.md with new patterns | 30m | All | `CLAUDE.md` |

**Milestone Total**: ~3.5 hours

---

## Total Effort Summary

| Priority | Milestones | Total Effort |
|----------|------------|--------------|
| P0 | M1 (SDK MCP), M2 (Hooks) | ~23.5 hours (~3 days) |
| P1 | M3 (Prompts, Permissions) | ~5.5 hours (~1 day) |
| P2 | M4 (Debug Writer) | ~2.5 hours |
| Docs | M5 (Documentation) | ~3.5 hours |
| **Total** | **5 milestones** | **~35 hours (~4.5 days)** |

**Note**: P1.3 (As*Error) and P3.1 (Sandbox Network) already complete, reducing effort by ~3-4 hours.

---

## Verification Matrix

| Feature | Checkpoint | Command | Expected |
|---------|-----------|---------|----------|
| **SDK MCP** | Server creation | `go test -run TestCreateSDKMcpServer ./claude` | PASS |
| **SDK MCP** | Tool registration | `go test -run TestNewTool ./claude` | PASS |
| **SDK MCP** | End-to-end | `go run examples/sdk_mcp_server/main.go` | Calculator tools execute |
| **Hooks** | Registration | `go test -run TestHookRegistration ./claude/shared` | PASS |
| **Hooks** | PreToolUse | `go test -run TestPreToolUseHook ./claude/shared` | Callback invoked |
| **Hooks** | Matcher regex | `go test -run TestHookMatcher ./claude/shared` | Tool filter works |
| **System Prompt** | Option parsing | `go test -run TestSystemPromptOption ./claude/v2` | PASS |
| **System Prompt** | CLI arg | `go test -run TestSystemPromptCLI ./subprocess` | `--system-prompt` passed |
| **Permissions** | Allow result | `go test -run TestPermissionResultAllow ./claude/shared` | Correct behavior |
| **Permissions** | Deny result | `go test -run TestPermissionResultDeny ./claude/shared` | Correct behavior + message |
| **Debug Writer** | Capture stderr | `go test -run TestDebugWriter ./claude/v2` | Buffer captures output |
| **Debug Writer** | Disable | `go test -run TestDebugDisabled ./claude/v2` | No output |

---

## Rollback Procedure

If any feature causes regressions:

1. **Revert commits by milestone**:
   ```bash
   # Identify problematic milestone
   git log --oneline --grep="M[1-5]"

   # Revert milestone commits in reverse order
   git revert <commit-hash-M1.9>..<commit-hash-M1.1>
   ```

2. **Feature flags** (if partial rollback needed):
   ```go
   // claude/shared/features.go
   const (
       FeatureSDKMcp = "sdk_mcp"
       FeatureHooks  = "hooks"
   )

   var enabledFeatures = map[string]bool{
       FeatureSDKMcp: false, // Disable via flag
       FeatureHooks:  true,
   }
   ```

3. **Compatibility shim** (for external users):
   - Export deprecated API as aliases to new API
   - Add deprecation notices in godoc
   - Remove in next major version

---

## Migration from severity1/claude-agent-sdk-go

For developers using severity1's fork who want to migrate to dotcommander:

### Import Path Changes

```diff
- import "github.com/severity1/claude-agent-sdk-go"
+ import "github.com/dotcommander/agent-sdk-go/claude"
```

### API Mapping

| severity1 API | dotcommander API | Notes |
|---------------|------------------|-------|
| `claudecode.NewTool()` | `claude.NewTool()` | Identical signature |
| `claudecode.CreateSDKMcpServer()` | `claude.CreateSDKMcpServer()` | Identical signature |
| `claudecode.WithSdkMcpServer()` | `claude.WithSdkMcpServer()` | Identical signature |
| `claudecode.WithSystemPrompt()` | `claude.v2.WithSystemPrompt()` | Identical signature |
| `claudecode.WithAppendSystemPrompt()` | `claude.v2.WithAppendSystemPrompt()` | Identical signature |
| `claudecode.WithDebugWriter()` | `claude.v2.WithDebugWriter()` | Identical signature |
| `claudecode.NewPermissionResultAllow()` | `claude.NewPermissionResultAllow()` | Identical signature |
| `claudecode.NewPermissionResultDeny()` | `claude.NewPermissionResultDeny()` | Identical signature |

### Code Example Migration

**Before (severity1):**
```go
import claudecode "github.com/severity1/claude-agent-sdk-go"

calculator := claudecode.CreateSDKMcpServer("calc", "1.0", addTool)
client := claudecode.NewClient(
    claudecode.WithSdkMcpServer("calc", calculator),
    claudecode.WithSystemPrompt("You are a calculator"),
)
```

**After (dotcommander):**
```go
import "github.com/dotcommander/agent-sdk-go/claude"
import "github.com/dotcommander/agent-sdk-go/claude/v2"

calculator := claude.CreateSDKMcpServer("calc", "1.0", addTool)
client, _ := v2.NewSession(ctx,
    claude.WithSdkMcpServer("calc", calculator),
    v2.WithSystemPrompt("You are a calculator"),
)
```

---

## Gotchas & Limitations

### SDK MCP Server

**Gotcha**: Tool names must follow `mcp__<server>__<tool>` format in `AllowedTools`
```go
// ❌ Wrong
claude.WithAllowedTools("add", "sqrt")

// ✅ Correct
claude.WithAllowedTools("mcp__calc__add", "mcp__calc__sqrt")
```

**Limitation**: SDK MCP tools run in-process → no isolation from main process. If tool handler panics, entire process crashes. Use `recover()` in handlers for production.

### Lifecycle Hooks

**Gotcha**: Hook timeout defaults to 30s. Long-running analytics should be async:
```go
// ❌ Blocks session for 60s
WithPreToolUseHook("", func(ctx context.Context, input any) (*SyncHookOutput, error) {
    time.Sleep(60 * time.Second) // BAD
    return &SyncHookOutput{Continue: true}, nil
})

// ✅ Non-blocking
WithPreToolUseHook("", func(ctx context.Context, input any) (*SyncHookOutput, error) {
    go sendAnalyticsAsync(input) // Good - async
    return &SyncHookOutput{Continue: true}, nil
})
```

**Limitation**: Hooks are NOT persisted across session resume. Re-register hooks when resuming.

### System Prompts

**Gotcha**: `WithSystemPrompt()` replaces, `WithAppendSystemPrompt()` appends. Order matters:
```go
// ❌ Wrong order - append is ignored
claude.WithAppendSystemPrompt("Extra context"),
claude.WithSystemPrompt("Base prompt"), // Replaces everything

// ✅ Correct order
claude.WithSystemPrompt("Base prompt"),
claude.WithAppendSystemPrompt("Extra context"), // Appends to base
```

### Debug Writer

**Gotcha**: `io.Writer` must be thread-safe. CLI writes stderr concurrently:
```go
// ❌ Not thread-safe
var buf bytes.Buffer
claude.WithDebugWriter(&buf) // Data races!

// ✅ Thread-safe
type SafeBuffer struct {
    mu sync.Mutex
    buf bytes.Buffer
}
func (s *SafeBuffer) Write(p []byte) (int, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    return s.buf.Write(p)
}
```

---

## Next Steps

1. **Prioritize P0 features** (SDK MCP, Hooks) → High impact, enables new workflows
2. **Implement P1 features** (System Prompts, Permissions) → Quick wins, improves DX
3. **Defer P2-P3** until after P0-P1 complete → Lower priority
4. **Create examples for each feature** → Documentation by demonstration
5. **Run integration tests** → Ensure no regressions in existing functionality
6. **Update CHANGELOG.md** → Document all new features

---

## References

- **severity1 SDK**: `/tmp/severity1-sdk/` (cloned for reference)
- **Existing spec**: `SPEC-TYPESCRIPT-PARITY.md` (partial overlap)
- **Hook types**: `claude/shared/hooks.go` (already implemented)
- **Sandbox config**: `claude/shared/sandbox.go` (already implemented)
- **Error helpers**: `claude/shared/errors.go` (already implemented)
- **TypeScript SDK**: https://github.com/anthropics/anthropic-sdk-typescript (parity reference)
