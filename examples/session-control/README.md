# Example: Session Control

## What This Demonstrates

This example shows session lifecycle and control patterns in the Claude Agent SDK. It demonstrates:

- Session creation and basic lifecycle
- Session resume and persistence
- Session limits (max turns, budget, thinking tokens)
- Session forking and continuation
- Graceful shutdown and interrupt handling
- Session state inspection

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

```bash
cd examples/session-control
go run main.go
```

## Expected Output

```
=== Session Control Example ===
This demonstrates session lifecycle and control patterns.

--- Basic Session Lifecycle ---
  Created session: session-1234567890
  Response:
    Hello

--- Session Resume ---
  Created session 1: session-1234567890
  Session 1 response: I'll remember that the secret word is "banana".
  Session 1 closed
  Resuming session...
  Resumed session 2: session-1234567890
  Session 2 response: The secret word is "banana".

--- Session Limits ---
  WithMaxTurns(n):
    Maximum number of conversation turns
    Example: v2.WithMaxTurns(10) // Limit to 10 exchanges

  WithMaxBudget(usd):
    Maximum cost in USD
    Example: v2.WithMaxBudget(1.00) // Limit to $1.00
...
```

## Key Patterns

### Pattern 1: Basic Session Lifecycle

Create, use, and close a session:

```go
// Create session
session, err := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithTimeout(60*time.Second),
)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// Get session ID for later resume
sessionID := session.SessionID()

// Send message
session.Send(ctx, "Hello!")

// Receive response
for msg := range session.Receive(ctx) {
    if msg.Type() == v2.V2EventTypeAssistant {
        fmt.Print(v2.ExtractAssistantText(msg))
    }
}
```

### Pattern 2: Session Resume

Resume a previous session to continue the conversation:

```go
// Save session ID from previous session
sessionID := "session-123456"

// Resume the session
session, err := v2.ResumeSession(ctx, sessionID,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithTimeout(60*time.Second),
)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// Check if session was successfully resumed
if session.IsResumed() {
    fmt.Println("Resumed existing session")
}

// Continue conversation with preserved context
session.Send(ctx, "What were we discussing?")
```

### Pattern 3: Session Limits

Configure limits to control costs and scope:

```go
session, err := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),

    // Conversation limits
    v2.WithMaxTurns(10),              // Max 10 exchanges
    v2.WithMaxBudgetUSD(1.00),        // Max $1.00 spend
    v2.WithMaxThinkingTokens(4096),   // Limit reasoning

    // Time limits
    v2.WithTimeout(5*time.Minute),    // Overall timeout
)
```

### Pattern 4: Session Persistence

Control whether sessions are saved to disk:

```go
// Persistent session (default)
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithPersistSession(true),  // Save to disk
)

// Ephemeral session
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithPersistSession(false), // Don't save
)
```

### Pattern 5: Session Forking

Fork a session to try different approaches:

```go
// Original session
session1, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
)
session1.Send(ctx, "Establish context here")
// ... receive response ...
sessionID := session1.SessionID()
session1.Close()

// Fork the session (creates a copy)
session2, _ := v2.ResumeSession(ctx, sessionID,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithForkSession(true),  // Fork instead of continue
)
// session2 has the same context but won't affect original
```

### Pattern 6: Graceful Shutdown

Handle interrupts cleanly:

```go
// Setup signal handling
ctx, cancel := signal.NotifyContext(context.Background(),
    syscall.SIGINT, syscall.SIGTERM)
defer cancel()

session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
)
defer session.Close()  // Always close

session.Send(ctx, prompt)

// Receive with cancellation support
for msg := range session.Receive(ctx) {
    select {
    case <-ctx.Done():
        fmt.Println("Interrupted, cleaning up...")
        return
    default:
        // Process message
    }
}
```

### Pattern 7: Session State Inspection

Query session information:

```go
// Get session ID
fmt.Printf("Session ID: %s\n", session.SessionID())

// Check if resumed
fmt.Printf("Is Resumed: %v\n", session.IsResumed())

// String representation
fmt.Printf("Session: %s\n", session.String())
// Output: V2Session{active: id="session-123", model=claude-sonnet-4-5-20250929}

// Access underlying client for advanced operations
client := session.GetClient()
```

## Session Options Reference

| Option | Type | Description |
|--------|------|-------------|
| `WithModel` | `string` | Claude model to use |
| `WithTimeout` | `time.Duration` | Operation timeout |
| `WithMaxTurns` | `int` | Max conversation turns |
| `WithMaxBudgetUSD` | `float64` | Max cost in USD |
| `WithMaxThinkingTokens` | `int` | Max reasoning tokens |
| `WithPersistSession` | `bool` | Save session to disk |
| `WithContinue` | `bool` | Continue most recent session |
| `WithResume` | `string` | Session ID to resume |
| `WithForkSession` | `bool` | Fork instead of continue |
| `WithResumeSessionAt` | `string` | Resume at specific message |

## Session Methods Reference

| Method | Returns | Description |
|--------|---------|-------------|
| `Send(ctx, msg)` | `error` | Queue message for sending |
| `Receive(ctx)` | `<-chan V2Message` | Receive responses |
| `Close()` | `error` | Close session |
| `SessionID()` | `string` | Get session ID |
| `IsResumed()` | `bool` | Check if resumed |
| `String()` | `string` | String representation |
| `GetClient()` | `claude.Client` | Get underlying client |

## TypeScript Equivalent

This ports session patterns from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/session-control

The TypeScript version uses:
```typescript
const session = await createSession({
    model: "claude-sonnet-4-5-20250929",
    maxTurns: 10,
    maxBudget: 1.00,
});

// Resume
const resumed = await resumeSession(sessionId);

// Fork
const forked = await resumeSession(sessionId, { fork: true });
```

## Related Documentation

- [Session Reference](../../docs/usage.md#sessions)
- [Configuration Options](../../docs/usage.md#options)
- [Error Handling](../../docs/usage.md#error-handling)
