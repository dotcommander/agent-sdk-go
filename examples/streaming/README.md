# Example: Streaming Responses

## What This Demonstrates

This example shows how to receive streaming responses from Claude, displaying text as it's generated rather than waiting for the complete response. This is essential for building responsive UIs and real-time applications.

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

```bash
cd examples/streaming
go run main.go
```

Or with a custom prompt:

```bash
go run main.go "Write a haiku about Go programming"
```

## Expected Output

```
=== Streaming Example ===
Prompt: Tell me a short story about a robot learning to paint.

Response (streaming):
----------------------------------------
In a quiet workshop, a small robot named Unit-7 watched an old painter
create masterpieces. One day, the painter handed Unit-7 a brush...
[text appears incrementally as Claude generates it]
----------------------------------------

Session ID: session-1737175678901234567
Characters: 847
Duration: 3.456s
Speed: 245.2 chars/sec
```

## Key Patterns

### Pattern 1: Session-Based Streaming

Create a session and stream responses via channels:

```go
session, err := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-20250514"),
    v2.WithEnablePartialMessages(true),
)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

session.Send(ctx, "Your prompt")

for msg := range session.Receive(ctx) {
    switch msg.Type() {
    case v2.V2EventTypeAssistant:
        fmt.Print(v2.ExtractAssistantText(msg))
    case v2.V2EventTypeStreamDelta:
        fmt.Print(v2.ExtractDeltaText(msg))
    case v2.V2EventTypeError:
        log.Fatal(v2.ExtractErrorMessage(msg))
    }
}
```

### Pattern 2: Message Type Handling

Different message types carry different information:

| Type | Purpose | Extraction |
|------|---------|------------|
| `V2EventTypeAssistant` | Complete text blocks | `v2.ExtractAssistantText(msg)` |
| `V2EventTypeStreamDelta` | Incremental text | `v2.ExtractDeltaText(msg)` |
| `V2EventTypeResult` | Final result | `v2.ExtractResultText(msg)` |
| `V2EventTypeError` | Error occurred | `v2.ExtractErrorMessage(msg)` |

### Pattern 3: Graceful Cancellation

Use signal handling for clean shutdown:

```go
ctx, cancel := signal.NotifyContext(context.Background(),
    syscall.SIGINT, syscall.SIGTERM)
defer cancel()
```

### Pattern 4: Iterator Alternative

For non-channel preference, use the iterator pattern:

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
    fmt.Print(v2.ExtractText(msg))
}
```

## TypeScript Equivalent

This ports the streaming pattern from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world-v2

TypeScript:
```typescript
await using session = unstable_v2_createSession({ model: 'sonnet' });
await session.send('Hello!');

for await (const msg of session.stream()) {
    if (msg.type === 'assistant') {
        const text = msg.message.content.find(c => c.type === 'text');
        console.log(text?.text);
    }
}
```

Go:
```go
session, _ := v2.CreateSession(ctx, v2.WithModel("claude-sonnet-4-20250514"))
defer session.Close()

session.Send(ctx, "Hello!")

for msg := range session.Receive(ctx) {
    if msg.Type() == v2.V2EventTypeAssistant {
        fmt.Print(v2.ExtractAssistantText(msg))
    }
}
```

## Related Documentation

- [V2 Session API](../../docs/usage.md)
- [Channel-Based Streaming](../../claude/v2/types.go)
