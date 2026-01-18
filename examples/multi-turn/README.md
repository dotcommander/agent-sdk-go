# Example: Multi-Turn Conversation

## What This Demonstrates

This example shows how to maintain context across multiple conversation turns with Claude. The V2 Session API preserves conversation history, allowing Claude to reference previous messages and maintain coherent dialogue.

This is the key advantage of the V2 API over one-shot prompts.

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

### Demo Mode (automated)

Shows context preservation with predefined messages:

```bash
cd examples/multi-turn
go run main.go demo
```

### Interactive Mode

Chat with Claude interactively:

```bash
go run main.go
```

## Expected Output (Demo Mode)

```
=== Multi-Turn Demo ===
This demonstrates how Claude maintains context across conversation turns.

Session ID: session-1737175678901234567

--- Turn 1 ---
User: My name is Alex and my favorite number is 42. Please acknowledge.
Claude: Hello Alex! I've noted that your favorite number is 42. Nice to meet you!

--- Turn 2 ---
User: What is my name?
Claude: Your name is Alex.

--- Turn 3 ---
User: What is my favorite number multiplied by 2?
Claude: Your favorite number is 42, so 42 x 2 = 84.

=== Demo Complete ===
Claude maintained context across all three turns!
```

## Key Patterns

### Pattern 1: Session Lifecycle

Create a session, use it for multiple turns, then close:

```go
session, err := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-20250514"),
    v2.WithTimeout(60*time.Second),
)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// Turn 1
session.Send(ctx, "My name is Alex")
for msg := range session.Receive(ctx) { /* process */ }

// Turn 2 - Claude remembers Turn 1
session.Send(ctx, "What is my name?")
for msg := range session.Receive(ctx) { /* Claude says "Alex" */ }
```

### Pattern 2: Send and Receive Helper

Collect a complete response into a string:

```go
func sendAndReceive(ctx context.Context, session v2.V2Session, message string) string {
    session.Send(ctx, message)

    var result strings.Builder
    for msg := range session.Receive(ctx) {
        result.WriteString(v2.ExtractText(msg))
    }
    return result.String()
}
```

### Pattern 3: Session Management

Create new sessions on demand while properly closing old ones:

```go
// Close current session
session.Close()

// Create a new session (fresh context)
session, err = v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-20250514"),
)
```

### Pattern 4: Interactive Commands

Handle special commands in an interactive loop:

```go
switch strings.ToLower(input) {
case "quit", "exit":
    return
case "new":
    // Start fresh session
case "id":
    fmt.Println(session.SessionID())
default:
    // Send to Claude
    session.Send(ctx, input)
}
```

## TypeScript Equivalent

This ports the multi-turn pattern from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world-v2

TypeScript:
```typescript
await using session = unstable_v2_createSession({ model: 'sonnet' });

// Turn 1
await session.send('What is 5 + 3? Just the number.');
for await (const msg of session.stream()) { /* 8 */ }

// Turn 2 - Claude remembers
await session.send('Multiply that by 2. Just the number.');
for await (const msg of session.stream()) { /* 16 */ }
```

Go:
```go
session, _ := v2.CreateSession(ctx, v2.WithModel("claude-sonnet-4-20250514"))
defer session.Close()

// Turn 1
session.Send(ctx, "What is 5 + 3? Just the number.")
for msg := range session.Receive(ctx) { /* 8 */ }

// Turn 2 - Claude remembers
session.Send(ctx, "Multiply that by 2. Just the number.")
for msg := range session.Receive(ctx) { /* 16 */ }
```

## Session Resume (Advanced)

The SDK supports resuming sessions by ID:

```go
// Save the session ID
sessionID := session.SessionID()

// Later, resume the session
resumedSession, err := v2.ResumeSession(ctx, sessionID,
    v2.WithModel("claude-sonnet-4-20250514"),
)
```

Note: Session resume requires the Claude CLI to support session persistence.

## Related Documentation

- [Session API Reference](../../claude/v2/session.go)
- [V2 Types](../../claude/v2/types.go)
