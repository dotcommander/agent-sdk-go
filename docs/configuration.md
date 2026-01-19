# Configuration

Complete reference for all SDK configuration options.

## Client Options

Configure the client using functional options:

```go
client, err := claude.NewClient(
    claude.WithModel("claude-sonnet-4-20250514"),
    claude.WithTimeout("60s"),
    // ... more options
)
```

### Available Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `WithModel` | `string` | `""` | Claude model to use |
| `WithTimeout` | `string` | `"30s"` | Request timeout |
| `WithSystemPrompt` | `string` | `""` | System prompt |
| `WithCLIPath` | `string` | auto-detect | Path to Claude CLI |
| `WithEnv` | `map[string]string` | `nil` | Environment variables |
| `WithCustomArgs` | `[]string` | `nil` | Additional CLI arguments |

## Models

Available model identifiers:

```go
// Claude 3.5 models
claude.WithModel("claude-3-5-sonnet-20241022")
claude.WithModel("claude-3-5-haiku-20241022")

// Claude 3 models
claude.WithModel("claude-3-opus-20240229")
claude.WithModel("claude-3-sonnet-20240229")

// Claude 4 models
claude.WithModel("claude-sonnet-4-20250514")
claude.WithModel("claude-opus-4-5-20251101")
```

### Model Selection Guide

| Model | Best For |
|-------|----------|
| `claude-3-5-haiku` | Fast, simple tasks |
| `claude-3-5-sonnet` | Balanced performance |
| `claude-sonnet-4` | Latest balanced model |
| `claude-3-opus` | Complex reasoning |
| `claude-opus-4-5` | Most capable |

## Timeout Configuration

```go
// String format (parsed as duration)
claude.WithTimeout("30s")
claude.WithTimeout("5m")
claude.WithTimeout("1h")

// Via transport config
config := &subprocess.TransportConfig{
    Timeout: 60 * time.Second,
}
```

### Timeout Guidelines

| Task Type | Recommended |
|-----------|-------------|
| Simple questions | `30s` |
| Code generation | `60s` - `120s` |
| Complex analysis | `300s` |
| Long-form writing | `600s` |

## System Prompts

Set persistent instructions:

```go
claude.WithSystemPrompt("You are a Go programming expert. Always provide idiomatic code examples.")
```

### Effective System Prompts

```go
// Role definition
claude.WithSystemPrompt("You are a senior software engineer reviewing code.")

// Output format
claude.WithSystemPrompt("Always respond in JSON format with 'answer' and 'confidence' fields.")

// Constraints
claude.WithSystemPrompt("Keep responses under 100 words. Be direct and concise.")

// Combined
claude.WithSystemPrompt(`You are a Go expert.
- Provide idiomatic Go code
- Include error handling
- Add brief comments
- Use standard library when possible`)
```

## CLI Path

Specify custom CLI location:

```go
// Explicit path
claude.WithCLIPath("/usr/local/bin/claude")

// From environment
claude.WithCLIPath(os.Getenv("CLAUDE_CLI_PATH"))
```

### Auto-Detection

By default, the SDK searches for the CLI in:

1. `claude` in `PATH`
2. `~/.local/bin/claude`
3. `/usr/local/bin/claude`

## Environment Variables

Pass environment variables to the CLI:

```go
claude.WithEnv(map[string]string{
    "ANTHROPIC_API_KEY": apiKey,
    "HTTP_PROXY":        proxyURL,
})
```

## Custom Arguments

Add extra CLI arguments:

```go
claude.WithCustomArgs([]string{
    "--verbose",
    "--no-color",
})
```

## Transport Configuration

For low-level control, configure the transport directly:

```go
import "github.com/dotcommander/agent-sdk-go/claude/subprocess"

config := &subprocess.TransportConfig{
    CLIPath:      "/usr/local/bin/claude",
    CLICommand:   "claude",
    Model:        "claude-sonnet-4-20250514",
    Timeout:      60 * time.Second,
    SystemPrompt: "You are helpful",
    CustomArgs:   []string{"--verbose"},
    Env: map[string]string{
        "DEBUG": "1",
    },
}

transport, err := subprocess.NewTransport(config)
```

### TransportConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `CLIPath` | `string` | Full path to CLI binary |
| `CLICommand` | `string` | CLI command name |
| `Model` | `string` | Model identifier |
| `Timeout` | `time.Duration` | Operation timeout |
| `SystemPrompt` | `string` | System instructions |
| `CustomArgs` | `[]string` | Additional arguments |
| `Env` | `map[string]string` | Environment variables |
| `McpServers` | `map[string]McpServerConfig` | MCP server configs |

## MCP Server Configuration

Configure Model Context Protocol servers:

```go
import "github.com/dotcommander/agent-sdk-go/claude"

config := &subprocess.TransportConfig{
    McpServers: map[string]claude.McpServerConfig{
        "filesystem": {
            Command: "mcp-server-filesystem",
            Args:    []string{"/home/user/projects"},
        },
        "github": {
            Command: "mcp-server-github",
            Env: map[string]string{
                "GITHUB_TOKEN": token,
            },
        },
    },
}
```

### MCP Server Types

```go
// Stdio server
claude.McpServerConfig{
    Command: "mcp-server",
    Args:    []string{"--port", "8080"},
}

// SSE server
claude.McpServerConfig{
    URL: "http://localhost:8080/sse",
}
```

## Permission Modes

Control CLI permissions:

```go
client.SetPermissionMode("default")    // Normal permissions
client.SetPermissionMode("plan")       // Read-only, no writes
client.SetPermissionMode("full")       // All permissions
```

## V2 Session Options

```go
import "github.com/dotcommander/agent-sdk-go/claude/v2"

session, err := v2.NewSession(ctx,
    v2.WithModel("claude-sonnet-4-20250514"),
    v2.WithSystemPrompt("You are a coding assistant"),
    v2.WithTimeout(60*time.Second),
    v2.WithResume("session-id"),
    v2.WithWorkingDirectory("/path/to/project"),
)
```

## Configuration Precedence

Options are applied in order, later options override earlier:

```go
client, _ := claude.NewClient(
    claude.WithModel("claude-3-5-haiku"),     // Set to haiku
    claude.WithModel("claude-3-5-sonnet"),    // Override to sonnet
)
// Final model: claude-3-5-sonnet
```

## Runtime Configuration

Some settings can be changed after client creation:

```go
client, _ := claude.NewClient()

// Change model
client.SetModel("claude-opus-4-5-20251101")

// Set session ID
client.SetSessionID("new-session")

// Change permission mode
client.SetPermissionMode("plan")
```

## Configuration Validation

The SDK validates configuration on client creation:

```go
client, err := claude.NewClient(
    claude.WithTimeout("invalid"),
)
if err != nil {
    // Error: invalid timeout format
}
```

## Environment Variable Reference

| Variable | Description |
|----------|-------------|
| `ANTHROPIC_API_KEY` | API key (if using direct API) |
| `CLAUDE_CLI_PATH` | Custom CLI path |
| `CLAUDE_MODEL` | Default model |
| `HTTP_PROXY` | HTTP proxy URL |
| `HTTPS_PROXY` | HTTPS proxy URL |
| `NO_PROXY` | Proxy bypass list |

## Next Steps

- [Advanced Usage](advanced.md) - Low-level configuration
- [Troubleshooting](troubleshooting.md) - Configuration issues
