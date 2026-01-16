# Data Flow and Control Flow Analysis

**Project:** agent-sdk-go
**Date:** 2026-01-16

---

## Overview

This document traces data and control flow through the agent-sdk-go codebase, covering:
1. **Message Flow:** How data moves from user input to Claude CLI and back
2. **Control Flow:** Lifecycle management and state transitions
3. **Concurrency Flow:** Goroutine orchestration and synchronization
4. **Error Flow:** How errors propagate through layers

---

## Architecture Layers

```
┌────────────────────────────────────────────────────────┐
│ Layer 5: Application (cmd/)                            │  User code
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ Layer 4: Client API (internal/claude/client.go)        │  High-level interface
│          + V2 Sessions (internal/claude/v2/session.go)  │
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ Layer 3: Message Handling (internal/claude/shared/)    │  Type system
│          + Parsing (internal/claude/parser/)            │
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ Layer 2: Transport (internal/claude/subprocess/)        │  Process I/O
└─────────────────┬──────────────────────────────────────┘
                  │
┌─────────────────▼──────────────────────────────────────┐
│ Layer 1: Claude CLI (external subprocess)              │  Binary execution
└────────────────────────────────────────────────────────┘
```

---

## Flow 1: One-Shot Query (V1 API)

### Entry Point: `client.Query(ctx, prompt)`

```go
// cmd/agent/main.go:92
client.Query(ctx, "Hello, Claude!")
```

### Flow Diagram

```
User Code
    │
    ├─ client.Query(ctx, "Hello")
    │      │
    │      ├─ client.QueryStream(ctx, "Hello")  ← Delegates to streaming
    │      │      │
    │      │      ├─ subprocess.NewTransportWithPrompt(config, "Hello")
    │      │      │      │
    │      │      │      └─ Transport{promptArg: &"Hello"}  ← One-shot mode
    │      │      │
    │      │      ├─ transport.Connect(ctx)
    │      │      │      │
    │      │      │      ├─ cli.DiscoverCLI() → /usr/local/bin/claude
    │      │      │      ├─ exec.CommandContext(ctx, "claude", "--print", "Hello", ...)
    │      │      │      ├─ cmd.Start()  ← Spawn subprocess
    │      │      │      ├─ go handleStdout()  ← Read goroutine
    │      │      │      └─ go handleStderr()  ← Error goroutine
    │      │      │
    │      │      └─ transport.ReceiveMessages(ctx) → (msgChan, errChan)
    │      │
    │      └─ Collect from channels into string
    │             │
    │             ├─ for msg := range msgChan:
    │             │      ├─ if AssistantMessage → extract text
    │             │      └─ result.WriteString(text)
    │             │
    │             └─ return result.String()
    │
    └─ "Response text"
```

### Data Transformations

1. **User Input → CLI Args**
   ```go
   "Hello, Claude!" → []string{"--print", "Hello, Claude!", "--output-format", "stream-json"}
   ```

2. **CLI Stdout → Raw JSON**
   ```
   {"type":"assistant","content":[{"type":"text","text":"Hello!"}]}\n
   ```

3. **Raw JSON → Go Struct**
   ```go
   &shared.AssistantMessage{
       MessageType: "assistant",
       Content: []shared.ContentBlock{
           &shared.TextBlock{Type: "text", Text: "Hello!"},
       },
   }
   ```

4. **Go Struct → User String**
   ```go
   "Hello!"  // Extracted from content blocks
   ```

---

## Flow 2: Streaming (V2 API)

### Entry Point: `session.Receive(ctx)`

```go
// V2 API usage
session, _ := v2.CreateSession(ctx)
session.Send(ctx, "Tell me a story")
for msg := range session.Receive(ctx) {
    fmt.Print(v2.ExtractText(msg))
}
```

### Flow Diagram

```
session.Send("story")
    │
    ├─ Store in session.pendingSend
    │
    └─ Return immediately (non-blocking)

session.Receive(ctx)
    │
    ├─ Check for pendingSend
    │      │
    │      └─ client.QueryStream(ctx, "story")  ← If pending
    │             │
    │             ├─ subprocess.NewTransportWithPrompt()
    │             ├─ transport.Connect(ctx)
    │             └─ return (msgChan, errChan)
    │
    ├─ wrapMessageChannel(msgChan, errChan)
    │      │
    │      ├─ go func() {
    │      │      for msg := range msgChan {
    │      │          v2Msg := convertToV2Message(msg, sessionID)
    │      │          out <- v2Msg  ← V2 wrapper
    │      │      }
    │      │  }()
    │      │
    │      └─ return out  ← V2Message channel
    │
    └─ User iterates: for msg := range out
```

### V2 Message Conversion

```go
// AssistantMessage → V2AssistantMessage
&shared.AssistantMessage{...}
    ↓ convertToV2Message()
&v2.V2AssistantMessage{
    TypeField: "assistant",
    Message: AssistantMessageContent{...},
    SessionID: "session-123456789",
}
```

---

## Flow 3: Subprocess Lifecycle

### Phase 1: Initialization

```
NewTransport(config)
    │
    ├─ Validate config
    ├─ Set defaults (model, timeout, CLI command)
    └─ Return &Transport{connected: false}
```

### Phase 2: Connection

```
transport.Connect(ctx)
    │
    ├─ Discover CLI path
    │      │
    │      └─ cli.DiscoverCLI("", "claude")
    │             │
    │             ├─ Check PATH for "claude"
    │             ├─ Try common locations
    │             └─ Return DiscoveryResult{Path: "/usr/local/bin/claude"}
    │
    ├─ Build command
    │      │
    │      └─ buildArgs() → ["--output-format", "stream-json", "--model", "..."]
    │
    ├─ Create exec.Cmd
    │      │
    │      └─ exec.CommandContext(ctx, cliPath, args...)
    │
    ├─ Set up pipes
    │      │
    │      ├─ cmd.StdinPipe() → t.stdin
    │      ├─ cmd.StdoutPipe() → t.stdout
    │      └─ cmd.StderrPipe() → t.stderr
    │
    ├─ Start process
    │      │
    │      └─ cmd.Start()  ← PID assigned
    │
    ├─ Initialize channels
    │      │
    │      ├─ t.msgChan = make(chan Message, 100)
    │      └─ t.errChan = make(chan error, 100)
    │
    ├─ Start goroutines
    │      │
    │      ├─ go handleStdout()  ← Read and parse JSON
    │      └─ go handleStderr()  ← Capture errors
    │
    └─ Set t.connected = true
```

### Phase 3: Message Exchange (Interactive Mode)

```
client.SendMessage(ctx, "message")
    │
    ├─ Create UserMessage
    │      │
    │      └─ &shared.UserMessage{Type: "user", Content: "message"}
    │
    ├─ Marshal to JSON
    │      │
    │      └─ {"type":"user","content":"message"}
    │
    └─ Write to stdin
           │
           └─ fmt.Fprintln(t.stdin, json)
                  │
                  └─ CLI receives via stdin
                         │
                         └─ CLI processes and writes to stdout

handleStdout() goroutine:
    │
    ├─ scanner.Scan()  ← Blocking read
    │
    ├─ Parse JSON line
    │      │
    │      ├─ json.Unmarshal(line, &rawMsg)
    │      └─ Discriminate by type field
    │
    ├─ Convert to typed message
    │      │
    │      ├─ "assistant" → parseAssistantMessage()
    │      ├─ "result" → parseResultMessage()
    │      ├─ "stream_event" → parseStreamEvent()
    │      └─ unknown → RawControlMessage
    │
    └─ Send to channel
           │
           └─ t.msgChan <- msg
```

### Phase 4: Shutdown

```
transport.Close()
    │
    ├─ Set connected = false
    │
    ├─ Cancel context
    │      │
    │      └─ t.cancel()  ← Signals goroutines to exit
    │
    ├─ Close stdin
    │      │
    │      └─ t.stdin.Close()  ← Sends EOF to CLI
    │
    ├─ Wait for goroutines (with timeout)
    │      │
    │      ├─ t.wg.Wait()  ← Up to 5 seconds
    │      └─ If timeout → force cleanup
    │
    ├─ Terminate process
    │      │
    │      ├─ cmd.Process.Kill()
    │      └─ cmd.Wait()
    │
    └─ Close channels
           │
           ├─ close(t.msgChan)
           └─ close(t.errChan)
```

---

## Flow 4: Error Propagation

### Error Sources

```
Layer 1 (CLI):
    ├─ Process crash → handleStderr() captures
    ├─ Invalid JSON output → handleStdout() parsing error
    └─ Non-zero exit → cmd.Wait() error

Layer 2 (Transport):
    ├─ Pipe errors → Read/Write failures
    ├─ Scanner errors → scanner.Err()
    └─ Timeout → context.DeadlineExceeded

Layer 3 (Parsing):
    ├─ JSON unmarshal failures → ParserError
    ├─ Type discrimination failures → ProtocolError
    └─ Validation errors → validation issues

Layer 4 (Client):
    ├─ Connection failures → ConnectionError
    ├─ Not connected → "not connected" error
    └─ CLI not found → CLINotFoundError

Layer 5 (Application):
    └─ User receives final error
```

### Error Flow Example

```
CLI exits unexpectedly
    │
    ├─ handleStderr() detects EOF
    │      │
    │      └─ errChan <- fmt.Errorf("CLI process exited")
    │
    ├─ User's receive loop
    │      │
    │      └─ err := <-errChan
    │             │
    │             └─ if err != nil { return err }
    │
    └─ Application error handling
           │
           └─ fmt.Fprintf(os.Stderr, "error: %v\n", err)
```

---

## Flow 5: Concurrency Patterns

### Goroutine Hierarchy

```
main()
    │
    ├─ client.Connect(ctx)
    │      │
    │      └─ transport.Connect(ctx)
    │             │
    │             ├─ cmd.Start()  ← Subprocess
    │             │
    │             ├─ go handleStdout()  ← Goroutine 1
    │             │      │
    │             │      └─ Reads stdout, parses JSON, sends to msgChan
    │             │
    │             └─ go handleStderr()  ← Goroutine 2
    │                    │
    │                    └─ Reads stderr, forwards to errChan
    │
    └─ client.Query(ctx, "prompt")
           │
           ├─ msgChan, errChan := client.QueryStream(ctx, "prompt")
           │
           └─ for { select { case msg := <-msgChan: ... } }  ← Main goroutine
```

### Synchronization Points

1. **Mutex Protection**
   ```go
   // client.go:43
   func (c *ClientImpl) Connect(ctx context.Context) error {
       c.mu.Lock()         // ← Acquire
       defer c.mu.Unlock() // ← Release
       // Critical section
   }
   ```

2. **Channel Communication**
   ```go
   // transport.go:349
   select {
   case t.msgChan <- msg:      // ← Send (may block)
   case <-t.ctx.Done():        // ← Cancellation
       return
   }
   ```

3. **WaitGroup for Cleanup**
   ```go
   // transport.go:209
   t.wg.Add(1)
   go t.handleStdout()

   // Later in Close():
   t.wg.Wait()  // ← Wait for goroutine to finish
   ```

### Race Conditions (Potential)

1. **pendingSend in V2 Session**
   ```go
   // v2/session.go:182
   pendingData := s.pendingSend  // ← Read without lock
   pendingDataCopy := pendingData // ← Copy
   s.mu.Unlock()                  // ← Then unlock

   // RACE: Another goroutine could modify pendingSend here
   if pendingDataCopy != nil && !pendingDataCopy.consumed {
       // Use pendingDataCopy
   }
   ```

2. **connected Flag Check**
   ```go
   // transport.go:467 (ReceiveMessages)
   t.mu.RLock()
   defer t.mu.RUnlock()

   if !t.connected {  // ← Read
       // Return error channels
   }

   return t.msgChan, t.errChan  // ← Return channels created during Connect
   ```
   If `Connect()` is called concurrently, channels could be nil.

---

## Flow 6: Configuration Loading

```
Application Start
    │
    ├─ app.Load()
    │      │
    │      ├─ Create empty config
    │      │
    │      ├─ Try config files (first wins):
    │      │      ├─ ./config.yaml
    │      │      ├─ ~/.config/agent-sdk-go/config.yaml
    │      │      └─ /etc/agent-sdk-go/config.yaml
    │      │
    │      ├─ Override with env vars:
    │      │      └─ ANTHROPIC_API_KEY → cfg.APIKey
    │      │
    │      └─ Return cfg
    │
    ├─ Create client with config
    │      │
    │      └─ claude.NewClient(
    │             WithModel(cfg.Model),
    │             WithTimeout(cfg.Timeout),
    │         )
    │
    └─ Use client
```

---

## Flow 7: Message Parsing

### Parser Registry Pattern

```
Initialization:
    │
    └─ parser.DefaultRegistry()
           │
           ├─ registry.Register("assistant", parseAssistantMessage)
           ├─ registry.Register("user", parseUserMessage)
           ├─ registry.Register("result", parseResultMessage)
           └─ registry.Register("stream_event", parseStreamEvent)

Parsing:
    │
    ├─ Raw JSON: `{"type":"assistant", ...}`
    │
    ├─ Unmarshal to map[string]any
    │
    ├─ Extract type field: "assistant"
    │
    ├─ registry.Lookup("assistant") → parseAssistantMessage
    │
    ├─ parseAssistantMessage(raw) → &AssistantMessage{...}
    │
    └─ Return typed message
```

---

## Key Takeaways

1. **Request Flow:** User → Client → Transport → CLI (subprocess) → Response
2. **Concurrency:** 2 goroutines per connection (stdout, stderr) + main goroutine
3. **Error Handling:** Errors propagate up via channels, wrapped with context
4. **Lifecycle:** Connect → Exchange → Close (with graceful shutdown)
5. **Parsing:** Registry-based message type discrimination
6. **Configuration:** File-based → env var override → client options

---

## Performance Characteristics

### Latency

- **Process spawn:** ~100-500ms (one-time per session)
- **Message roundtrip:** ~50-200ms (network + CLI processing)
- **JSON parsing:** ~1-5ms per message (negligible)

### Throughput

- **Sequential:** Limited by CLI subprocess (1 request at a time per session)
- **Parallel:** Can spawn multiple sessions (multiple processes)

### Memory

- **Channel buffers:** 100 messages × ~1KB = ~100KB per connection
- **Goroutines:** 2 per connection × ~2KB = ~4KB per connection
- **Process overhead:** ~10-50MB per Claude CLI subprocess

---

## Optimization Opportunities

1. **Process pooling:** Reuse CLI processes instead of spawning per session
2. **Zero-copy parsing:** Use `json.RawMessage` for delayed parsing
3. **Streaming aggregation:** Reduce channel sends by batching deltas
4. **Connection reuse:** Keep transport alive between queries (V1 API supports this)

---

## Anti-Patterns in Flow

1. **Synchronous Query wraps Streaming:** `Query()` internally uses `QueryStream()` and blocks collecting results. Duplication.
2. **V2 Session stores pending send in struct:** Requires synchronization; could use channel.
3. **Error channel overuse:** Many error paths send to channel instead of returning directly.
4. **No backpressure:** Unbounded sends to msgChan can block if consumer slow.
