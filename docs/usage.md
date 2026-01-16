# Usage Guide

This guide covers how to use `agent-sdk-go` to interact with Claude CLI programmatically.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Sessions](#sessions)
- [One-Shot Prompts](#one-shot-prompts)
- [Options](#options)
- [Message Types](#message-types)
- [Permissions](#permissions)
- [Hooks](#hooks)
- [MCP Servers](#mcp-servers)
- [Agents](#agents)
- [Sandbox](#sandbox)
- [Tools](#tools)

---

## Installation

```bash
go get agent-sdk-go
```

**Prerequisites:** Claude CLI must be installed and authenticated.

```bash
# Install Claude CLI
npm install -g @anthropic-ai/claude-cli

# Authenticate
claude auth login
```

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "agent-sdk-go/internal/claude/v2"
)

func main() {
    ctx := context.Background()

    // One-shot prompt
    result, err := v2.Prompt(ctx, "What is 2+2?",
        v2.WithPromptModel("claude-sonnet-4-20250514"),
    )
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(result.Result)
}
```

---

## Sessions

Sessions maintain conversation state across multiple turns.

### Creating a Session

```go
session, err := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-20250514"),
    v2.WithTimeout(60*time.Second),
    v2.WithSystemPrompt("You are a helpful assistant."),
)
if err != nil {
    log.Fatal(err)
}
defer session.Close()
```

### Sending Messages

```go
// Send a message
err = session.Send(ctx, "Hello, Claude!")
if err != nil {
    log.Fatal(err)
}

// Receive response via channel
for msg := range session.Receive(ctx) {
    switch msg.Type() {
    case v2.V2EventTypeAssistant:
        fmt.Println(v2.ExtractText(msg))
    case v2.V2EventTypeResult:
        fmt.Println("Done:", v2.ExtractResultText(msg))
    case v2.V2EventTypeError:
        fmt.Println("Error:", v2.ExtractErrorMessage(msg))
    }
}
```

### Using Iterator Pattern

```go
iter := session.ReceiveIterator(ctx)
defer iter.Close()

for {
    msg, err := iter.Next(ctx)
    if err == v2.ErrNoMoreMessages {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(v2.ExtractText(msg))
}
```

### Resuming Sessions

```go
// Save the session ID
sessionID := session.SessionID()

// Later, resume the session
resumedSession, err := v2.ResumeSession(ctx, sessionID,
    v2.WithModel("claude-sonnet-4-20250514"),
)
```

---

## One-Shot Prompts

For single question/answer interactions:

```go
result, err := v2.Prompt(ctx, "Explain quantum computing in one sentence.",
    v2.WithPromptModel("claude-sonnet-4-20250514"),
    v2.WithPromptTimeout(30*time.Second),
    v2.WithPromptSystemPrompt("Be concise."),
)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Response: %s\n", result.Result)
fmt.Printf("Duration: %v\n", result.Duration)
```

---

## Options

### Model Selection

```go
// Full model name
v2.WithModel("claude-sonnet-4-20250514")

// Short names (resolved automatically)
v2.WithModel("sonnet")  // -> claude-sonnet-4-5-20250929
v2.WithModel("opus")    // -> claude-opus-4-5-20251101
v2.WithModel("haiku")   // -> claude-3-5-haiku-20241022
```

### Session Control

```go
v2.WithContinue(true)                    // Continue recent conversation
v2.WithResume("session-id")              // Resume specific session
v2.WithResumeSessionAt("message-uuid")   // Resume at specific message
v2.WithForkSession(true)                 // Fork instead of continue
v2.WithPersistSession(true)              // Save to disk (default)
```

### Tool Configuration

```go
v2.WithAllowedTools("Read", "Write", "Bash")
v2.WithDisallowedTools("WebFetch")
```

### Limits

```go
v2.WithMaxTurns(10)
v2.WithMaxThinkingTokens(4096)
v2.WithMaxBudgetUSD(1.0)
v2.WithFallbackModel("claude-3-5-haiku-20241022")
```

### Advanced Options

```go
v2.WithAdditionalDirectories("/extra/path")
v2.WithBetas("context-1m-2025-08-07")
v2.WithEnableFileCheckpointing(true)
v2.WithStrictMcpConfig(true)
```

---

## Message Types

### Core Message Types

| Type | Description |
|------|-------------|
| `assistant` | Response from Claude |
| `result` | Final result with usage stats |
| `stream_event` | Streaming delta updates |
| `system` | System messages (init, status, hooks) |
| `tool_progress` | Tool execution progress |
| `auth_status` | Authentication status |

### Extracting Content

```go
// From assistant messages
text := v2.ExtractAssistantText(msg)

// From result messages
result := v2.ExtractResultText(msg)

// From stream deltas
delta := v2.ExtractDeltaText(msg)

// From any message with text
text := v2.ExtractText(msg)
```

### Type Checking

```go
if v2.IsAssistantMessage(msg) { ... }
if v2.IsResultMessage(msg) { ... }
if v2.IsStreamDelta(msg) { ... }
if v2.IsErrorMessage(msg) { ... }
```

### Result Subtypes

```go
import "agent-sdk-go/internal/claude/shared"

switch resultMsg.Subtype {
case shared.ResultSubtypeSuccess:
    // Normal completion
case shared.ResultSubtypeErrorMaxTurns:
    // Hit turn limit
case shared.ResultSubtypeErrorMaxBudgetUSD:
    // Hit budget limit
case shared.ResultSubtypeErrorDuringExecution:
    // Error occurred
}
```

---

## Permissions

### Permission Modes

```go
import "agent-sdk-go/internal/claude/shared"

v2.WithPermissionMode(string(shared.PermissionModeDefault))
v2.WithPermissionMode(string(shared.PermissionModeAcceptEdits))
v2.WithPermissionMode(string(shared.PermissionModeBypassPermissions))
v2.WithPermissionMode(string(shared.PermissionModePlan))
```

| Mode | Description |
|------|-------------|
| `default` | Ask for each tool use |
| `acceptEdits` | Auto-accept file edits |
| `bypassPermissions` | Skip all prompts (requires `AllowDangerouslySkipPermissions`) |
| `plan` | Plan mode, no execution |
| `delegate` | Delegate to parent |
| `dontAsk` | Deny without asking |

### Permission Results

```go
// Allow with modifications
result := shared.PermissionResult{
    Behavior:     shared.PermissionBehaviorAllow,
    UpdatedInput: map[string]any{"path": "/safe/path"},
}

// Deny with message
result := shared.PermissionResult{
    Behavior:  shared.PermissionBehaviorDeny,
    Message:   "Operation not allowed",
    Interrupt: true,
}
```

---

## Hooks

Hooks allow intercepting and modifying behavior at various points.

### Hook Events

| Event | When |
|-------|------|
| `PreToolUse` | Before tool execution |
| `PostToolUse` | After successful tool execution |
| `PostToolUseFailure` | After tool failure |
| `SessionStart` | Session begins |
| `SessionEnd` | Session ends |
| `UserPromptSubmit` | User submits prompt |
| `Notification` | System notification |
| `Stop` | Stop requested |
| `SubagentStart` | Subagent spawned |
| `SubagentStop` | Subagent finished |
| `PreCompact` | Before conversation compaction |
| `PermissionRequest` | Permission prompt shown |

### Hook Input Types

```go
import "agent-sdk-go/internal/claude/shared"

// PreToolUse hook input
input := shared.PreToolUseHookInput{
    BaseHookInput: shared.BaseHookInput{
        SessionID:      "sess-123",
        TranscriptPath: "/path/to/transcript.jsonl",
        Cwd:            "/working/dir",
    },
    ToolName:  "Bash",
    ToolInput: map[string]any{"command": "ls -la"},
    ToolUseID: "tool-456",
}
```

### Hook Outputs

```go
// Synchronous response
output := shared.SyncHookOutput{
    Continue:   true,
    Decision:   "approve",
    Reason:     "Safe operation",
}

// Async response
output := shared.AsyncHookOutput{
    Async:        true,
    AsyncTimeout: 30,
}
```

---

## MCP Servers

Configure Model Context Protocol servers for extended capabilities.

### Server Types

```go
import "agent-sdk-go/internal/claude/shared"

// Stdio server
stdioServer := shared.McpStdioServerConfig{
    Type:    "stdio",
    Command: "node",
    Args:    []string{"server.js"},
    Env:     map[string]string{"DEBUG": "true"},
}

// SSE server
sseServer := shared.McpSSEServerConfig{
    Type:    "sse",
    URL:     "http://localhost:3000/events",
    Headers: map[string]string{"Authorization": "Bearer token"},
}

// HTTP server
httpServer := shared.McpHttpServerConfig{
    Type: "http",
    URL:  "http://localhost:3000/api",
}
```

### Configuring MCP Servers

```go
servers := map[string]shared.McpServerConfig{
    "my-server": shared.McpStdioServerConfig{
        Command: "npx",
        Args:    []string{"-y", "@my/mcp-server"},
    },
}

v2.WithMcpServers(servers)
```

### Server Status

```go
status := shared.McpServerStatus{
    Name:   "my-server",
    Status: "connected", // "connected" | "failed" | "needs-auth" | "pending"
    ServerInfo: &shared.McpServerInfo{
        Name:    "My Server",
        Version: "1.0.0",
    },
}
```

---

## Agents

Define custom subagents for specialized tasks.

### Agent Definition

```go
import "agent-sdk-go/internal/claude/shared"

agent := shared.AgentDefinition{
    Description:     "Code review specialist",
    Prompt:          "You are a code reviewer. Focus on security and performance.",
    Model:           shared.AgentModelSonnet,
    Tools:           []string{"Read", "Grep", "Glob"},
    DisallowedTools: []string{"Write", "Bash"},
}
```

### Configuring Agents

```go
agents := map[string]shared.AgentDefinition{
    "reviewer": {
        Description: "Code reviewer",
        Prompt:      "Review code for issues.",
        Model:       shared.AgentModelSonnet,
        Tools:       []string{"Read", "Grep"},
    },
    "writer": {
        Description: "Code writer",
        Prompt:      "Write clean, tested code.",
        Model:       shared.AgentModelOpus,
    },
}

v2.WithAgents(agents)
v2.WithAgent("reviewer") // Set main thread agent
```

### Agent Models

| Model | Description |
|-------|-------------|
| `sonnet` | Balanced performance |
| `opus` | Highest capability |
| `haiku` | Fast, lightweight |
| `inherit` | Use parent's model |

---

## Sandbox

Configure command execution isolation.

### Sandbox Settings

```go
import "agent-sdk-go/internal/claude/shared"

sandbox := &shared.SandboxSettings{
    Enabled:                  true,
    AutoAllowBashIfSandboxed: true,
    AllowUnsandboxedCommands: false,
    ExcludedCommands:         []string{"rm", "dd"},
}

v2.WithSandbox(sandbox)
```

### Network Configuration

```go
sandbox := &shared.SandboxSettings{
    Enabled: true,
    Network: &shared.SandboxNetworkConfig{
        AllowedDomains:    []string{"api.example.com"},
        AllowLocalBinding: true,
        HttpProxyPort:     8080,
    },
}
```

---

## Tools

Tool input types for building custom integrations.

### Common Tool Inputs

```go
import "agent-sdk-go/internal/claude/shared/tools"

// Bash
bash := tools.BashInput{
    Command:     "ls -la",
    Description: "List files",
    Timeout:     5000,
}

// File operations
read := tools.FileReadInput{
    FilePath: "/path/to/file.go",
    Offset:   0,
    Limit:    1000,
}

write := tools.FileWriteInput{
    FilePath: "/path/to/output.txt",
    Content:  "Hello, World!",
}

edit := tools.FileEditInput{
    FilePath:   "/path/to/file.go",
    OldString:  "oldFunc",
    NewString:  "newFunc",
    ReplaceAll: true,
}

// Search
grep := tools.GrepInput{
    Pattern:    "TODO",
    Path:       "/src",
    OutputMode: "content",
    HeadLimit:  100,
}

glob := tools.GlobInput{
    Pattern: "**/*.go",
    Path:    "/src",
}
```

### Agent Tool Input

```go
agent := tools.AgentInput{
    Description:     "Explore codebase",
    Prompt:          "Find all API endpoints",
    SubagentType:    "Explore",
    Model:           "sonnet",
    MaxTurns:        10,
    RunInBackground: false,
}
```

### Interactive Tools

```go
// Ask user question
question := tools.AskUserQuestionInput{
    Questions: []tools.Question{
        {
            Question:    "Which framework?",
            Header:      "Framework",
            MultiSelect: false,
            Options: []tools.QuestionOption{
                {Label: "React", Description: "UI library"},
                {Label: "Vue", Description: "Progressive framework"},
            },
        },
    },
}

// Todo list
todos := tools.TodoWriteInput{
    Todos: []tools.TodoItem{
        {Content: "Implement feature", Status: "in_progress", ActiveForm: "Implementing feature"},
        {Content: "Write tests", Status: "pending", ActiveForm: "Writing tests"},
    },
}
```

---

## Error Handling

```go
result, err := v2.Prompt(ctx, "Hello")
if err != nil {
    // Check for specific errors
    var protoErr *shared.ProtocolError
    if errors.As(err, &protoErr) {
        fmt.Printf("Protocol error: %s - %s\n", protoErr.Code, protoErr.Message)
    }

    var parseErr *shared.ParserError
    if errors.As(err, &parseErr) {
        fmt.Printf("Parse error at line %d: %s\n", parseErr.LineNumber, parseErr.Message)
    }
}
```

---

## Best Practices

1. **Always close sessions** - Use `defer session.Close()`
2. **Set appropriate timeouts** - Prevent hung operations
3. **Handle all message types** - Don't ignore errors or unexpected types
4. **Use context cancellation** - Allow graceful shutdown
5. **Validate inputs** - Check options before creating sessions

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

session, err := v2.CreateSession(ctx,
    v2.WithModel("sonnet"),
    v2.WithTimeout(60*time.Second),
)
if err != nil {
    log.Fatal(err)
}
defer session.Close()
```
