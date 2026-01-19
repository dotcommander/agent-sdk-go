# Streaming

Streaming allows you to process Claude's response in real-time as it's generated, rather than waiting for the complete response.

## Why Stream?

- **Faster perceived response** - Show output immediately
- **Progressive UI updates** - Update your interface as content arrives
- **Memory efficient** - Process large responses without buffering
- **Early termination** - Cancel if the response isn't what you need

## Basic Streaming

```go
client, _ := claude.NewClient()
ctx := context.Background()

msgChan, errChan := client.QueryStream(ctx, "Write a poem about coding")

for {
    select {
    case msg, ok := <-msgChan:
        if !ok {
            return // Stream complete
        }
        if text := claude.GetContentText(msg); text != "" {
            fmt.Print(text) // Print as it arrives
        }
    case err := <-errChan:
        if err != nil {
            log.Fatal(err)
        }
        return
    }
}
```

## Channel Semantics

### Message Channel (`msgChan`)

- Receives parsed message objects
- Closes when the response is complete
- May receive multiple message types

### Error Channel (`errChan`)

- Receives errors during streaming
- Typically only one error (if any)
- Check after `msgChan` closes for final errors

## Handling Different Message Types

```go
import "github.com/dotcommander/agent-sdk-go/claude"

msgChan, errChan := client.QueryStream(ctx, prompt)

for msg := range msgChan {
    switch m := msg.(type) {
    case *claude.AssistantMessage:
        // Main content - may arrive in chunks
        for _, block := range m.Content {
            if tb, ok := block.(*claude.TextBlock); ok {
                fmt.Print(tb.Text)
            }
        }

    case *claude.ToolUseMessage:
        // Claude wants to use a tool
        fmt.Printf("Tool: %s\n", m.Name)
        fmt.Printf("Input: %v\n", m.Input)

    case *claude.ResultMessage:
        // Response complete
        fmt.Printf("\nTokens used: %d\n", m.Usage.TotalTokens)

    case *claude.ErrorMessage:
        // Error during generation
        log.Printf("Error: %s\n", m.Error)
    }
}

// Check for streaming errors
for err := range errChan {
    log.Printf("Stream error: %v\n", err)
}
```

## Collecting Full Response

To stream but also collect the full response:

```go
func streamAndCollect(client claude.Client, ctx context.Context, prompt string) (string, error) {
    msgChan, errChan := client.QueryStream(ctx, prompt)

    var builder strings.Builder

    for {
        select {
        case msg, ok := <-msgChan:
            if !ok {
                return builder.String(), nil
            }
            if text := claude.GetContentText(msg); text != "" {
                fmt.Print(text)        // Stream to console
                builder.WriteString(text) // Collect
            }
        case err := <-errChan:
            return builder.String(), err
        }
    }
}
```

## Cancellation

Cancel streaming using context:

```go
ctx, cancel := context.WithCancel(context.Background())

msgChan, errChan := client.QueryStream(ctx, "Write a very long essay...")

go func() {
    time.Sleep(5 * time.Second)
    cancel() // Cancel after 5 seconds
}()

for {
    select {
    case msg, ok := <-msgChan:
        if !ok {
            return
        }
        fmt.Print(claude.GetContentText(msg))
    case err := <-errChan:
        if errors.Is(err, context.Canceled) {
            fmt.Println("\n[Cancelled]")
            return
        }
        log.Fatal(err)
    case <-ctx.Done():
        fmt.Println("\n[Stopped]")
        return
    }
}
```

## Timeout

Add a timeout to streaming:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

msgChan, errChan := client.QueryStream(ctx, prompt)

for {
    select {
    case msg, ok := <-msgChan:
        if !ok {
            return
        }
        processMessage(msg)
    case err := <-errChan:
        log.Fatal(err)
    case <-ctx.Done():
        if ctx.Err() == context.DeadlineExceeded {
            log.Println("Response timed out")
        }
        return
    }
}
```

## Progress Indicators

Show a progress indicator while streaming:

```go
func streamWithProgress(client claude.Client, ctx context.Context, prompt string) {
    msgChan, errChan := client.QueryStream(ctx, prompt)

    spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
    spinIdx := 0
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()

    started := false

    for {
        select {
        case msg, ok := <-msgChan:
            if !ok {
                fmt.Println()
                return
            }
            if text := claude.GetContentText(msg); text != "" {
                if !started {
                    fmt.Print("\r\033[K") // Clear spinner
                    started = true
                }
                fmt.Print(text)
            }
        case err := <-errChan:
            if err != nil {
                fmt.Printf("\nError: %v\n", err)
            }
            return
        case <-ticker.C:
            if !started {
                fmt.Printf("\r%s Thinking...", spinner[spinIdx])
                spinIdx = (spinIdx + 1) % len(spinner)
            }
        }
    }
}
```

## Buffered Processing

Process messages in batches:

```go
func streamBuffered(client claude.Client, ctx context.Context, prompt string) {
    msgChan, errChan := client.QueryStream(ctx, prompt)

    buffer := make([]string, 0, 10)
    flushInterval := 500 * time.Millisecond
    ticker := time.NewTicker(flushInterval)
    defer ticker.Stop()

    flush := func() {
        if len(buffer) > 0 {
            fmt.Print(strings.Join(buffer, ""))
            buffer = buffer[:0]
        }
    }

    for {
        select {
        case msg, ok := <-msgChan:
            if !ok {
                flush()
                return
            }
            if text := claude.GetContentText(msg); text != "" {
                buffer = append(buffer, text)
            }
        case err := <-errChan:
            flush()
            if err != nil {
                log.Fatal(err)
            }
            return
        case <-ticker.C:
            flush()
        }
    }
}
```

## Concurrent Streams

Run multiple streams concurrently:

```go
func parallelQueries(client claude.Client, ctx context.Context, prompts []string) []string {
    results := make([]string, len(prompts))
    var wg sync.WaitGroup
    var mu sync.Mutex

    for i, prompt := range prompts {
        wg.Add(1)
        go func(idx int, p string) {
            defer wg.Done()

            var builder strings.Builder
            msgChan, errChan := client.QueryStream(ctx, p)

            for {
                select {
                case msg, ok := <-msgChan:
                    if !ok {
                        mu.Lock()
                        results[idx] = builder.String()
                        mu.Unlock()
                        return
                    }
                    builder.WriteString(claude.GetContentText(msg))
                case err := <-errChan:
                    if err != nil {
                        mu.Lock()
                        results[idx] = fmt.Sprintf("Error: %v", err)
                        mu.Unlock()
                    }
                    return
                }
            }
        }(i, prompt)
    }

    wg.Wait()
    return results
}
```

## Best Practices

### 1. Always Handle Both Channels

```go
// Good
for {
    select {
    case msg, ok := <-msgChan:
        if !ok { return }
        process(msg)
    case err := <-errChan:
        handle(err)
        return
    }
}

// Bad - may miss errors
for msg := range msgChan {
    process(msg)
}
```

### 2. Use Context for Control

```go
// Good - can cancel
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
defer cancel()
msgChan, errChan := client.QueryStream(ctx, prompt)

// Bad - no way to stop
msgChan, errChan := client.QueryStream(context.Background(), prompt)
```

### 3. Don't Block the Receive Loop

```go
// Good - process quickly or async
case msg, ok := <-msgChan:
    go processAsync(msg)

// Bad - blocks streaming
case msg, ok := <-msgChan:
    time.Sleep(time.Second) // Blocks!
    process(msg)
```

## Next Steps

- [Sessions](sessions.md) - Multi-turn conversations
- [Advanced Usage](advanced.md) - Low-level transport
- [Troubleshooting](troubleshooting.md) - Common issues
