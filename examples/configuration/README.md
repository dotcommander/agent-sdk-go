# Example: Configuration

## What This Demonstrates

This example shows all available configuration options for the Claude Agent SDK. It demonstrates:

- Model selection with short aliases and fallbacks
- Timeout and limit settings (turns, tokens, budget)
- Tool restrictions (allowed/disallowed tools)
- Context configuration (system prompt, files, directories)
- MCP server configuration (stdio, SSE, HTTP)
- Environment and subprocess settings
- Advanced options (persistence, checkpointing, output format)

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

```bash
cd examples/configuration
go run main.go

# Or run with interactive demo (requires Claude CLI)
go run main.go demo
```

## Expected Output

```
=== Configuration Example ===
This demonstrates all available configuration options.

--- Basic Options ---
Model Options:
  claude-sonnet-4-5-20250929
  claude-3-5-haiku-20241022
  claude-opus-4-20250514
  sonnet -> claude-sonnet-4-5-20250929
  haiku -> claude-3-5-haiku-20241022
  opus -> claude-opus-4-20250514

Timeout Options:
  30s - Quick queries, simple questions
  1m0s - Standard operations
  2m0s - Complex tasks, code generation
  5m0s - Long-running jobs, analysis

--- Limit Options ---
MaxTurns - Limits conversation turns:
  1   - Single response (no follow-up)
  5   - Short conversations
  10  - Extended discussions
  nil - Unlimited (default)
...
```

## Key Patterns

### Pattern 1: Model Selection

Choose models with full names or short aliases:

```go
// Full model names
session, _ := v2.CreateSession(ctx, v2.WithModel("claude-sonnet-4-5-20250929"))

// Short aliases (auto-resolved)
session, _ := v2.CreateSession(ctx, v2.WithModel("sonnet"))
session, _ := v2.CreateSession(ctx, v2.WithModel("haiku"))
session, _ := v2.CreateSession(ctx, v2.WithModel("opus"))

// With fallback
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithFallbackModel("claude-3-5-haiku-20241022"),
)
```

### Pattern 2: Limits Configuration

Control costs and conversation scope:

```go
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithMaxTurns(5),              // Max 5 conversation turns
    v2.WithMaxThinkingTokens(4096),  // Limit internal reasoning
    v2.WithMaxBudgetUSD(1.00),       // $1 spending limit
    v2.WithTimeout(60*time.Second),  // 60s operation timeout
)
```

### Pattern 3: Tool Restrictions

Control which tools Claude can use:

```go
// Allow only specific tools
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithAllowedTools("Read", "Grep", "Glob"),  // Read-only
)

// Block specific tools
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithDisallowedTools("Bash", "Write"),  // No shell or writes
)
```

### Pattern 4: Context Configuration

Provide context and instructions:

```go
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithSystemPrompt("You are a Go expert. Follow idiomatic patterns."),
    v2.WithContextFiles("main.go", "go.mod", "README.md"),
    v2.WithAdditionalDirectories("/tmp", "/var/data"),
)
```

### Pattern 5: MCP Server Configuration

Connect to MCP servers for extended capabilities:

```go
servers := map[string]shared.McpServerConfig{
    "filesystem": {
        Type:    "stdio",
        Command: "npx",
        Args:    []string{"-y", "@modelcontextprotocol/server-filesystem"},
    },
    "database": {
        Type: "http",
        URL:  "http://localhost:8080/mcp",
    },
    "realtime": {
        Type: "sse",
        URL:  "http://localhost:3000/sse",
    },
}

session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithMcpServers(servers),
)
```

### Pattern 6: Structured Output

Get responses in a specific format:

```go
format := &shared.OutputFormat{
    Type: "json_schema",
    Schema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "answer": map[string]any{"type": "string"},
            "confidence": map[string]any{"type": "number"},
        },
        "required": []string{"answer", "confidence"},
    },
}

session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithOutputFormat(format),
)
```

### Pattern 7: Environment Configuration

Set subprocess environment:

```go
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithEnv(map[string]string{
        "DEBUG":    "true",
        "LOG_PATH": "/tmp/claude.log",
    }),
    v2.WithCustomArgs("--verbose", "--no-color"),
)
```

## Configuration Reference

| Option | Type | Description |
|--------|------|-------------|
| `WithModel` | `string` | Claude model to use |
| `WithFallbackModel` | `string` | Fallback if primary fails |
| `WithTimeout` | `time.Duration` | Operation timeout |
| `WithMaxTurns` | `int` | Max conversation turns |
| `WithMaxThinkingTokens` | `int` | Max reasoning tokens |
| `WithMaxBudgetUSD` | `float64` | USD spending limit |
| `WithSystemPrompt` | `string` | System instructions |
| `WithContextFiles` | `...string` | Files to include |
| `WithAdditionalDirectories` | `...string` | Accessible paths |
| `WithAllowedTools` | `...string` | Tools whitelist |
| `WithDisallowedTools` | `...string` | Tools blacklist |
| `WithPermissionMode` | `string` | Permission handling |
| `WithMcpServers` | `map[string]McpServerConfig` | MCP servers |
| `WithOutputFormat` | `*OutputFormat` | Response format |
| `WithPersistSession` | `bool` | Save session to disk |
| `WithEnableFileCheckpointing` | `bool` | Track file changes |
| `WithEnv` | `map[string]string` | Environment vars |
| `WithCustomArgs` | `...string` | CLI arguments |
| `WithBetas` | `...string` | Beta features |

## TypeScript Equivalent

This ports configuration patterns from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/configuration

The TypeScript version uses:
```typescript
const session = await createSession({
    model: "claude-sonnet-4-5-20250929",
    systemPrompt: "You are helpful.",
    maxTurns: 5,
    maxBudget: 1.00,
    mcpServers: { ... },
});
```

## Related Documentation

- [Options Reference](../../docs/usage.md#options)
- [MCP Configuration](../../docs/usage.md#mcp-servers)
- [Model Selection](../../docs/usage.md#models)
