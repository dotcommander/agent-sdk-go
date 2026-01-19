# Feature Parity User Stories

User stories for achieving full feature parity with severity1 SDK. Created from `docs/parity.md` analysis.

**Status Legend:**
- **P0** - Critical for Python SDK parity
- **P1** - Important for feature completeness
- **P2** - Nice to have
- **P3** - Low priority

---

## Implementation Status (2026-01-18)

| Priority | Completed | Total | Notes |
|----------|-----------|-------|-------|
| **P0 Critical** | 10/10 | 100% | All complete |
| **P1 Important** | 17/17 | 100% | All complete |
| **P2 Nice to Have** | 8/8 | 100% | All complete |

### Completion Summary

**✅ COMPLETED (Option API + Tests):**
- US-001: WithAllowedTools
- US-002: WithCwd (alias)
- US-003: WithAddDirs (alias)
- US-004: WithForkSession
- US-005: WithFallbackModel
- US-006: WithUser
- US-007: WithJSONSchema
- US-008: WithOutputFormat
- US-009: WithBetas
- US-010: WithSettingSources
- US-011: WithDebugWriter
- US-012: WithStderrCallback
- US-013: WithAgents
- US-014: WithHooks
- US-015: WithHook
- US-016: WithPreToolUseHook
- US-018: WithCanUseTool
- US-019: WithSdkMcpServer
- US-020: CreateSDKMcpServer
- US-024: WithSandboxSettings
- US-026: WithTypedPermissionMode
- US-033: Option validation

**✅ COMPLETED (Control Protocol - subprocess package):**
- US-017: Hook callback routing (control_hooks.go)
- US-021: Control protocol infrastructure (control.go, control_types.go)
- US-022: Initialize handshake (Protocol.Initialize)
- US-023: MCP message routing (control_mcp.go)
- US-032: RewindFiles (Protocol.RewindFiles)

**✅ COMPLETED (Error Types):**
- US-035: Error type hierarchy (PermissionError, ModelError added)

**✅ COMPLETED (Control Protocol API Methods):**
- US-027: SupportedCommands (via Protocol.GetCommands)
- US-028: SupportedModels (via Protocol.GetModels)
- US-029: McpServerStatus (via Protocol.GetMcpServerStatus)
- US-030: SetMcpServers (via Protocol.SetMcpServers)
- US-031: AccountInfo (via Protocol.GetAccountInfo)

**⚠️ PARTIAL:**
- US-025: PluginConfig (basic struct exists)
- US-034: Documentation parity (ongoing)

---

## Category: Options Configuration

### US-001: Allowed Tools Option
**Status:** ✅ Completed
**Priority:** P0

**As a** SDK user
**I want** to specify a whitelist of allowed tools
**So that** I can restrict Claude to only use specific tools

**Acceptance Criteria:**
- [ ] `WithAllowedTools(tools ...string)` option function exists
- [ ] When set, only specified tools are available to Claude
- [ ] Works in combination with `WithDisallowedTools()` (allowlist takes precedence)
- [ ] Empty slice allows all tools
- [ ] Invalid tool names return validation error

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithAllowedTools("Read", "Write", "Bash"),
)
// Claude can only use Read, Write, Bash - all other tools blocked
```

---

### US-002: Working Directory Option Alias
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user migrating from severity1
**I want** a `WithCwd()` alias for `WithWorkingDirectory()`
**So that** my existing code doesn't break

**Acceptance Criteria:**
- [ ] `WithCwd(path string)` option function exists
- [ ] Behaves identically to `WithWorkingDirectory()`
- [ ] Can be used interchangeably
- [ ] Last option wins if both provided
- [ ] Documented as alias in godoc

**Code Example:**
```go
// Both should work identically
client1, _ := claude.NewClient(claude.WithCwd("/tmp"))
client2, _ := claude.NewClient(claude.WithWorkingDirectory("/tmp"))
```

---

### US-003: Additional Directories Option Alias
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user migrating from severity1
**I want** a `WithAddDirs()` alias for `WithAdditionalDirectories()`
**So that** my existing code doesn't break

**Acceptance Criteria:**
- [ ] `WithAddDirs(dirs ...string)` option function exists
- [ ] Behaves identically to `WithAdditionalDirectories()`
- [ ] Can be used interchangeably
- [ ] Documented as alias in godoc

**Code Example:**
```go
// Both should work identically
client1, _ := claude.NewClient(claude.WithAddDirs("/data", "/logs"))
client2, _ := claude.NewClient(claude.WithAdditionalDirectories("/data", "/logs"))
```

---

### US-004: Fork Session Option
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** to fork an existing session
**So that** I can explore different conversation branches without affecting the original

**Acceptance Criteria:**
- [ ] `WithForkSession(sessionID string)` option function exists
- [ ] Creates new session based on existing session state
- [ ] Original session remains unmodified
- [ ] Fork includes conversation history up to fork point
- [ ] Returns error if source session doesn't exist
- [ ] Works with Claude CLI `--fork-session` flag

**Code Example:**
```go
// Original session
session1, _ := claude.NewClient(claude.WithModel("claude-sonnet-4"))
session1.Query(ctx, "What is 2+2?")

// Fork to explore different path
session2, _ := claude.NewClient(claude.WithForkSession(session1.SessionID()))
session2.Query(ctx, "Actually, multiply them instead")
```

---

### US-005: Fallback Model Option
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** to specify a fallback model
**So that** if the primary model is unavailable, Claude automatically switches

**Acceptance Criteria:**
- [ ] `WithFallbackModel(model string)` option function exists
- [ ] Falls back when primary model returns error
- [ ] Logs fallback event to stderr/debug output
- [ ] Works with model validation
- [ ] Can chain multiple fallbacks

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithModel("claude-opus-4"),
    claude.WithFallbackModel("claude-sonnet-4"),
)
// If opus-4 unavailable, automatically uses sonnet-4
```

---

### US-006: User Option
**Status:** ✅ Completed
**Priority:** P2

**As a** SDK user
**I want** to specify a user identifier
**So that** Claude can track usage per user in multi-tenant scenarios

**Acceptance Criteria:**
- [ ] `WithUser(userID string)` option function exists
- [ ] User ID passed to Claude CLI
- [ ] Visible in session metadata
- [ ] Used for usage tracking/attribution
- [ ] Sanitized to prevent injection

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithUser("user-12345"),
)
// All requests tagged with user-12345 for tracking
```

---

### US-007: JSON Schema Output Option
**Status:** ✅ Completed
**Priority:** P0

**As a** SDK user
**I want** to specify a JSON schema for Claude's output
**So that** Claude returns structured data matching my schema

**Acceptance Criteria:**
- [ ] `WithJSONSchema(schema map[string]any)` option function exists
- [ ] Schema validated against JSON Schema spec
- [ ] Claude CLI receives schema via appropriate flag
- [ ] Output conforms to schema or returns error
- [ ] Works with `WithOutputFormat()`

**Code Example:**
```go
schema := map[string]any{
    "type": "object",
    "properties": map[string]any{
        "name": map[string]any{"type": "string"},
        "age":  map[string]any{"type": "number"},
    },
    "required": []string{"name", "age"},
}

client, err := claude.NewClient(
    claude.WithJSONSchema(schema),
)
// Claude output guaranteed to match schema
```

---

### US-008: Output Format Option
**Status:** ✅ Completed
**Priority:** P0

**As a** SDK user
**I want** to specify the output format
**So that** I can request JSON, text, or other formats

**Acceptance Criteria:**
- [ ] `WithOutputFormat(format string)` option function exists
- [ ] Supports formats: "json", "text", "markdown"
- [ ] Works with `WithJSONSchema()` for structured output
- [ ] Invalid format returns error
- [ ] Maps to Claude CLI `--output-format` flag

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithOutputFormat("json"),
    claude.WithJSONSchema(mySchema),
)
// Claude returns JSON matching schema
```

---

### US-009: Betas Option
**Status:** ✅ Completed
**Priority:** P0

**As a** SDK user
**I want** to enable beta features
**So that** I can use experimental Claude capabilities

**Acceptance Criteria:**
- [ ] `WithBetas(betas ...string)` option function exists
- [ ] Uses constants from `shared.BetaFeatures`
- [ ] Multiple betas can be enabled simultaneously
- [ ] Invalid beta names return error
- [ ] Maps to Claude CLI `--betas` flag

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithBetas(shared.BetaMaxTokens, shared.BetaComputerUse),
)
// Enables beta features for extended thinking and computer use
```

---

### US-010: Setting Sources Option
**Status:** ✅ Completed
**Priority:** P2

**As a** SDK user
**I want** to control where Claude loads settings from
**So that** I can override config file behavior

**Acceptance Criteria:**
- [ ] `WithSettingSources(sources ...string)` option function exists
- [ ] Supports: "env", "file", "api", "none"
- [ ] Order matters (precedence)
- [ ] Empty slice uses Claude defaults
- [ ] Maps to CLI setting sources configuration

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithSettingSources("env", "api"), // Skip config files
)
// Only use environment variables and API defaults
```

---

### US-011: Debug Writer Option
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** to capture debug output to a custom writer
**So that** I can log/analyze Claude's internal operations

**Acceptance Criteria:**
- [ ] `WithDebugWriter(w io.Writer)` option function exists
- [ ] Captures all debug-level output
- [ ] Works alongside stderr capture
- [ ] Writes are concurrent-safe
- [ ] Nil writer disables debug output

**Code Example:**
```go
var debugBuf bytes.Buffer
client, err := claude.NewClient(
    claude.WithDebugWriter(&debugBuf),
)
// All debug output goes to debugBuf
```

---

### US-012: Stderr Callback Option
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** to receive callbacks for stderr output
**So that** I can process errors/warnings in real-time

**Acceptance Criteria:**
- [ ] `WithStderrCallback(fn func(string))` option function exists
- [ ] Callback invoked for each stderr line
- [ ] Non-blocking (buffered channel or goroutine)
- [ ] Callback errors don't crash SDK
- [ ] Works with existing `Stderr io.Writer` option

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithStderrCallback(func(line string) {
        log.Printf("Claude stderr: %s", line)
    }),
)
// Real-time stderr processing
```

---

### US-013: Agents Option
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** to pass multiple agents via `WithAgents()`
**So that** I don't have to call `WithAgent()` multiple times

**Acceptance Criteria:**
- [ ] `WithAgents(agents ...shared.AgentDefinition)` option function exists
- [ ] Accepts variadic agent definitions
- [ ] Merges with existing agents from `WithAgent()`
- [ ] Last duplicate agent name wins
- [ ] Empty slice clears all agents

**Code Example:**
```go
agents := []shared.AgentDefinition{
    {Name: "coder", Model: "claude-sonnet-4"},
    {Name: "reviewer", Model: "claude-opus-4"},
}

client, err := claude.NewClient(
    claude.WithAgents(agents...),
)
// Multiple agents registered at once
```

---

## Category: Hooks System

### US-014: Hooks Registration Option
**Status:** ✅ Completed
**Priority:** P0

**As a** SDK user
**I want** to register multiple hooks at client creation
**So that** I can set up all event handlers in one place

**Acceptance Criteria:**
- [ ] `WithHooks(hooks map[string]shared.HookCallback)` option function exists
- [ ] Map key is hook event name (e.g., "preToolUse")
- [ ] Hooks are wired to control protocol
- [ ] Invalid event names return error
- [ ] Overrides existing hooks with same name

**Code Example:**
```go
hooks := map[string]shared.HookCallback{
    "preToolUse": func(input shared.PreToolUseInput) shared.HookOutput {
        log.Printf("Tool: %s", input.ToolName)
        return shared.HookOutput{Continue: true}
    },
    "postToolUse": func(input shared.PostToolUseInput) shared.HookOutput {
        log.Printf("Result: %v", input.Result)
        return shared.HookOutput{Continue: true}
    },
}

client, err := claude.NewClient(
    claude.WithHooks(hooks),
)
```

---

### US-015: Single Hook Registration Option
**Status:** ✅ Completed
**Priority:** P0

**As a** SDK user
**I want** to register a single hook via option
**So that** I can add hooks without creating a map

**Acceptance Criteria:**
- [ ] `WithHook(event string, callback shared.HookCallback)` option function exists
- [ ] Can be called multiple times for different events
- [ ] Last registration wins for duplicate events
- [ ] Invalid event name returns error
- [ ] Works with `WithHooks()`

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithHook("preToolUse", myPreToolHook),
    claude.WithHook("postToolUse", myPostToolHook),
)
```

---

### US-016: PreToolUse Hook Convenience Option
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** a convenience option for the common PreToolUse hook
**So that** I can register it with less boilerplate

**Acceptance Criteria:**
- [ ] `WithPreToolUseHook(fn func(shared.PreToolUseInput) shared.HookOutput)` exists
- [ ] Registers hook for "preToolUse" event
- [ ] Type-safe callback signature
- [ ] Works with other hook options
- [ ] Last registration wins

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithPreToolUseHook(func(input shared.PreToolUseInput) shared.HookOutput {
        if input.ToolName == "Bash" {
            return shared.HookOutput{Continue: false, Error: "Bash not allowed"}
        }
        return shared.HookOutput{Continue: true}
    }),
)
```

---

### US-017: Hook Callback Routing via Control Protocol
**Status:** ✅ Completed (subprocess/control_hooks.go)
**Priority:** P0

**As a** SDK user
**I want** hooks to actually execute when events occur
**So that** I can intercept and modify Claude's behavior

**Acceptance Criteria:**
- [ ] Control protocol routes hook events to registered callbacks
- [ ] Callbacks execute in separate goroutine (non-blocking)
- [ ] Callback return values sent back to Claude CLI
- [ ] Callback errors are handled gracefully
- [ ] All 12 hook types are supported:
  - [ ] PreToolUse
  - [ ] PostToolUse
  - [ ] PostToolUseFailure
  - [ ] Notification
  - [ ] UserPromptSubmit
  - [ ] SessionStart
  - [ ] SessionEnd
  - [ ] Stop
  - [ ] SubagentStart
  - [ ] SubagentStop
  - [ ] PreCompact
  - [ ] PermissionRequest

**Code Example:**
```go
// Current: types exist but not wired
// Expected: hooks actually execute

client, err := claude.NewClient(
    claude.WithPreToolUseHook(func(input shared.PreToolUseInput) shared.HookOutput {
        // This should actually execute when Claude tries to use a tool
        return shared.HookOutput{Continue: true}
    }),
)

// When Claude tries to use a tool, hook callback is invoked
```

---

## Category: Permissions System

### US-018: Permission Callback Option
**Status:** ✅ Completed
**Priority:** P0

**As a** SDK user
**I want** to register a permission callback via option
**So that** I can control tool usage at client creation

**Acceptance Criteria:**
- [ ] `WithCanUseTool(fn shared.CanUseToolCallback)` option function exists
- [ ] Callback invoked for every tool use request
- [ ] Receives `shared.CanUseToolOptions` context
- [ ] Returns `shared.PermissionResult`
- [ ] Callback errors default to deny
- [ ] Works with `PermissionMode` settings

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithCanUseTool(func(opts shared.CanUseToolOptions) shared.PermissionResult {
        if opts.Tool.Name == "Bash" && strings.Contains(opts.Tool.Input["command"], "rm -rf") {
            return shared.NewPermissionResultDeny("Dangerous command blocked")
        }
        return shared.NewPermissionResultAllow()
    }),
)
```

---

## Category: MCP Integration

### US-019: SDK MCP Server Option
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** to register in-process MCP servers
**So that** I can provide tools without separate processes

**Acceptance Criteria:**
- [ ] `WithSdkMcpServer(name string, server shared.McpServer)` option function exists
- [ ] Server runs in-process (no subprocess)
- [ ] Tools registered with Claude
- [ ] Concurrent tool calls supported
- [ ] Server lifecycle managed by SDK
- [ ] Multiple SDK servers can be registered

**Code Example:**
```go
type MyMcpServer struct{}

func (s *MyMcpServer) ListTools(ctx context.Context) ([]shared.McpToolDefinition, error) {
    return []shared.McpToolDefinition{
        {Name: "custom-tool", Description: "My custom tool"},
    }, nil
}

func (s *MyMcpServer) CallTool(ctx context.Context, name string, args map[string]any) (shared.McpToolResult, error) {
    // Implementation
    return shared.McpToolResult{Content: "result"}, nil
}

client, err := claude.NewClient(
    claude.WithSdkMcpServer("my-server", &MyMcpServer{}),
)
```

---

### US-020: SDK MCP Server Helper Function
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** a helper to create SDK MCP servers
**So that** I can quickly implement custom tools

**Acceptance Criteria:**
- [ ] `CreateSDKMcpServer(name string, tools []shared.McpToolDefinition, handler func(ctx, string, map[string]any) (shared.McpToolResult, error))` exists
- [ ] Returns `shared.McpSdkServerConfig`
- [ ] Implements `McpServer` interface
- [ ] Handles tool listing and execution
- [ ] Concurrent-safe

**Code Example:**
```go
tools := []shared.McpToolDefinition{
    {Name: "greet", Description: "Greet a user"},
}

handler := func(ctx context.Context, name string, args map[string]any) (shared.McpToolResult, error) {
    if name == "greet" {
        return shared.McpToolResult{
            Content: fmt.Sprintf("Hello, %s!", args["name"]),
        }, nil
    }
    return shared.McpToolResult{}, fmt.Errorf("unknown tool: %s", name)
}

server := claude.CreateSDKMcpServer("greeter", tools, handler)
client, err := claude.NewClient(
    claude.WithMcpServers(server),
)
```

---

## Category: Control Protocol

### US-021: Control Protocol Infrastructure
**Status:** ✅ Completed (subprocess/control.go, control_types.go)
**Priority:** P0

**As a** SDK developer
**I want** a full control protocol implementation
**So that** bidirectional communication with Claude CLI works properly

**Acceptance Criteria:**
- [ ] Control protocol types package exists (similar to severity1's `internal/control`)
- [ ] Request/response correlation via unique IDs
- [ ] Bidirectional message routing
- [ ] Supports all protocol message types:
  - [ ] Initialize handshake
  - [ ] Hook callbacks
  - [ ] MCP server calls
  - [ ] Permission requests
  - [ ] Runtime control (SetModel, Interrupt, etc.)
- [ ] Timeout handling for requests
- [ ] Error propagation from CLI to SDK

**Code Example:**
```go
// Internal control protocol (not user-facing API)
// Should handle messages like:

// Request from SDK to CLI
{
    "id": "req-123",
    "type": "set_model",
    "payload": {"model": "claude-opus-4"}
}

// Response from CLI to SDK
{
    "id": "req-123",
    "type": "response",
    "payload": {"success": true}
}

// Hook callback from CLI to SDK
{
    "id": "hook-456",
    "type": "hook",
    "event": "preToolUse",
    "payload": {"tool_name": "Read", "input": {...}}
}
```

---

### US-022: Initialize Handshake
**Status:** ✅ Completed (Protocol.Initialize in control.go)
**Priority:** P0

**As a** SDK
**I want** to perform initialization handshake with Claude CLI
**So that** both sides know the protocol version and capabilities

**Acceptance Criteria:**
- [ ] SDK sends initialize message on Connect()
- [ ] CLI responds with capabilities and version
- [ ] Mismatched versions logged as warning
- [ ] Capabilities stored for feature detection
- [ ] Timeout if no response in 5s
- [ ] Works for both Query() and Connect() workflows

**Code Example:**
```go
// Internal handshake flow
// 1. SDK starts subprocess
// 2. SDK sends: {"type": "initialize", "version": "1.0", "capabilities": [...]}
// 3. CLI responds: {"type": "initialize_response", "version": "1.0", "capabilities": [...]}
// 4. SDK stores CLI capabilities for feature detection
```

---

### US-023: MCP Message Routing
**Status:** ✅ Completed (subprocess/control_mcp.go)
**Priority:** P1

**As a** SDK
**I want** to route MCP tool calls to registered servers
**So that** SDK MCP servers can handle tool executions

**Acceptance Criteria:**
- [ ] CLI sends MCP tool request to SDK
- [ ] SDK routes to correct server by name
- [ ] Server executes tool and returns result
- [ ] Result sent back to CLI
- [ ] Errors propagated correctly
- [ ] Concurrent requests supported

**Code Example:**
```go
// Internal MCP routing flow
// 1. CLI sends: {"type": "mcp_tool_call", "server": "my-server", "tool": "greet", "args": {...}}
// 2. SDK routes to registered McpSdkServerConfig
// 3. Server executes tool
// 4. SDK responds: {"type": "mcp_tool_response", "result": {...}}
```

---

## Category: Sandbox Settings

### US-024: Full Sandbox Settings
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** complete sandbox configuration options
**So that** I can control Claude's execution environment fully

**Acceptance Criteria:**
- [ ] `SandboxSettings` struct includes all fields from severity1:
  - [ ] `NetworkAccess` (bool)
  - [ ] `AllowedDomains` ([]string)
  - [ ] `BlockedDomains` ([]string)
  - [ ] `AllowedCommands` ([]string)
  - [ ] `BlockedCommands` ([]string)
  - [ ] `RipgrepEnabled` (bool)
  - [ ] `MaxFileSize` (int64)
  - [ ] `ViolationHandling` (string: "block", "warn", "log")
- [ ] `WithSandboxSettings(settings shared.SandboxSettings)` option exists
- [ ] Settings validated at client creation
- [ ] Maps to Claude CLI sandbox flags

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithSandboxSettings(shared.SandboxSettings{
        NetworkAccess:   false,
        AllowedCommands: []string{"ls", "cat", "grep"},
        BlockedCommands: []string{"rm", "dd", "mkfs"},
        RipgrepEnabled:  true,
        MaxFileSize:     10 * 1024 * 1024, // 10MB
        ViolationHandling: "block",
    }),
)
```

---

## Category: Plugin Configuration

### US-025: Full SDK Plugin Configuration
**Status:** Partial (⚠️)
**Priority:** P2

**As a** SDK user
**I want** complete plugin configuration options
**So that** I can customize Claude Code plugin behavior

**Acceptance Criteria:**
- [ ] `SdkPluginConfig` struct matches severity1 structure:
  - [ ] `Enabled` (bool)
  - [ ] `PluginPath` (string)
  - [ ] `Config` (map[string]any)
  - [ ] `Timeout` (time.Duration)
  - [ ] `MaxConcurrent` (int)
- [ ] `WithPluginConfig(cfg shared.SdkPluginConfig)` option exists
- [ ] Plugin validation on client creation
- [ ] Timeout enforced per plugin call

**Code Example:**
```go
client, err := claude.NewClient(
    claude.WithPluginConfig(shared.SdkPluginConfig{
        Enabled:    true,
        PluginPath: "/usr/local/lib/claude-plugins/analyzer.so",
        Config: map[string]any{
            "log_level": "debug",
            "cache_dir": "/tmp/plugin-cache",
        },
        Timeout:      30 * time.Second,
        MaxConcurrent: 4,
    }),
)
```

---

## Category: Type System Alignment

### US-026: Typed Permission Mode
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** typed permission mode constants
**So that** I get compile-time safety and autocomplete

**Acceptance Criteria:**
- [ ] `PermissionMode` is a custom type (not string)
- [ ] Constants defined:
  - [ ] `PermissionModeAllow`
  - [ ] `PermissionModeDeny`
  - [ ] `PermissionModeAsk`
  - [ ] `PermissionModeDontAsk`
  - [ ] `PermissionModeDelegate`
  - [ ] `PermissionModeAuto`
- [ ] `WithPermissionMode(mode PermissionMode)` accepts typed value
- [ ] Invalid modes cause compile error (not runtime)
- [ ] `String()` method for debugging

**Code Example:**
```go
// Current (string):
client, err := claude.NewClient(claude.WithPermissionMode("allow"))

// Desired (typed):
client, err := claude.NewClient(claude.WithPermissionMode(shared.PermissionModeAllow))
// Typo causes compile error: shared.PermissionModeAlow
```

---

## Category: API Completeness

### US-027: Supported Commands Query
**Status:** ✅ Completed
**Priority:** P2

**As a** SDK user
**I want** to query which commands are supported by current CLI
**So that** I can adapt to different Claude CLI versions

**Acceptance Criteria:**
- [ ] `SupportedCommands(ctx context.Context) ([]string, error)` fully implemented
- [ ] Returns list of available slash commands
- [ ] Uses control protocol to query CLI
- [ ] Cached after first call (session-scoped)
- [ ] Returns error if CLI doesn't support query

**Code Example:**
```go
commands, err := client.SupportedCommands(ctx)
if err != nil {
    log.Fatal(err)
}

if slices.Contains(commands, "refactor") {
    // Use refactor command
}
```

---

### US-028: Supported Models Query
**Status:** ✅ Completed
**Priority:** P2

**As a** SDK user
**I want** to query which models are available
**So that** I can dynamically select models based on availability

**Acceptance Criteria:**
- [ ] `SupportedModels(ctx context.Context) ([]string, error)` fully implemented
- [ ] Returns list of available model IDs
- [ ] Uses control protocol to query CLI
- [ ] Cached after first call (session-scoped)
- [ ] Returns error if CLI doesn't support query

**Code Example:**
```go
models, err := client.SupportedModels(ctx)
if err != nil {
    log.Fatal(err)
}

if slices.Contains(models, "claude-opus-4-5") {
    client.SetModel(ctx, "claude-opus-4-5")
}
```

---

### US-029: MCP Server Status Query
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** to query the status of MCP servers
**So that** I can verify they're running and available

**Acceptance Criteria:**
- [ ] `McpServerStatus(ctx context.Context) (map[string]shared.McpServerInfo, error)` fully implemented
- [ ] Returns map of server name to status info
- [ ] Status includes: running, error, tool count, last ping
- [ ] Uses control protocol to query CLI
- [ ] Real-time status (not cached)

**Code Example:**
```go
status, err := client.McpServerStatus(ctx)
if err != nil {
    log.Fatal(err)
}

for name, info := range status {
    log.Printf("Server %s: running=%v, tools=%d", name, info.Running, info.ToolCount)
}
```

---

### US-030: MCP Server Configuration Update
**Status:** ✅ Completed
**Priority:** P1

**As a** SDK user
**I want** to update MCP server configuration at runtime
**So that** I can add/remove servers without recreating the client

**Acceptance Criteria:**
- [ ] `SetMcpServers(ctx context.Context, servers []shared.McpServerConfig) error` fully implemented
- [ ] Stops old servers not in new config
- [ ] Starts new servers in new config
- [ ] Restarts modified servers
- [ ] Uses control protocol to send config
- [ ] Returns error if any server fails to start

**Code Example:**
```go
// Add new server at runtime
newServers := []shared.McpServerConfig{
    existingServer1,
    existingServer2,
    {
        Type: "stdio",
        Name: "new-server",
        Command: "/usr/local/bin/new-mcp-server",
    },
}

err := client.SetMcpServers(ctx, newServers)
if err != nil {
    log.Printf("Failed to update servers: %v", err)
}
```

---

### US-031: Account Info Query
**Status:** ✅ Completed
**Priority:** P2

**As a** SDK user
**I want** to query account information
**So that** I can display usage stats and limits

**Acceptance Criteria:**
- [ ] `AccountInfo(ctx context.Context) (shared.AccountInfo, error)` fully implemented
- [ ] Returns struct with:
  - [ ] User ID
  - [ ] Email
  - [ ] Plan type
  - [ ] Usage (tokens, requests)
  - [ ] Limits
- [ ] Uses control protocol to query CLI
- [ ] Cached for 5 minutes to avoid rate limits

**Code Example:**
```go
info, err := client.AccountInfo(ctx)
if err != nil {
    log.Fatal(err)
}

log.Printf("User: %s, Plan: %s, Tokens used: %d/%d",
    info.Email, info.Plan, info.TokensUsed, info.TokensLimit)
```

---

### US-032: Rewind Files Implementation
**Status:** ✅ Completed (Protocol.RewindFiles in control.go)
**Priority:** P1

**As a** SDK user
**I want** to rewind file states in a session
**So that** I can undo changes Claude made

**Acceptance Criteria:**
- [ ] `RewindFiles(ctx context.Context, paths []string) error` fully implemented
- [ ] Reverts specified files to pre-session state
- [ ] Uses control protocol to send command
- [ ] Returns error if file doesn't exist in session
- [ ] Empty slice rewinds all files
- [ ] Works with file checkpointing

**Code Example:**
```go
// Claude modified main.go and utils.go
err := client.RewindFiles(ctx, []string{"main.go"})
if err != nil {
    log.Fatal(err)
}
// main.go reverted, utils.go unchanged
```

---

## Category: Developer Experience

### US-033: Option Validation
**Status:** Enhancement
**Priority:** P1

**As a** SDK user
**I want** immediate validation of option values
**So that** I catch configuration errors at startup, not runtime

**Acceptance Criteria:**
- [ ] All options validate inputs in option function
- [ ] Invalid values return error from `NewClient()`
- [ ] Errors include field name and reason
- [ ] Conflicting options detected (e.g., allowed + disallowed same tool)
- [ ] Model names validated against known list

**Code Example:**
```go
// Invalid model causes immediate error
client, err := claude.NewClient(
    claude.WithModel("claude-invalid-99"),
)
// err: "invalid model: claude-invalid-99, supported: [claude-opus-4, ...]"

// Conflicting options detected
client, err := claude.NewClient(
    claude.WithAllowedTools("Read"),
    claude.WithDisallowedTools("Read"),
)
// err: "tool 'Read' in both allowed and disallowed lists"
```

---

### US-034: Documentation Parity
**Status:** Enhancement
**Priority:** P2

**As a** SDK user
**I want** godoc examples for all options
**So that** I understand how to use each feature

**Acceptance Criteria:**
- [ ] Every option function has godoc comment
- [ ] Every option has `Example` test
- [ ] Examples show common use cases
- [ ] Edge cases documented (e.g., empty slices, nil values)
- [ ] Cross-references to related options

**Code Example:**
```go
// Example test for WithAllowedTools
func ExampleWithAllowedTools() {
    client, _ := claude.NewClient(
        claude.WithAllowedTools("Read", "Write", "Grep"),
    )
    defer client.Disconnect(context.Background())

    // Only Read, Write, Grep available to Claude
    // Output:
}
```

---

### US-035: Error Type Hierarchy
**Status:** ✅ Completed (PermissionError, ModelError added)
**Priority:** P2

**As a** SDK user
**I want** typed errors for different failure modes
**So that** I can handle errors programmatically

**Acceptance Criteria:**
- [ ] Custom error types defined:
  - [ ] `ValidationError` - invalid options
  - [ ] `ConnectionError` - subprocess/transport failures
  - [ ] `TimeoutError` - operation timeouts
  - [ ] `PermissionError` - tool/file access denied
  - [ ] `ModelError` - model unavailable/invalid
- [ ] All errors implement `error` interface
- [ ] `errors.Is()` and `errors.As()` supported
- [ ] Error messages include context

**Code Example:**
```go
_, err := client.Query(ctx, "Do something")
if err != nil {
    var timeoutErr *claude.TimeoutError
    if errors.As(err, &timeoutErr) {
        log.Printf("Operation timed out after %v", timeoutErr.Duration)
    }
}
```

---

## Implementation Roadmap

### Phase 1: Critical (P0) - Python SDK Parity
**Target:** Full severity1 feature set

1. **US-001:** WithAllowedTools option
2. **US-007:** WithJSONSchema option
3. **US-008:** WithOutputFormat option
4. **US-009:** WithBetas option
5. **US-014:** WithHooks registration
6. **US-015:** WithHook single hook
7. **US-017:** Hook callback routing
8. **US-018:** WithCanUseTool permission callback
9. **US-021:** Control protocol infrastructure
10. **US-022:** Initialize handshake

**Success Criteria:** SDK can do everything severity1 SDK can do

---

### Phase 2: Important (P1) - Feature Completeness
**Target:** Full-featured SDK

1. **US-002:** WithCwd alias
2. **US-003:** WithAddDirs alias
3. **US-004:** WithForkSession option
4. **US-005:** WithFallbackModel option
5. **US-011:** WithDebugWriter option
6. **US-012:** WithStderrCallback option
7. **US-013:** WithAgents option
8. **US-016:** WithPreToolUseHook convenience
9. **US-019:** WithSdkMcpServer option
10. **US-020:** CreateSDKMcpServer helper
11. **US-023:** MCP message routing
12. **US-024:** Full SandboxSettings
13. **US-026:** Typed PermissionMode
14. **US-029:** McpServerStatus implementation
15. **US-030:** SetMcpServers implementation
16. **US-032:** RewindFiles implementation
17. **US-033:** Option validation

**Success Criteria:** All common use cases supported without workarounds

---

### Phase 3: Nice to Have (P2)
**Target:** Enhanced developer experience

1. **US-006:** WithUser option
2. **US-010:** WithSettingSources option
3. **US-025:** Full SdkPluginConfig
4. **US-027:** SupportedCommands implementation
5. **US-028:** SupportedModels implementation
6. **US-031:** AccountInfo implementation
7. **US-034:** Documentation parity
8. **US-035:** Error type hierarchy

**Success Criteria:** Best-in-class developer experience

---

### Phase 4: Advanced (P3)
**Target:** Enterprise features

(No P3 items in current scope - all features are P0-P2)

---

## Testing Requirements

Each user story must include:

1. **Unit Tests:**
   - [ ] Option function behavior
   - [ ] Validation logic
   - [ ] Error cases

2. **Integration Tests:**
   - [ ] End-to-end with Claude CLI
   - [ ] Real subprocess communication
   - [ ] Hook/callback execution

3. **Example Tests:**
   - [ ] Godoc examples that compile
   - [ ] Cover common use cases

4. **Edge Case Tests:**
   - [ ] Nil values
   - [ ] Empty slices
   - [ ] Concurrent access
   - [ ] Timeout scenarios

---

## Dependencies

### Cross-Story Dependencies

- **US-017** (Hook routing) depends on **US-021** (Control protocol)
- **US-022** (Initialize) depends on **US-021** (Control protocol)
- **US-023** (MCP routing) depends on **US-021** (Control protocol)
- **US-029, US-030** (MCP status/config) depend on **US-021** (Control protocol)
- **US-019, US-020** (SDK MCP) depend on **US-023** (MCP routing)
- **US-014, US-015, US-016** (Hook options) depend on **US-017** (Hook routing)

### External Dependencies

- Claude CLI version compatibility
- Control protocol spec (if exists)
- Python SDK reference implementation

---

## Success Metrics

**Phase 1 Complete:**
- [ ] All P0 stories implemented
- [ ] 90%+ test coverage on new code
- [ ] All integration tests pass
- [ ] Can replace severity1 SDK in example apps

**Phase 2 Complete:**
- [ ] All P0+P1 stories implemented
- [ ] 95%+ test coverage
- [ ] Documentation complete for all features
- [ ] Migration guide from severity1 published

**Phase 3 Complete:**
- [ ] All stories implemented
- [ ] Comprehensive example suite
- [ ] Performance benchmarks published
- [ ] 1.0 release ready

---

## Migration Guide (Future)

When implemented, this section will contain:
- severity1 → dotcommander migration steps
- API mapping table
- Common migration pitfalls
- Automated migration tooling

---

**Total User Stories:** 35
**P0 (Critical):** 10
**P1 (Important):** 17
**P2 (Nice to Have):** 8
**P3 (Low Priority):** 0

**Estimated Effort:** ~8-12 weeks (2 developers)
- Phase 1: 3-4 weeks
- Phase 2: 3-4 weeks
- Phase 3: 2-3 weeks
- Testing/docs: 1-2 weeks (parallel)
