# Example: Error Handling

## What This Demonstrates

This example shows comprehensive error handling patterns for the Claude Agent SDK. It demonstrates:

- CLI availability checking before operations
- Error type discrimination using the type-safe error system
- Context cancellation and timeout handling
- Error wrapping and unwrapping with `errors.Is`/`errors.As`
- Circuit breaker pattern for resilience
- Graceful degradation with fallback strategies

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

```bash
cd examples/error-handling
go run main.go
```

## Expected Output

```
=== Error Handling Example ===
This demonstrates graceful error handling patterns in the SDK.

--- CLI Availability Check ---
Claude CLI is available
CLI path: /usr/local/bin/claude

--- Error Type Discrimination ---
  Error: CLI not found at /usr/local/bin/claude (command: claude)
    Type: CLINotFoundError
    Recovery: Install Claude CLI or check PATH

  Error: connection error: failed to establish connection
    Type: ConnectionError
    Recovery: Retry with backoff or check network

  Error: timeout during query after 30s
    Type: TimeoutError
    Recovery: Increase timeout or reduce query complexity
...

--- Context Cancellation ---
  Operation was cancelled (expected)
  Recovery: Clean up resources, notify user

--- Timeout Handling ---
  Operation timed out (expected)
  Recovery options:
    1. Increase timeout duration
    2. Reduce query complexity
    3. Use streaming for long responses

--- Error Wrapping ---
  Top-level error: query failed: session creation failed: connection error: network unreachable
  Unwrapping chain:
    Found ConnectionError in chain
    Original error: connection error: network unreachable

--- Circuit Breaker Pattern ---
  Simulating failures to trip circuit:
    Attempt 1: State=closed, Error=true
    Attempt 2: State=closed, Error=true
    Attempt 3: State=open, Error=true
  Waiting for recovery timeout...
  State after recovery: half-open
  After success: State=closed, Error=false

--- Graceful Degradation ---
  Session created: session-1234567890
  Response: Hello! How can I help you today?

=== Error Handling Example Complete ===
```

## Key Patterns

### Pattern 1: CLI Availability Check

Always check CLI availability before operations:

```go
if !cli.IsCLIAvailable() {
    log.Fatal("Claude CLI not found")
}
cliPath := cli.GetCLIPath()
fmt.Printf("CLI path: %s\n", cliPath)
```

### Pattern 2: Error Type Discrimination

Use type-safe error checking with helper functions:

```go
switch {
case shared.IsCLINotFound(err):
    // Install CLI or check PATH
case shared.IsConnectionError(err):
    // Retry with backoff
case shared.IsTimeoutError(err):
    // Increase timeout or simplify query
case shared.IsParserError(err):
    // Check CLI version compatibility
case shared.IsProtocolError(err):
    // Check SDK version
default:
    // Log and report unknown error
}
```

### Pattern 3: Context Cancellation

Properly handle context cancellation:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

err := longRunningOperation(ctx)
if errors.Is(err, context.Canceled) {
    // Clean up resources
}
```

### Pattern 4: Timeout Handling

Use context timeouts with proper error checking:

```go
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()

result, err := session.Query(ctx, prompt)
if errors.Is(err, context.DeadlineExceeded) {
    // Consider increasing timeout or streaming
}
```

### Pattern 5: Error Wrapping

Wrap errors with context for debugging:

```go
result, err := session.Query(ctx, prompt)
if err != nil {
    return fmt.Errorf("query failed: %w", err)
}
```

And unwrap to find root causes:

```go
var connErr *shared.ConnectionError
if errors.As(err, &connErr) {
    // Handle connection-specific recovery
}
```

### Pattern 6: Circuit Breaker

Use circuit breaker for resilience:

```go
cb := shared.NewStubCircuitBreaker(shared.CircuitBreakerConfig{
    FailureThreshold:    3,
    RecoveryTimeout:     30 * time.Second,
    HalfOpenMaxRequests: 1,
})

err := cb.Execute(ctx, func() error {
    return session.Query(ctx, prompt)
})
```

### Pattern 7: Graceful Degradation

Implement fallback strategies:

```go
func createSessionWithFallback(ctx context.Context) (v2.V2Session, error) {
    // Try primary model
    session, err := v2.CreateSession(ctx, v2.WithModel("claude-sonnet-4-5-20250929"))
    if err == nil {
        return session, nil
    }
    log.Printf("Primary failed: %v, trying fallback", err)

    // Try fallback model
    return v2.CreateSession(ctx, v2.WithModel("claude-3-5-haiku-20241022"))
}
```

### Pattern 8: Retry with Exponential Backoff

```go
func queryWithRetry(ctx context.Context, session v2.V2Session, prompt string, maxRetries int) (string, error) {
    var lastErr error
    for attempt := 0; attempt < maxRetries; attempt++ {
        if attempt > 0 {
            backoff := time.Duration(1<<uint(attempt-1)) * 100 * time.Millisecond
            select {
            case <-time.After(backoff):
            case <-ctx.Done():
                return "", ctx.Err()
            }
        }

        response, err := executeQuery(ctx, session, prompt)
        if err == nil {
            return response, nil
        }
        lastErr = err

        // Don't retry on context errors
        if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
            return "", err
        }
    }
    return "", fmt.Errorf("query failed after %d attempts: %w", maxRetries, lastErr)
}
```

## Error Types Reference

| Error Type | Description | Recovery Strategy |
|------------|-------------|-------------------|
| `CLINotFoundError` | Claude CLI not installed | Install CLI, check PATH |
| `ConnectionError` | Failed to connect to CLI | Retry with backoff |
| `TimeoutError` | Operation timed out | Increase timeout, use streaming |
| `ParserError` | Failed to parse CLI output | Check CLI version compatibility |
| `ProtocolError` | Unexpected protocol message | Check SDK version |
| `ConfigurationError` | Invalid configuration | Fix configuration values |
| `ProcessError` | CLI process error | Check CLI status, restart |

## TypeScript Equivalent

This ports error handling patterns from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/error-handling

The TypeScript version uses:
```typescript
try {
    const response = await client.query(prompt);
} catch (err) {
    if (err instanceof CLINotFoundError) {
        // Handle CLI not found
    } else if (err instanceof ConnectionError) {
        // Handle connection error
    }
}
```

## Related Documentation

- [Error Handling Guide](../../docs/usage.md#error-handling)
- [Circuit Breaker Pattern](../../docs/usage.md#resilience)
- [Session Management](../../docs/usage.md#sessions)
