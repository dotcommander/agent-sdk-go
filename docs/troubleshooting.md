# Troubleshooting

Solutions for common issues with the Claude Code Agent SDK.

## Quick Diagnostics

Run these commands to diagnose issues:

```bash
# Check Claude CLI is installed
claude --version

# Verify authentication
claude "Hello, respond with just 'OK'"

# Check Go version
go version

# Verify SDK installation
go list -m github.com/dotcommander/agent-sdk-go
```

## Common Errors

### "claude: command not found"

**Cause:** Claude Code CLI is not installed or not in PATH.

**Solutions:**

1. Install Claude Code CLI:
   ```bash
   # Follow instructions at:
   # https://docs.anthropic.com/en/docs/claude-code
   ```

2. Add to PATH:
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export PATH="$PATH:$HOME/.local/bin"
   ```

3. Specify path explicitly:
   ```go
   client, err := claude.NewClient(
       claude.WithCLIPath("/full/path/to/claude"),
   )
   ```

### "context deadline exceeded"

**Cause:** Query took longer than the timeout.

**Solutions:**

1. Increase timeout:
   ```go
   client, err := claude.NewClient(
       claude.WithTimeout("300s"), // 5 minutes
   )
   ```

2. Use context timeout:
   ```go
   ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
   defer cancel()
   response, err := client.Query(ctx, prompt)
   ```

3. Use streaming for long responses:
   ```go
   msgChan, errChan := client.QueryStream(ctx, prompt)
   // Process incrementally
   ```

### "not authenticated"

**Cause:** Claude CLI is not authenticated.

**Solution:**

```bash
# Run claude manually to authenticate
claude "Hello"
# Follow the authentication prompts
```

### "already connected"

**Cause:** Calling `Connect()` on an already-connected client.

**Solution:**

```go
// Check before connecting
if client.GetSessionID() == "" {
    err := client.Connect(ctx)
}

// Or disconnect first
client.Disconnect()
err := client.Connect(ctx)
```

### "not connected"

**Cause:** Calling session methods without connecting.

**Solution:**

```go
// Ensure connected before receiving
err := client.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Disconnect()

msgChan, errChan := client.ReceiveMessages(ctx)
```

### Empty Response

**Cause:** Various parsing or communication issues.

**Debugging:**

```go
msgChan, errChan := client.QueryStream(ctx, prompt)

for {
    select {
    case msg, ok := <-msgChan:
        if !ok {
            fmt.Println("Channel closed")
            return
        }
        // Log all messages to debug
        fmt.Printf("Message type: %T\n", msg)
        fmt.Printf("Message content: %+v\n", msg)

    case err := <-errChan:
        fmt.Printf("Error: %v\n", err)
        return
    }
}
```

### JSON Parse Errors

**Cause:** Unexpected output format from CLI.

**Debugging:**

```go
import "github.com/dotcommander/agent-sdk-go/claude/subprocess"

config := &subprocess.TransportConfig{
    Model:      "claude-sonnet-4-20250514",
    CustomArgs: []string{"--verbose"},
}
```

Check CLI output directly:
```bash
claude --output-format stream-json -p "Hello"
```

## Performance Issues

### Slow First Query

**Cause:** Claude CLI initialization overhead.

**Solution:** The first query is always slower due to CLI startup. Subsequent queries in a session are faster.

```go
// Pre-warm with a simple query
client.Query(ctx, "Hello")

// Now run actual queries
response, err := client.Query(ctx, actualPrompt)
```

### High Memory Usage

**Cause:** Not consuming message channels.

**Solution:** Always read from channels:

```go
// Bad - channels may buffer
msgChan, errChan := client.QueryStream(ctx, prompt)
// Not reading from channels!

// Good - consume all messages
for {
    select {
    case msg, ok := <-msgChan:
        if !ok {
            return
        }
        _ = msg // At minimum, drain the channel
    case err := <-errChan:
        return
    }
}
```

### Many Concurrent Queries Slow

**Cause:** Each query spawns a new process.

**Solutions:**

1. Use sessions for related queries:
   ```go
   client.Connect(ctx)
   // Multiple queries share one process
   ```

2. Limit concurrency:
   ```go
   semaphore := make(chan struct{}, 5) // Max 5 concurrent

   for _, prompt := range prompts {
       semaphore <- struct{}{}
       go func(p string) {
           defer func() { <-semaphore }()
           client.Query(ctx, p)
       }(prompt)
   }
   ```

## Connection Issues

### Process Hangs

**Cause:** CLI waiting for input or stuck.

**Solutions:**

1. Use context with timeout:
   ```go
   ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
   defer cancel()
   ```

2. Interrupt the client:
   ```go
   go func() {
       time.Sleep(30 * time.Second)
       client.Interrupt()
   }()
   ```

### Session Disconnects

**Cause:** CLI process terminated unexpectedly.

**Solution:** Implement reconnection logic:

```go
func resilientQuery(client claude.Client, ctx context.Context, prompt string) (string, error) {
    for attempt := 0; attempt < 3; attempt++ {
        response, err := client.Query(ctx, prompt)
        if err == nil {
            return response, nil
        }

        // Check if we should retry
        if strings.Contains(err.Error(), "broken pipe") ||
           strings.Contains(err.Error(), "connection reset") {
            time.Sleep(time.Second)
            continue
        }

        return "", err
    }
    return "", errors.New("max retries exceeded")
}
```

## Debugging Tips

### Enable Verbose Logging

```go
// Add verbose flag
config := &subprocess.TransportConfig{
    CustomArgs: []string{"--verbose"},
}
```

### Log All Messages

```go
func debugStream(client claude.Client, ctx context.Context, prompt string) {
    msgChan, errChan := client.QueryStream(ctx, prompt)

    for {
        select {
        case msg, ok := <-msgChan:
            if !ok {
                log.Println("Stream complete")
                return
            }
            data, _ := json.MarshalIndent(msg, "", "  ")
            log.Printf("Message:\n%s\n", data)

        case err := <-errChan:
            log.Printf("Error: %v", err)
            return
        }
    }
}
```

### Check CLI Directly

Test the CLI independently:

```bash
# Simple test
claude "Say hello"

# JSON output mode (what SDK uses)
claude --output-format stream-json -p "Hello"

# With specific model
claude --model claude-sonnet-4-20250514 "Hello"

# Check for errors
claude "Hello" 2>&1 | head -20
```

### Environment Debugging

```go
// Print environment
for _, env := range os.Environ() {
    if strings.HasPrefix(env, "CLAUDE") || strings.HasPrefix(env, "ANTHROPIC") {
        fmt.Println(env)
    }
}

// Check PATH
fmt.Println("PATH:", os.Getenv("PATH"))

// Check which claude
if path, err := exec.LookPath("claude"); err == nil {
    fmt.Println("Claude at:", path)
}
```

## Getting Help

### Collect Diagnostic Information

When reporting issues, include:

```go
func collectDiagnostics() {
    fmt.Println("=== SDK Diagnostics ===")

    // Go version
    fmt.Println("Go version:", runtime.Version())

    // OS/Arch
    fmt.Println("OS/Arch:", runtime.GOOS, runtime.GOARCH)

    // Claude CLI version
    cmd := exec.Command("claude", "--version")
    output, _ := cmd.Output()
    fmt.Println("Claude CLI:", string(output))

    // SDK version (from go.mod)
    fmt.Println("SDK: github.com/dotcommander/agent-sdk-go")
}
```

### Resources

- [GitHub Issues](https://github.com/dotcommander/agent-sdk-go/issues)
- [Claude Code Documentation](https://docs.anthropic.com/en/docs/claude-code)

### Filing a Bug Report

Include:
1. SDK version
2. Go version
3. OS and architecture
4. Claude CLI version
5. Minimal reproducible code
6. Expected vs actual behavior
7. Error messages (full text)
