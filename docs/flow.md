# Flow Analysis: agent-sdk-go

Comprehensive flow analysis of the agent-sdk-go codebase, a CLI subprocess wrapper around the Claude CLI.

**Last Updated:** 2026-01-16
**Analysis Scope:** All Go source files in cmd/ and internal/

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Entry Points](#entry-points)
3. [Data Flow](#data-flow)
4. [Control Flow](#control-flow)
5. [Key Interfaces](#key-interfaces)
6. [Error Propagation](#error-propagation)
7. [Concurrency Patterns](#concurrency-patterns)
8. [State Management](#state-management)
9. [Architecture Patterns](#architecture-patterns)

---

## System Overview

**Architecture:** Subprocess wrapper CLI
**Primary Responsibility:** Spawn Claude CLI as subprocess, parse JSON output, expose Go API

```
┌─────────────────────────────────────────────────────────────────┐
│                        User Application                          │
└──────────────────┬──────────────────────────────────────────────┘
                   │
                   │ Go API Call
                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                     agent-sdk-go (This)                          │
│  ┌──────────────┐   ┌──────────────┐   ┌──────────────┐        │
│  │   V2 API     │   │    Client    │   │   Parser     │        │
│  │  (session/   │──▶│  (ClientImpl)│──▶│  (JSON)      │        │
│  │   prompt)    │   └──────┬───────┘   └──────────────┘        │
│  └──────────────┘          │                                     │
│                            ▼                                     │
│                   ┌──────────────────┐                          │
│                   │    Transport     │                          │
│                   │   (subprocess)   │                          │
│                   └────────┬─────────┘                          │
└────────────────────────────┼──────────────────────────────────┘
                             │ exec.Command + pipes
                             ▼
                   ┌──────────────────┐
                   │   Claude CLI     │ (External subprocess)
                   │   (Anthropic)    │
                   └──────────────────┘
```

---

## Entry Points

### 1. CLI Entry Point: `cmd/agent/main.go`

**Purpose:** Demo CLI showcasing agent-sdk-go usage

**Entry flow:**
```
main()
  ├─ Parse command (run/tool/stream)
  ├─ Create flag.FlagSet for subcommand
  ├─ Check CLI availability via cli.IsCLIAvailable()
  ├─ Create client via claude.NewClient(options...)
  ├─ Connect to Claude CLI
  ├─ Execute subcommand logic
  └─ Defer disconnect
```

**Subcommands:**
| Command | Purpose | Key Flow |
|---------|---------|----------|
| `run` | One-shot query | `client.Query(ctx, message)` → collect response → print |
| `tool` | Tool use demo | `client.QueryStream(ctx, message)` → detect tool uses → print |
| `stream` | Streaming demo | `client.QueryStream(ctx, message)` → stream deltas → print |

**Key characteristics:**
- Simple flag parsing (no cobra/viper)
- Direct client usage (no V2 session wrapper)
- Synchronous execution per subcommand

### 2. V2 API Entry Points: `internal/claude/v2/`

#### 2.1 `v2.Prompt()` - One-shot query

**Signature:** `Prompt(ctx, prompt, opts...) (*V2Result, error)`

**Flow:**
```
v2.Prompt()
  ├─ Apply options → DefaultPromptOptions()
  ├─ Validate options
  ├─ Check CLI availability
  ├─ Create client via factory.NewClient()
  ├─ Connect to Claude CLI
  ├─ Call client.QueryStream(ctx, prompt)
  ├─ Collect messages from channel
  ├─ Extract text via ExtractText()
  └─ Return V2Result{Result, SessionID, Duration}
```

**Return path:** Synchronous - blocks until complete response received

#### 2.2 `v2.CreateSession()` - Multi-turn session

**Signature:** `CreateSession(ctx, opts...) (V2Session, error)`

**Flow:**
```
v2.CreateSession()
  ├─ Apply session options
  ├─ Validate options
  ├─ Check CLI availability
  ├─ Create client via factory.NewClient()
  ├─ Generate session ID
  ├─ Connect to Claude CLI
  └─ Return v2SessionImpl{client, sessionID, options}
```

**Session usage pattern:**
```
session.Send(ctx, message)  // Queue message
for msg := range session.Receive(ctx) {  // Stream responses
    // Process messages
}
session.Close()  // Cleanup
```

#### 2.3 `v2.ResumeSession()` - Resume existing session

**Signature:** `ResumeSession(ctx, sessionID, opts...) (V2Session, error)`

**Flow:** Same as `CreateSession()` but sets `sessionID` on client for context continuity

### 3. Direct Client API: `internal/claude/client.go`

**Primary interface:** `claude.Client`

**Creation flow:**
```
claude.NewClient(opts...)
  ├─ DefaultClientOptions()
  ├─ Apply functional options
  ├─ Validate options
  └─ Return &ClientImpl{options}
```

**Usage patterns:**

| Pattern | Method | Behavior |
|---------|--------|----------|
| **Connect/Disconnect** | `Connect(ctx)` / `Disconnect()` | Lifecycle management |
| **One-shot query** | `Query(ctx, prompt)` | Synchronous response |
| **Streaming query** | `QueryStream(ctx, prompt)` | Channel-based streaming |
| **Session query** | `QueryWithSession(ctx, sessionID, prompt)` | Session-aware query |
| **Receive** | `ReceiveMessages(ctx)` / `ReceiveResponse(ctx)` | Low-level message reception |
| **Control** | `Interrupt()` / `SetModel()` / `SetPermissionMode()` | Runtime control |

---

## Data Flow

### 1. Configuration Loading Flow

```
app.Load()
  ├─ config := &Config{}
  ├─ Try load from files (precedence order):
  │   ├─ ./config.yaml
  │   ├─ ~/.config/agent-sdk-go/config.yaml
  │   └─ /etc/agent-sdk-go/config.yaml
  ├─ Override with environment variables:
  │   └─ ANTHROPIC_API_KEY
  └─ Return config
```

**Note:** V2 SDK bypasses app.Config - uses direct client creation

### 2. Message Flow: Query → Response

#### Phase 1: Query Submission

```
User Code
  ├─ client.QueryStream(ctx, prompt)
  │
  └─▶ ClientImpl.QueryStream()
       ├─ Create TransportConfig
       ├─ NewTransportWithPrompt(config, prompt)  // One-shot mode
       ├─ transport.Connect(ctx)
       │   ├─ Discover CLI path via cli.DiscoverCLI()
       │   ├─ Build CLI args: ["-p", "--output-format", "stream-json", "--verbose", prompt]
       │   ├─ exec.CommandContext(ctx, cliPath, args...)
       │   ├─ Create stdout/stderr pipes (NO stdin for one-shot)
       │   ├─ cmd.Start()
       │   ├─ Launch goroutines:
       │   │   ├─ handleStdout() → msgChan
       │   │   └─ handleStderr() → errChan
       │   └─ Return
       │
       └─ Return msgChan, errChan
```

#### Phase 2: Message Parsing (Concurrent)

**Goroutine:** `transport.handleStdout()`

```
handleStdout()
  ├─ bufio.Scanner(stdout)
  ├─ For each line:
  │   ├─ json.Unmarshal → map[string]any
  │   ├─ Extract "type" field
  │   ├─ registry.Parse(messageType, line, lineNumber)
  │   │   ├─ Lookup parser by type
  │   │   ├─ parseAssistantMessage() / parseStreamEvent() / parseResultMessage()
  │   │   └─ Return typed message
  │   └─ Send to msgChan
  │
  └─ On EOF/error:
      ├─ Close msgChan
      ├─ Close errChan
      └─ Set connected = false
```

**Parser Registry (OCP pattern):**

```
MessageParserRegistry
  ├─ parsers: map[string]MessageParserFunc
  ├─ Registered types:
  │   ├─ "user"             → parseUserMessage
  │   ├─ "assistant"        → parseAssistantMessage
  │   ├─ "system"           → parseSystemMessage
  │   ├─ "result"           → parseResultMessage
  │   ├─ "stream_event"     → parseStreamEvent
  │   ├─ "control_request"  → parseControlRequest
  │   └─ "control_response" → parseControlResponse
  │
  └─ Parse(type, json) → delegates to registered func
```

#### Phase 3: Message Consumption

```
User Code
  ├─ msgChan, errChan := client.QueryStream(ctx, prompt)
  │
  └─▶ for loop:
       ├─ select:
       │   ├─ case msg := <-msgChan:
       │   │   ├─ Type assertion:
       │   │   │   ├─ *AssistantMessage → Extract content blocks
       │   │   │   ├─ *StreamEvent → Extract delta
       │   │   │   └─ *ResultMessage → Extract result
       │   │   └─ Process message
       │   │
       │   ├─ case err := <-errChan:
       │   │   └─ Handle error
       │   │
       │   └─ case <-ctx.Done():
       │       └─ Cancel
       │
       └─ Exit when channels closed
```

### 3. Message Types Data Flow

```
Claude CLI stdout
  │
  ├─ {"type":"assistant","content":[...]}
  │   └─▶ AssistantMessage{Content: []ContentBlock}
  │        └─▶ TextBlock{Text: "..."}
  │        └─▶ ThinkingBlock{Thinking: "...", Signature: "..."}
  │        └─▶ ToolUseBlock{Name: "...", Input: {...}}
  │
  ├─ {"type":"stream_event","event":{...}}
  │   └─▶ StreamEvent{Event: map[string]any}
  │        └─▶ ExtractDelta() → text delta string
  │
  ├─ {"type":"result","result":"..."}
  │   └─▶ ResultMessage{Result: *string}
  │
  └─ {"type":"system","subtype":"..."}
      └─▶ SystemMessage{Subtype: string, Data: map[string]any}
```

### 4. V2 Wrapper Data Flow

```
v2SessionImpl.Receive(ctx)
  ├─ Check pendingSend
  ├─ If pending:
  │   ├─ client.QueryStream(ctx, pendingSend.message)
  │   └─ Mark consumed
  ├─ Else:
  │   └─ client.ReceiveMessages(ctx)
  │
  └─▶ wrapMessageChannel(msgChan, errChan)
       ├─ For each message:
       │   ├─ convertToV2Message(msg, sessionID)
       │   │   ├─ AssistantMessage → V2AssistantMessage
       │   │   ├─ ResultMessage → V2ResultMessage
       │   │   ├─ StreamEvent → V2StreamDelta
       │   │   └─ error → V2Error
       │   └─ Send to out channel
       │
       └─ Return out channel
```

---

## Control Flow

### 1. Client Lifecycle

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Lifecycle                          │
└─────────────────────────────────────────────────────────────┘

NewClient(opts...)
  │
  ├─ State: CREATED (transport = nil)
  │
  ▼
Connect(ctx)
  │
  ├─ Create Transport
  ├─ transport.Connect(ctx)
  │   ├─ Spawn subprocess
  │   ├─ Launch goroutines
  │   └─ Set connected = true
  │
  ├─ State: CONNECTED
  │
  ▼
Query operations
  │
  ├─ QueryStream() / Query() / ReceiveMessages()
  │
  ├─ State: ACTIVE
  │
  ▼
Disconnect()
  │
  ├─ transport.Close()
  │   ├─ Cancel context
  │   ├─ Close stdin
  │   ├─ Wait for goroutines (5s timeout)
  │   ├─ Kill subprocess
  │   └─ Close channels
  │
  ├─ transport = nil
  │
  └─ State: DISCONNECTED
```

### 2. Transport Modes

#### Interactive Mode (Multi-turn)

```
NewTransport(config)  // promptArg = nil
  │
  ├─ Args: ["--output-format", "stream-json", "--input-format", "stream-json", "--model", model]
  │
  ├─ Create stdin pipe ✓
  │
  ├─ Connect() → spawn subprocess
  │
  └─ Usage:
      ├─ SendMessage(ctx, message)  // Write to stdin
      └─ ReceiveMessages(ctx)       // Read from stdout
```

#### One-shot Mode (Single query)

```
NewTransportWithPrompt(config, prompt)  // promptArg = &prompt
  │
  ├─ Validate prompt (no shell escapes)
  │
  ├─ Args: ["-p", "--output-format", "stream-json", "--verbose", prompt]
  │
  ├─ NO stdin pipe (causes hanging)
  │
  ├─ Connect() → spawn subprocess with prompt as arg
  │
  └─ Usage:
      └─ ReceiveMessages(ctx)  // Read response from stdout
          └─ Subprocess exits when done
```

**Critical difference:** One-shot mode does NOT create stdin pipe (line 187-193 in transport.go) to prevent hanging

### 3. Goroutine Lifecycle

```
transport.Connect(ctx)
  │
  ├─ ctx, cancel := context.WithCancel(ctx)
  │
  ├─ wg.Add(1)
  │   └─▶ go handleStdout()
  │        ├─ defer wg.Done()
  │        ├─ defer close(msgChan)
  │        ├─ defer close(errChan)
  │        │
  │        └─ for scanner.Scan():
  │             ├─ Check ctx.Done()
  │             ├─ Parse line → message
  │             └─ Send to msgChan
  │
  └─ wg.Add(1)
      └─▶ go handleStderr()
           ├─ defer wg.Done()
           │
           └─ for scanner.Scan():
                ├─ Check ctx.Done()
                └─ Send errors to errChan
```

**Cleanup flow:**
```
transport.Close()
  │
  ├─ cancel()  // Signal goroutines to stop
  │
  ├─ Close stdin
  │
  ├─ wg.Wait() with 5s timeout
  │   ├─ If timeout:
  │   │   ├─ cancel() again
  │   │   └─ Sleep 1s
  │   └─ Else: goroutines finished cleanly
  │
  ├─ Close stdout/stderr
  │
  └─ Kill subprocess
```

### 4. Session State Machine (V2)

```
┌─────────────────────────────────────────────────────────────┐
│                  V2 Session State                            │
└─────────────────────────────────────────────────────────────┘

CreateSession(ctx, opts...)
  │
  ├─ State: {closed: false, pendingSend: nil}
  │
  ▼
Send(ctx, message)
  │
  ├─ pendingSend = &pendingSendData{message, time.Now(), false}
  │
  ├─ State: {closed: false, pendingSend: SET}
  │
  ▼
Receive(ctx)
  │
  ├─ Check pendingSend
  │   ├─ If set && !consumed:
  │   │   ├─ client.QueryStream(ctx, pendingSend.message)
  │   │   └─ pendingSend.consumed = true
  │   └─ Else:
  │       └─ client.ReceiveMessages(ctx)
  │
  ├─ State: {closed: false, pendingSend: CONSUMED}
  │
  ▼
Close()
  │
  ├─ client.Disconnect()
  │
  └─ State: {closed: true}
```

---

## Key Interfaces

### 1. Core Interfaces

#### Client Interface (Composed)

```go
type Client interface {
    Connector        // Connect, Disconnect
    Querier          // Query, QueryWithSession, QueryStream
    Receiver         // ReceiveMessages, ReceiveResponse
    Controller       // Interrupt, SetModel, SetPermissionMode
    ContextManager   // RewindFiles, GetOptions
}
```

**Implementation:** `ClientImpl` (internal/claude/client.go)

**Key characteristics:**
- Interface segregation (ISP)
- Composed from focused sub-interfaces
- Thread-safe (sync.RWMutex)

#### Message Interface

```go
type Message interface {
    Type() string  // Returns message type discriminator
}
```

**Implementations:**
- `UserMessage`
- `AssistantMessage`
- `SystemMessage`
- `ResultMessage`
- `StreamEvent`
- `RawControlMessage`

**Pattern:** Discriminated union via Type() method

#### ContentBlock Interface

```go
type ContentBlock interface {
    BlockType() string
}
```

**Implementations:**
- `TextBlock`
- `ThinkingBlock`
- `ToolUseBlock`
- `ToolResultBlock`

### 2. Factory Pattern

#### Generic Factory Interface

```go
type Factory[T any, O any] interface {
    New(opts ...func(*O)) (T, error)
}

type FactoryFunc[T any, O any] func(opts ...func(*O)) (T, error)
```

**Usage in V2:**
```go
type ClientFactory interface {
    NewClient(opts ...claude.ClientOption) (claude.Client, error)
}

// Default implementation
DefaultClientFactory() ClientFactory
```

**Purpose:** Dependency injection for testability (DIP)

### 3. Parser Registry (OCP Pattern)

```go
type MessageParserRegistry struct {
    mu      sync.RWMutex
    parsers map[string]MessageParserFunc
}

type MessageParserFunc func(jsonStr string, lineNumber int) (shared.Message, error)
```

**Methods:**
- `Register(messageType, parser)` - Add custom parser
- `Parse(type, json, line)` - Delegate to registered parser
- `HasParser(type)` - Check if parser exists

**Pattern:** Open/Closed Principle - extend without modification

### 4. V2 Session Interface

```go
type V2Session interface {
    Send(ctx context.Context, message string) error
    Receive(ctx context.Context) <-chan V2Message
    ReceiveIterator(ctx context.Context) V2MessageIterator
    Close() error
    SessionID() string
}
```

**Implementation:** `v2SessionImpl`

**Key methods:**
- `Send()` - Queue message for next Receive()
- `Receive()` - Returns channel of V2Message
- `ReceiveIterator()` - Alternative iterator-based API

---

## Error Propagation

### 1. Error Type Hierarchy

```
error (base interface)
  │
  ├─ CLINotFoundError
  │   ├─ Path: string
  │   └─ Command: string
  │
  ├─ ConnectionError
  │   ├─ Reason: string
  │   └─ Inner: error (wrapped)
  │
  ├─ TimeoutError
  │   ├─ Operation: string
  │   └─ Timeout: string
  │
  ├─ ParserError
  │   ├─ Line: int
  │   ├─ Offset: int
  │   ├─ Data: string
  │   └─ Reason: string
  │
  ├─ ProtocolError
  │   ├─ MessageType: string
  │   └─ Reason: string
  │
  └─ ProcessError
      ├─ PID: int
      ├─ Command: string
      ├─ Reason: string
      └─ Signal: string
```

### 2. Error Propagation Flow

```
Layer 1: Transport (subprocess)
  │
  ├─ CLI not found
  │   └─▶ NewCLINotFoundError(path, cmd)
  │
  ├─ Process spawn failure
  │   └─▶ fmt.Errorf("start CLI process: %w", err)
  │
  ├─ JSON parse error
  │   └─▶ Send to errChan
  │
  └─ Subprocess crash
      └─▶ NewProcessError(pid, cmd, reason, signal)

  ▼
Layer 2: Client (client.go)
  │
  ├─ Wrap transport errors
  │   └─▶ fmt.Errorf("connect transport: %w", err)
  │
  └─ Query timeout
      └─▶ ctx.Err() (context deadline exceeded)

  ▼
Layer 3: V2 API (v2/)
  │
  ├─ Wrap client errors
  │   └─▶ fmt.Errorf("create client: %w", err)
  │
  └─ Convert to V2Error messages
      └─▶ V2Error{Type: "error", Error: err.Error()}

  ▼
Layer 4: User Code
  │
  └─ Handle errors via:
      ├─ errors.Is(err, target)
      ├─ errors.As(err, &typed)
      └─ Type check functions (IsCLINotFound, IsConnectionError)
```

### 3. Error Handling Patterns

#### Pattern 1: Channel-based error propagation

```go
msgChan, errChan := client.QueryStream(ctx, prompt)

for {
    select {
    case msg, ok := <-msgChan:
        if !ok {
            return nil  // Normal completion
        }
        // Process message
    case err, ok := <-errChan:
        if !ok {
            return nil  // No error
        }
        return err  // Propagate error
    }
}
```

#### Pattern 2: Context wrapping

```go
if err := transport.Connect(ctx); err != nil {
    return fmt.Errorf("connect transport: %w", err)
}
```

#### Pattern 3: Sentinel errors

```go
var ErrNoMoreMessages = errors.New("no more messages")

func (it *v2MessageIteratorImpl) Next(ctx) (V2Message, error) {
    select {
    case msg, ok := <-it.ch:
        if !ok {
            return nil, ErrNoMoreMessages
        }
        return msg, nil
    }
}
```

### 4. Error Recovery

```
transport.Close()
  │
  ├─ Attempt graceful shutdown:
  │   ├─ Cancel context
  │   ├─ Close stdin
  │   └─ Wait for goroutines (5s timeout)
  │
  ├─ On timeout:
  │   ├─ Force cancel
  │   ├─ Sleep 1s (allow cleanup)
  │   └─ Proceed to kill
  │
  └─ Force cleanup:
      ├─ Close stdout/stderr
      └─ subprocess.Kill()
```

**Guarantees:**
- Resources always released (defer)
- No goroutine leaks (WaitGroup + timeout)
- Channels closed exactly once (defer in handleStdout)

---

## Concurrency Patterns

### 1. Goroutine Management

#### Pattern: Bounded goroutine lifecycle

```go
type Transport struct {
    wg     sync.WaitGroup
    ctx    context.Context
    cancel context.CancelFunc
}

func (t *Transport) Connect(ctx) error {
    t.ctx, t.cancel = context.WithCancel(ctx)

    t.wg.Add(1)
    go t.handleStdout()

    t.wg.Add(1)
    go t.handleStderr()
}

func (t *Transport) Close() error {
    t.cancel()  // Signal goroutines

    done := make(chan struct{})
    go func() {
        t.wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        // Clean shutdown
    case <-time.After(5 * time.Second):
        // Timeout - force cleanup
    }
}
```

**Key properties:**
- WaitGroup tracks all spawned goroutines
- Context cancellation signals shutdown
- Timeout prevents indefinite blocking
- Cleanup always executes (defer)

### 2. Channel Patterns

#### Pattern 1: Buffered channels for async communication

```go
const channelBufferSize = 100

msgChan := make(chan shared.Message, channelBufferSize)
errChan := make(chan error, channelBufferSize)
```

**Rationale:**
- Prevents goroutine blocking on send
- Tolerates burst traffic
- Size 100 balances memory vs throughput

#### Pattern 2: Channel ownership

```go
func (t *Transport) handleStdout() {
    defer t.wg.Done()
    defer func() {
        close(t.msgChan)  // Owner closes
        close(t.errChan)
    }()

    // ... read and send messages
}
```

**Rule:** Creator of channel closes it (prevents double-close panic)

#### Pattern 3: Non-blocking send with context

```go
select {
case t.msgChan <- msg:
    // Sent successfully
case <-t.ctx.Done():
    return  // Cancelled
}
```

**Rationale:** Prevents goroutine from blocking on closed channels

### 3. Thread-Safe State Management

#### Pattern: RWMutex for read-heavy operations

```go
type ClientImpl struct {
    transport *subprocess.Transport
    options   *ClientOptions
    sessionID string
    mu        sync.RWMutex
}

func (c *ClientImpl) Query(ctx, prompt) (string, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    // Read-only access to transport
}

func (c *ClientImpl) Connect(ctx) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    // Mutate transport
}
```

**Access patterns:**
- `RLock()` for reads: `Query`, `QueryStream`, `ReceiveMessages`, `GetOptions`, `GetSessionID`
- `Lock()` for writes: `Connect`, `Disconnect`, `SetModel`, `SetSessionID`

### 4. Concurrency Safety in Parser

```go
type MessageParserRegistry struct {
    mu      sync.RWMutex
    parsers map[string]MessageParserFunc
}

func (r *MessageParserRegistry) Parse(msgType, json, line) (Message, error) {
    r.mu.RLock()
    parser, ok := r.parsers[msgType]
    r.mu.RUnlock()  // Release before expensive parse

    if !ok {
        return nil, fmt.Errorf("unknown type")
    }

    return parser(json, line)
}
```

**Pattern:** Lock only map access, not expensive operations

### 5. Goroutine Communication Flow

```
Main Goroutine                  handleStdout()              handleStderr()
      │                               │                          │
      ├─ Connect()                    │                          │
      │   ├─ Start subprocess         │                          │
      │   ├─ spawn ──────────────────▶│                          │
      │   └─ spawn ──────────────────────────────────────────────▶│
      │                               │                          │
      ├─ ReceiveMessages()            │                          │
      │   └─ returns (msgChan, errCh) │                          │
      │                               │                          │
      ├─ select on channels           │                          │
      │   ◀────── msgChan ─────────────┤                          │
      │   ◀────── errChan ─────────────┼─────────────────────────▶│
      │                               │                          │
      ├─ Close()                      │                          │
      │   ├─ cancel() ────────────────▶│ (ctx.Done())            │
      │   │                           │                          │
      │   │                           │                          │
      │   ├─ wg.Wait() ◀──────────────┤ wg.Done()                │
      │   │            ◀──────────────────────────────────────────┤
      │   └─ cleanup                  │ (exited)                 │
      │                               │                          │
```

### 6. Race Condition Prevention

#### Issue: Concurrent access to transport

```go
func (c *ClientImpl) QueryStream(ctx, prompt) (<-chan Message, <-chan error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    // Create NEW transport for one-shot query
    // Does NOT use c.transport (which may be nil or in use)
    transport, err := subprocess.NewTransportWithPrompt(config, prompt)
}
```

**Pattern:** One-shot queries create isolated transports (no shared state)

#### Issue: pendingSend in session

```go
func (s *v2SessionImpl) Receive(ctx) <-chan V2Message {
    s.mu.Lock()

    // Take COPY of pendingSend while holding lock
    pendingDataCopy := s.pendingSend

    s.mu.Unlock()  // Release before I/O

    // Work with copy outside lock
    if pendingDataCopy != nil && !pendingDataCopy.consumed {
        // ...
    }
}
```

**Pattern:** Copy under lock, process outside lock

---

## State Management

### 1. Client State

```go
type ClientImpl struct {
    transport *subprocess.Transport  // nil = disconnected, non-nil = connected
    options   *ClientOptions         // immutable after creation
    sessionID string                 // mutable via SetSessionID
    mu        sync.RWMutex            // protects all fields
}
```

**State transitions:**
```
CREATED:      transport = nil
CONNECTED:    transport != nil, transport.IsConnected() = true
DISCONNECTED: transport = nil (after Close)
```

**Invariants:**
- `Connect()` can only be called when `transport == nil`
- `Disconnect()` is idempotent (safe to call multiple times)

### 2. Transport State

```go
type Transport struct {
    connected bool               // atomic state flag
    cmd       *exec.Cmd          // subprocess handle
    stdin     io.WriteCloser     // nil in one-shot mode
    stdout    io.ReadCloser      // always present
    stderr    io.ReadCloser      // always present
    msgChan   chan Message       // created in Connect()
    errChan   chan error         // created in Connect()
    ctx       context.Context    // created in Connect()
    cancel    context.CancelFunc // created in Connect()
    wg        sync.WaitGroup     // tracks goroutines
    mu        sync.RWMutex       // protects state
}
```

**State transitions:**
```
CREATED:    connected = false, cmd = nil, channels = nil
CONNECTED:  connected = true, cmd running, channels open, goroutines running
CLOSING:    connected = false, cancel() called, goroutines exiting
CLOSED:     connected = false, cmd = nil, channels closed
```

### 3. Session State (V2)

```go
type v2SessionImpl struct {
    client      claude.Client
    sessionID   string
    closed      bool
    pendingSend *pendingSendData  // nil or set
    resumed     bool              // true if ResumeSession
    mu          sync.RWMutex
}

type pendingSendData struct {
    message  string
    sentAt   time.Time
    consumed bool  // true after first Receive()
}
```

**State transitions:**
```
CREATED:         closed = false, pendingSend = nil
MESSAGE_QUEUED:  closed = false, pendingSend != nil, consumed = false
MESSAGE_SENT:    closed = false, pendingSend != nil, consumed = true
CLOSED:          closed = true
```

### 4. Parser State

```go
type Parser struct {
    buffer     string  // incomplete JSON
    lineNumber int     // current line number
    registry   *MessageParserRegistry
}
```

**State transitions:**
```
ParseMessage(raw):
  buffer += raw
  ├─ Complete JSON found:
  │   ├─ Parse and return message
  │   ├─ Update buffer (remove parsed)
  │   └─ Increment lineNumber
  └─ Incomplete JSON:
      ├─ Buffer remains
      └─ Return (nil, nil)
```

---

## Architecture Patterns

### 1. Dependency Injection (DIP)

**Example: Client Factory in V2**

```go
// Abstraction
type ClientFactory interface {
    NewClient(opts ...claude.ClientOption) (claude.Client, error)
}

// Default implementation
type defaultClientFactory struct{}

func (f *defaultClientFactory) NewClient(opts ...) (claude.Client, error) {
    return claude.NewClient(opts...)
}

// Injection point
type V2SessionOptions struct {
    clientFactory ClientFactory  // nil = use default
}

// Usage
func CreateSession(ctx, opts...) (V2Session, error) {
    factory := options.clientFactory
    if factory == nil {
        factory = DefaultClientFactory()
    }

    client, err := factory.NewClient(...)  // Use abstraction
}
```

**Benefits:**
- Testable (inject mock factory)
- Flexible (swap implementations)
- Follows SOLID principles

### 2. Interface Segregation (ISP)

**Example: Client interface composition**

```go
type Client interface {
    Connector       // 2 methods
    Querier         // 3 methods
    Receiver        // 2 methods
    Controller      // 3 methods
    ContextManager  // 2 methods
}
```

**Benefits:**
- Clients depend only on methods they use
- Easier to mock (partial interfaces)
- Clear responsibility boundaries

### 3. Open/Closed Principle (OCP)

**Example: Parser Registry**

```go
// Open for extension
registry := parser.DefaultRegistry()
registry.Register("custom_type", func(json, line) (Message, error) {
    // Custom parser
})

// Closed for modification
// No need to modify parser.go or registry.go
```

**Benefits:**
- Add custom message types without modifying SDK
- Third-party extensions possible
- Backward compatible

### 4. Single Responsibility Principle (SRP)

**Example: Option structs**

```go
type ClientOptions struct {
    ConnectionOptions  // CLI path, timeout, env
    BufferOptions      // buffer size, max messages
    ModelOptions       // model name, context files
    DebugOptions       // trace, cache, logger
}
```

**Benefits:**
- Each struct has one reason to change
- Composition over inheritance
- Reusable option groups

### 5. Factory Pattern

**Example: Transport creation**

```go
func createTransport(config, promptArg) (*Transport, error) {
    // Common initialization
}

func NewTransport(config) (*Transport, error) {
    return createTransport(config, nil)  // Interactive mode
}

func NewTransportWithPrompt(config, prompt) (*Transport, error) {
    return createTransport(config, &prompt)  // One-shot mode
}
```

**Benefits:**
- Encapsulate complex creation logic
- Ensure consistent initialization
- Support multiple creation strategies

### 6. Strategy Pattern

**Example: Message parsing**

```go
type MessageParserFunc func(json, line) (Message, error)

// Different strategies for different message types
registry.Register("assistant", parseAssistantMessage)
registry.Register("stream_event", parseStreamEvent)
registry.Register("result", parseResultMessage)
```

**Benefits:**
- Runtime algorithm selection
- Easy to add new strategies
- Decouples algorithm from context

### 7. Adapter Pattern

**Example: V2 Message Wrapping**

```go
func convertToV2Message(msg claude.Message, sessionID) V2Message {
    switch m := msg.(type) {
    case *shared.AssistantMessage:
        return &V2AssistantMessage{...}
    case *shared.ResultMessage:
        return &V2ResultMessage{...}
    }
}
```

**Benefits:**
- Adapt internal types to public API
- Isolate implementation details
- Version compatibility

### 8. Resource Management Pattern

**Example: Cleanup with defer**

```go
func (c *ClientImpl) Query(ctx, prompt) (string, error) {
    // ... create transport

    if err := transport.Connect(ctx); err != nil {
        return "", err
    }
    defer transport.Close()  // Always cleanup

    // ... use transport
}
```

**Benefits:**
- Guaranteed cleanup
- Exception-safe
- Clear resource ownership

---

## Appendix: Key Files Reference

| File Path | Responsibility | Key Types |
|-----------|----------------|-----------|
| `cmd/agent/main.go` | CLI demo entry point | main(), runCmd(), toolCmd(), streamCmd() |
| `internal/app/config.go` | Configuration loading | Config, Load() |
| `internal/claude/client.go` | Client implementation | ClientImpl, Query(), QueryStream() |
| `internal/claude/types.go` | Core interfaces | Client, Message, Transport |
| `internal/claude/options.go` | Client options | ClientOptions, WithModel(), etc. |
| `internal/claude/subprocess/transport.go` | Subprocess management | Transport, Connect(), handleStdout() |
| `internal/claude/parser/json.go` | JSON parsing | Parser, ParseMessage() |
| `internal/claude/parser/registry.go` | Parser registry (OCP) | MessageParserRegistry, Register() |
| `internal/claude/cli/discovery.go` | CLI discovery | DiscoverCLI(), IsCLIAvailable() |
| `internal/claude/v2/session.go` | V2 session API | CreateSession(), v2SessionImpl |
| `internal/claude/v2/prompt.go` | V2 one-shot API | Prompt() |
| `internal/claude/shared/factory.go` | Generic factory | Factory[T,O], DefaultFactoryHolder[T,O] |

---

## Summary

**Architecture:** Clean layered architecture with clear separation of concerns

**Key strengths:**
1. **SOLID compliance** - Interface segregation, dependency injection, SRP
2. **Concurrency safety** - RWMutex, bounded goroutines, context cancellation
3. **Extensibility** - Parser registry (OCP), factory injection (DIP)
4. **Resource safety** - Guaranteed cleanup, no leaks
5. **Error handling** - Typed errors, context wrapping, sentinel errors

**Critical flows:**
1. **One-shot query:** NewClient → Connect → QueryStream → Parse stdout → Channels → Disconnect
2. **Session:** CreateSession → Send (queue) → Receive (execute) → Close
3. **Subprocess:** exec.Command → pipes → goroutines → channels → cleanup

**Concurrency model:**
- Bounded goroutine lifecycle (WaitGroup + timeout)
- Channel-based async communication (buffered, owner-closes)
- Thread-safe state (RWMutex, copy-on-read)
- Context-driven cancellation

**Data flow:**
- CLI stdout → Scanner → JSON parse → Registry → Typed Message → Channel → User code
- Errors propagated via dedicated error channel (parallel to message channel)
