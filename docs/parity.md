# Feature Parity Comparison: dotcommander vs severity1

Comparison of this SDK (`github.com/dotcommander/agent-sdk-go`) against the severity1 SDK (`github.com/severity1/claude-agent-sdk-go`).

## Package Structure

| Aspect | severity1 | dotcommander (this project) |
|--------|-----------|----------------------------|
| Root Package | `claudecode` | `claude` |
| Architecture | Internal packages (`internal/*`) | Public packages (`claude/*`) |
| Import Path | `github.com/severity1/claude-agent-sdk-go` | `github.com/dotcommander/agent-sdk-go/claude` |

## Core Client API

| Feature | severity1 | dotcommander |
|---------|-----------|--------------|
| `NewClient()` | ✅ Returns `Client` | ✅ Returns `(Client, error)` |
| `WithClient()` | ✅ Resource management | ✅ Same pattern |
| `Query()` one-shot | ✅ Returns `MessageIterator` | ✅ Returns `MessageIterator` |
| `QueryText()` | ❌ | ✅ Convenience method |
| `QueryWithSession()` | ✅ | ✅ |
| `Connect/Disconnect` | ✅ | ✅ |
| `ReceiveMessages` | ✅ Returns `<-chan Message` | ✅ Returns `(<-chan Message, <-chan error)` |
| `ReceiveResponse` | ✅ Returns `MessageIterator` | ✅ Returns `(Message, error)` |
| `ReceiveResponseIterator` | ❌ | ✅ Returns `MessageIterator` |
| Interface Composition | ❌ Single interface | ✅ `Connector`, `Querier`, `Receiver`, `Controller`, `ContextManager` |

## Runtime Control Methods

| Feature | severity1 | dotcommander |
|---------|-----------|--------------|
| `SetModel()` | ✅ | ✅ |
| `SetPermissionMode()` | ✅ | ✅ |
| `RewindFiles()` | ✅ | ✅ |
| `Interrupt()` | ✅ | ✅ |
| `InterruptGraceful()` | ❌ | ✅ |
| `GetStreamIssues()` | ✅ | ✅ |
| `GetStreamStats()` | ✅ | ✅ |
| `GetServerInfo()` | ✅ | ✅ |
| `IsProtocolActive()` | ❌ | ✅ |
| `McpServerStatus()` | ❌ | ✅ (stub) |
| `SetMcpServers()` | ❌ | ✅ (stub) |

## Options Configuration

| Feature | severity1 | dotcommander |
|---------|-----------|--------------|
| `WithModel()` | ✅ | ✅ |
| `WithSystemPrompt()` | ✅ | ✅ |
| `WithAppendSystemPrompt()` | ✅ | ✅ |
| `WithAllowedTools()` | ✅ | ✅ |
| `WithDisallowedTools()` | ✅ | ✅ |
| `WithToolsPreset()` | ✅ | ✅ |
| `WithClaudeCodeTools()` | ✅ | ✅ |
| `WithPermissionMode()` | ✅ (typed `PermissionMode`) | ✅ (string + `WithTypedPermissionMode`) |
| `WithMaxTurns()` | ✅ | ✅ |
| `WithMaxThinkingTokens()` | ✅ | ✅ |
| `WithMaxBudgetUSD()` | ✅ | ✅ |
| `WithCwd()` | ✅ | ✅ (alias) |
| `WithWorkingDirectory()` | ❌ | ✅ |
| `WithAddDirs()` | ✅ | ✅ (alias) |
| `WithResume()` | ✅ | ✅ |
| `WithContinue()` | ❌ | ✅ |
| `WithForkSession()` | ✅ | ✅ |
| `WithFallbackModel()` | ✅ | ✅ |
| `WithUser()` | ✅ | ✅ |
| `WithEnv()` | ✅ | ✅ |
| `WithTimeout()` | ❌ | ✅ |
| `WithBufferSize()` | ❌ | ✅ |
| `WithFileCheckpointing()` | ✅ | ✅ |
| `WithAgent()` | ❌ | ✅ |

## Hooks System

| Feature | severity1 | dotcommander |
|---------|-----------|--------------|
| Hook Events | 6 events | 12 events |
| `PreToolUse` | ✅ | ✅ |
| `PostToolUse` | ✅ | ✅ |
| `PostToolUseFailure` | ❌ | ✅ |
| `Notification` | ❌ | ✅ |
| `UserPromptSubmit` | ✅ | ✅ |
| `SessionStart` | ❌ | ✅ |
| `SessionEnd` | ❌ | ✅ |
| `Stop` | ✅ | ✅ |
| `SubagentStart` | ❌ | ✅ |
| `SubagentStop` | ✅ | ✅ |
| `PreCompact` | ✅ | ✅ |
| `PermissionRequest` | ❌ | ✅ |
| `WithHooks()` option | ✅ | ✅ |
| `WithHook()` single hook | ✅ | ✅ |
| `WithPreToolUseHook()` | ✅ | ✅ |
| `WithPostToolUseHook()` | ❌ | ✅ |
| `WithSessionStartHook()` | ❌ | ✅ |
| `WithSessionEndHook()` | ❌ | ✅ |
| Hook callback routing | ✅ (control protocol) | ⚠️ Types defined, wiring pending |

## Permissions System

| Feature | severity1 | dotcommander |
|---------|-----------|--------------|
| `CanUseToolCallback` | ✅ | ✅ |
| `WithCanUseTool()` | ✅ | ✅ |
| `PermissionResultAllow` | ✅ Separate type | ✅ Struct with behavior |
| `PermissionResultDeny` | ✅ Separate type | ✅ Struct with behavior |
| `NewPermissionResultAllow()` | ✅ | ✅ |
| `NewPermissionResultDeny()` | ✅ | ✅ |
| `ToolPermissionContext` | ✅ | ✅ (as `CanUseToolOptions`) |
| `PermissionUpdate` | ✅ | ✅ |
| Permission modes | 4 modes | 6 modes (has delegate, dontAsk) |

## MCP Integration

| Feature | severity1 | dotcommander |
|---------|-----------|--------------|
| `McpServerConfig` interface | ✅ | ✅ |
| `McpStdioServerConfig` | ✅ | ✅ |
| `McpSSEServerConfig` | ✅ | ✅ |
| `McpHTTPServerConfig` | ✅ | ✅ |
| `McpSdkServerConfig` (in-process) | ✅ | ✅ |
| `McpServer` interface | ✅ | ✅ |
| `McpToolDefinition` | ✅ | ✅ |
| `McpToolResult` | ✅ | ✅ |
| `WithMcpServers()` | ✅ | ✅ |
| `WithSdkMcpServer()` | ✅ | ✅ |
| `CreateSDKMcpServer()` helper | ✅ | ✅ |

## Control Protocol

| Feature | severity1 | dotcommander |
|---------|-----------|--------------|
| Protocol types | ✅ Full `internal/control` | ❌ In subprocess |
| Request/response correlation | ✅ | ❌ |
| Initialize handshake | ✅ | ❌ |
| Hook callback routing | ✅ | ❌ |
| MCP message routing | ✅ | ❌ |

## Additional Features

| Feature | severity1 | dotcommander |
|---------|-----------|--------------|
| `SandboxSettings` (full) | ✅ Network config, violations | ✅ Basic + `WithSandboxSettings()` |
| `SdkPluginConfig` | ✅ | ⚠️ Basic `PluginConfig` |
| `AgentDefinition` | ✅ In options | ✅ Separate |
| `WithAgents()` option | ✅ | ✅ |
| `OutputFormat` | ✅ | ✅ |
| `WithJSONSchema()` | ✅ | ✅ |
| `WithOutputFormat()` | ✅ | ✅ |
| `WithBetas()` | ✅ | ✅ |
| `SdkBeta` constants | ✅ | ✅ |
| `WithSettingSources()` | ✅ | ✅ |
| `WithDebugWriter()` | ✅ | ✅ |
| `WithStderrCallback()` | ✅ | ✅ |
| `WithPartialStreaming()` | ✅ | ✅ (`WithIncludePartialMessages`) |
| Parser registry | ❌ | ✅ |
| V2 Session API | ❌ | ✅ (`claude/v2` package) |

## Summary

### Implementation Status (as of 2026-01-18)

**P0 Critical - COMPLETED:**
1. ✅ `WithCanUseTool()` - permission callback option
2. ✅ `WithHooks()` / `WithHook()` - hook registration
3. ✅ `WithAllowedTools()` - tool restriction
4. ✅ `WithBetas()` - beta feature enablement
5. ✅ `WithPreToolUseHook()`, `WithPostToolUseHook()`, etc. - typed hook helpers
6. ✅ `WithJSONSchema()` / `WithOutputFormat()` - structured output

**P1 Important - COMPLETED:**
7. ✅ `WithSdkMcpServer()` - in-process MCP servers
8. ✅ `WithForkSession()` - session forking
9. ✅ `WithFallbackModel()` - fallback model
10. ✅ `WithDebugWriter()` / `WithStderrCallback()` - debug output
11. ✅ `WithCwd()` / `WithAddDirs()` - severity1 compatibility aliases
12. ✅ `WithUser()` - user identifier
13. ✅ `WithAgents()` - multiple agent definitions
14. ✅ `WithSettingSources()` - setting source control
15. ✅ `WithSandboxSettings()` - sandbox configuration

**Remaining:**
- Control protocol initialization and callback routing (types exist, wiring pending)
- Full bidirectional communication implementation

### severity1 Strengths

- Single cohesive `Options` struct matching Python SDK
- Full control protocol with bidirectional communication
- Hook callback routing through control protocol

### dotcommander Strengths

- Interface composition (more Go-idiomatic)
- More hook event types defined (12 vs 6)
- V2 Session API package
- Parser registry for extensibility
- More permission modes (6 vs 4)
- Full P0+P1 option parity now achieved
