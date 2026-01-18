# Example: Basic Query

## What This Demonstrates

This example shows the simplest way to interact with Claude using the agent-sdk-go: a one-shot query that sends a prompt and receives a complete response.

The `v2.Prompt()` function is equivalent to `unstable_v2_prompt()` in the TypeScript SDK.

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

```bash
cd examples/basic-query
go run main.go
```

Or with a custom prompt:

```bash
go run main.go "What is 2 + 2?"
```

## Expected Output

```
=== Basic Query Example ===
Prompt: What is the capital of France? Answer in one word.

Response:
----------------------------------------
Paris
----------------------------------------

Session ID: prompt-1737175678901234567
Duration: 1.234s
```

## Key Patterns

### Pattern 1: One-Shot Prompt

The simplest way to query Claude - send a prompt, get a response:

```go
result, err := v2.Prompt(ctx, "Your question here",
    v2.WithPromptModel("claude-sonnet-4-20250514"),
    v2.WithPromptTimeout(60*time.Second),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Result)
```

### Pattern 2: CLI Availability Check

Always check that the Claude CLI is available before making queries:

```go
if !cli.IsCLIAvailable() {
    log.Fatal("Claude CLI not found")
}
```

### Pattern 3: Context with Timeout

Use Go's context for cancellation and timeouts:

```go
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()
```

## TypeScript Equivalent

This ports the TypeScript example from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world-v2

TypeScript:
```typescript
const result = await unstable_v2_prompt('What is the capital of France?', { model: 'sonnet' });
if (result.subtype === 'success') {
    console.log(`Answer: ${result.result}`);
}
```

Go:
```go
result, err := v2.Prompt(ctx, "What is the capital of France?",
    v2.WithPromptModel("claude-sonnet-4-20250514"),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Result)
```

## Related Documentation

- [V2 API Overview](../../docs/usage.md)
- [TypeScript SDK Reference](https://docs.anthropic.com/en/docs/claude-code/sdk/sdk-overview)
