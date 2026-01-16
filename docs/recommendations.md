# Prioritized Improvement Recommendations

**Project:** agent-sdk-go
**Date:** 2026-01-16
**For:** Development team / Maintainers

---

## Executive Summary

This document provides **actionable recommendations** to improve the agent-sdk-go codebase, prioritized by impact and effort.

**Strategic Goals:**
1. **Make it usable as a library** (currently blocked)
2. **Align documentation with reality** (fix misleading claims)
3. **Improve test coverage** (reduce CI dependency on Claude CLI)
4. **Clarify API surface** (V1 vs V2, deprecations)
5. **Harden for production** (error handling, resource limits)

---

## Recommendation Framework

| Priority | Impact | Effort | Timeline |
|----------|--------|--------|----------|
| **P0** | Critical, blocking users | Low-Medium | Week 1-2 |
| **P1** | High value, improves core | Medium | Week 3-6 |
| **P2** | Quality of life, nice-to-have | Medium-High | Week 7-12 |
| **P3** | Future enhancements | High | Backlog |

---

## P0 Recommendations (Critical - Week 1-2)

### 1. Create Public SDK Package

**Problem:** All code in `internal/`, cannot be imported by users.

**Solution:**

**Step 1:** Create SDK directory structure
```bash
mkdir -p sdk
```

**Step 2:** Move public API to `sdk/`
```
sdk/
├── client.go          ← From internal/claude/client.go
├── types.go           ← Public types only
├── options.go         ← WithModel, WithTimeout, etc.
├── session.go         ← From internal/claude/v2/session.go
├── errors.go          ← Public error types
└── doc.go             ← Package documentation
```

**Step 3:** Keep implementation in `internal/`
```
internal/
├── subprocess/        ← Transport stays internal
├── parser/            ← Parsing stays internal
└── shared/            ← Shared utilities
```

**Step 4:** Update imports
```go
// Before:
import "agent-sdk-go/internal/claude"

// After:
import "agent-sdk-go/sdk"
```

**Effort:** 1 week
**Impact:** ✅ Unblocks library usage
**Risk:** Low (mostly moving files)

---

### 2. Fix Misleading Documentation

**Problem:** Documentation claims features that don't exist or misrepresents architecture.

**Solution:**

**Fix 1:** Update README.md
```markdown
# Before:
A Go implementation of the Anthropic Claude Agent SDK,
ported from TypeScript.

# After:
A Go SDK for interacting with Claude via the Claude CLI.

**Architecture:** This SDK wraps the Claude CLI as a subprocess
(not direct HTTP API calls like the TypeScript SDK).
```

**Fix 2:** Remove tool system example from CLAUDE.md
```markdown
# Remove this section (tool registration doesn't work):
### Tool Development Pattern

# Replace with:
### Tool Observation Pattern
The SDK can observe tool uses in Claude's responses:
```go
if shared.HasToolUses(msg) {
    toolUses := shared.ExtractToolUses(msg)
    // Process tool use blocks
}
```
```

**Fix 3:** Fix import path in README
```bash
# Before:
go get agent-sdk-go/sdk

# After (until public package exists):
# This SDK is currently in development.
# All APIs are in internal/ and subject to change.
```

**Effort:** 2 days
**Impact:** ✅ Sets correct expectations
**Risk:** None

---

### 3. Remove Deprecated Code

**Problem:** Dead code confuses users and increases maintenance burden.

**Solution:**

**Step 1:** Delete deprecated functions
```go
// DELETE from internal/app/config.go:
func NewClientFromConfig(cfg *Config) (any, error)
func NewClient() (any, error)
```

**Step 2:** Remove deprecated entry point
```bash
# Move cmd/main.go to examples/simple-demo/main.go
mkdir -p examples/simple-demo
mv cmd/main.go examples/simple-demo/main.go
```

**Step 3:** Update references
```bash
# Update build instructions in README
# Remove deprecated usage examples
```

**Effort:** 1 day
**Impact:** ✅ Cleaner API surface
**Risk:** None (code already marked deprecated)

---

### 4. Add Prominent CLI Dependency Notice

**Problem:** Users expect HTTP SDK, get subprocess wrapper.

**Solution:**

**Add to README.md (top):**
```markdown
## ⚠️ Important: Claude CLI Required

This SDK requires the [Claude CLI](https://github.com/anthropics/claude-cli)
to be installed and available in your PATH.

**This is NOT a direct HTTP API client.** It spawns the Claude CLI as a
subprocess and communicates via stdin/stdout.

### Installation
```bash
# macOS
brew install claude

# Linux
# See: https://github.com/anthropics/claude-cli/releases

# Verify installation
claude --version
```
```

**Effort:** 1 hour
**Impact:** ✅ Prevents user confusion
**Risk:** None

---

## P1 Recommendations (High Value - Week 3-6)

### 5. Implement Mock Transport for Testing

**Problem:** Core client untested, tests require Claude CLI.

**Solution:**

**Step 1:** Define MockTransport
```go
// internal/claude/mock/transport.go
type MockTransport struct {
    Messages []shared.Message
    Errors   []error
    Calls    []string  // Track method calls
}

func (m *MockTransport) Connect(ctx context.Context) error {
    m.Calls = append(m.Calls, "Connect")
    return nil
}

func (m *MockTransport) ReceiveMessages(ctx context.Context) (<-chan shared.Message, <-chan error) {
    msgChan := make(chan shared.Message, len(m.Messages))
    errChan := make(chan error, len(m.Errors))

    for _, msg := range m.Messages {
        msgChan <- msg
    }
    close(msgChan)

    for _, err := range m.Errors {
        errChan <- err
    }
    close(errChan)

    return msgChan, errChan
}
```

**Step 2:** Add client tests
```go
// internal/claude/client_test.go
func TestClientQuery(t *testing.T) {
    mock := &mock.MockTransport{
        Messages: []shared.Message{
            &shared.AssistantMessage{
                MessageType: "assistant",
                Content: []shared.ContentBlock{
                    &shared.TextBlock{Text: "Hello!"},
                },
            },
        },
    }

    client := &ClientImpl{transport: mock}

    result, err := client.Query(context.Background(), "Hi")
    assert.NoError(t, err)
    assert.Equal(t, "Hello!", result)
    assert.Contains(t, mock.Calls, "Connect")
}
```

**Effort:** 1 week
**Impact:** ✅ Core client testable without CLI
**Risk:** Low

---

### 6. Consolidate V1/V2 APIs

**Problem:** Two API paradigms, no clear guidance on which to use.

**Solution:**

**Option A: Deprecate V1, promote V2**
```go
// sdk/client.go
// Deprecated: Use session-based API (sdk.CreateSession) instead.
// This API will be removed in v2.0.
func NewClient(opts ...Option) (*Client, error) {
    // ...
}
```

**Option B: V1 for simple, V2 for advanced**

Add to documentation:
```markdown
## API Selection Guide

### Use V1 (Client API) when:
- One-shot queries (no conversation history)
- Simple request/response
- Don't need session management

### Use V2 (Session API) when:
- Multi-turn conversations
- Need conversation history
- Complex agent workflows
- Streaming responses
```

**Recommendation:** Choose Option B + add migration guide.

**Effort:** 3 days (docs) + 2 weeks (migration guide + examples)
**Impact:** ✅ Clarifies API usage
**Risk:** Low (additive, no breaking changes)

---

### 7. Add Error Handling Best Practices

**Problem:** Errors may be lost in channels during shutdown.

**Solution:**

**Fix 1:** Drain error channel on shutdown
```go
// subprocess/transport.go:Close()
func (t *Transport) Close() error {
    // ... existing shutdown logic ...

    // Drain error channel to prevent goroutine blocking
    go func() {
        for range t.errChan {
            // Discard remaining errors
        }
    }()

    close(t.msgChan)
    close(t.errChan)

    return nil
}
```

**Fix 2:** Add error callback option
```go
// options.go
type ClientOptions struct {
    // ...
    ErrorHandler func(error)  // Called for non-fatal errors
}

// In transport:
if err != nil {
    if t.options.ErrorHandler != nil {
        t.options.ErrorHandler(err)
    }
    select {
    case t.errChan <- err:
    case <-t.ctx.Done():
    }
}
```

**Effort:** 1 week
**Impact:** ✅ No lost errors, better observability
**Risk:** Low

---

### 8. Add Benchmarks

**Problem:** Buffer sizes chosen arbitrarily, no performance baseline.

**Solution:**

**Step 1:** Add benchmark suite
```go
// internal/claude/subprocess/transport_bench_test.go
func BenchmarkTransportThroughput(b *testing.B) {
    for _, bufSize := range []int{10, 50, 100, 500} {
        b.Run(fmt.Sprintf("BufferSize%d", bufSize), func(b *testing.B) {
            // Benchmark with different buffer sizes
        })
    }
}

func BenchmarkJSONParsing(b *testing.B) {
    raw := `{"type":"assistant","content":[{"type":"text","text":"..."}]}`
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        parseAssistantMessage(raw)
    }
}
```

**Step 2:** Run and optimize
```bash
go test -bench=. -benchmem ./internal/...
# Adjust buffer sizes based on results
```

**Effort:** 1 week
**Impact:** ✅ Data-driven performance tuning
**Risk:** None

---

### 9. Add Session Persistence

**Problem:** No way to save/restore sessions.

**Solution:**

**Step 1:** Define session state
```go
// sdk/session.go
type SessionState struct {
    ID       string                 `json:"id"`
    Model    string                 `json:"model"`
    History  []shared.Message       `json:"history"`
    Created  time.Time              `json:"created"`
    Modified time.Time              `json:"modified"`
}

func (s *v2SessionImpl) Save(path string) error {
    state := SessionState{
        ID:       s.sessionID,
        Model:    s.options.Model,
        History:  s.getHistory(),  // Capture from client
        Created:  s.created,
        Modified: time.Now(),
    }

    data, err := json.MarshalIndent(state, "", "  ")
    if err != nil {
        return fmt.Errorf("marshal state: %w", err)
    }

    return os.WriteFile(path, data, 0600)
}
```

**Step 2:** Load session
```go
func LoadSession(ctx context.Context, path string, opts ...SessionOption) (V2Session, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("read state: %w", err)
    }

    var state SessionState
    if err := json.Unmarshal(data, &state); err != nil {
        return nil, fmt.Errorf("unmarshal state: %w", err)
    }

    // Resume session with history
    session, err := ResumeSession(ctx, state.ID, opts...)
    if err != nil {
        return nil, err
    }

    // Restore history (if CLI supports it)
    // ...

    return session, nil
}
```

**Effort:** 2 weeks
**Impact:** ✅ Enables stateful applications
**Risk:** Medium (depends on CLI support for history)

---

## P2 Recommendations (Quality of Life - Week 7-12)

### 10. Add CI/CD Pipeline

**Problem:** No automated testing, linting, or releases.

**Solution:**

**Step 1:** GitHub Actions workflow
```yaml
# .github/workflows/ci.yml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.txt ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.txt

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: golangci/golangci-lint-action@v3
        with:
          version: latest
```

**Step 2:** golangci-lint config
```yaml
# .golangci.yml
linters:
  enable:
    - gofmt
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
```

**Effort:** 3 days
**Impact:** ✅ Automated quality checks
**Risk:** Low

---

### 11. Add Docker Support

**Problem:** Claude CLI dependency complicates containerization.

**Solution:**

**Step 1:** Dockerfile
```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o agent-sdk-go ./cmd/agent

FROM ubuntu:22.04
RUN apt-get update && apt-get install -y curl ca-certificates

# Install Claude CLI (example - adjust based on actual installation)
RUN curl -L -o /usr/local/bin/claude https://... \
    && chmod +x /usr/local/bin/claude

COPY --from=builder /app/agent-sdk-go /usr/local/bin/
CMD ["agent-sdk-go", "run"]
```

**Step 2:** docker-compose for development
```yaml
# docker-compose.yml
version: '3'
services:
  sdk:
    build: .
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
    volumes:
      - ./config.yaml:/etc/agent-sdk-go/config.yaml
```

**Effort:** 2 days
**Impact:** ✅ Easier deployment
**Risk:** Medium (CLI installation may change)

---

### 12. Add Middleware System

**Problem:** Cannot intercept/modify requests without forking.

**Solution:**

**Step 1:** Define middleware interface
```go
// sdk/middleware.go
type Middleware func(next Handler) Handler

type Handler interface {
    Handle(ctx context.Context, msg Message) (Message, error)
}

type ClientWithMiddleware struct {
    client     Client
    middleware []Middleware
}

func (c *ClientWithMiddleware) Use(m Middleware) {
    c.middleware = append(c.middleware, m)
}
```

**Step 2:** Example middleware
```go
// Logging middleware
func LoggingMiddleware(next Handler) Handler {
    return HandlerFunc(func(ctx context.Context, msg Message) (Message, error) {
        log.Printf("Request: %+v", msg)
        resp, err := next.Handle(ctx, msg)
        log.Printf("Response: %+v, Error: %v", resp, err)
        return resp, err
    })
}

// Usage:
client.Use(LoggingMiddleware)
client.Use(MetricsMiddleware)
client.Query(ctx, "Hello")
```

**Effort:** 1 week
**Impact:** ✅ Extensible without forking
**Risk:** Low

---

### 13. Add Retry Logic

**Problem:** Users must implement retries themselves.

**Solution:**

**Step 1:** Retry options
```go
// options.go
type ClientOptions struct {
    // ...
    RetryAttempts int
    RetryBackoff  BackoffStrategy
}

type BackoffStrategy interface {
    Duration(attempt int) time.Duration
}

type ExponentialBackoff struct {
    Base time.Duration
    Max  time.Duration
}

func (e *ExponentialBackoff) Duration(attempt int) time.Duration {
    d := e.Base * (1 << uint(attempt))
    if d > e.Max {
        return e.Max
    }
    return d
}
```

**Step 2:** Retry wrapper
```go
// client.go
func (c *ClientImpl) QueryWithRetry(ctx context.Context, prompt string) (string, error) {
    var lastErr error

    for attempt := 0; attempt <= c.options.RetryAttempts; attempt++ {
        if attempt > 0 {
            backoff := c.options.RetryBackoff.Duration(attempt - 1)
            time.Sleep(backoff)
        }

        result, err := c.Query(ctx, prompt)
        if err == nil {
            return result, nil
        }

        // Only retry on transient errors
        if !isRetryable(err) {
            return "", err
        }

        lastErr = err
    }

    return "", fmt.Errorf("max retries exceeded: %w", lastErr)
}
```

**Effort:** 1 week
**Impact:** ✅ More resilient
**Risk:** Low

---

### 14. Add Resource Limits

**Problem:** Can spawn unlimited Claude CLI processes.

**Solution:**

**Step 1:** Session pool
```go
// sdk/pool.go
type SessionPool struct {
    max      int
    sessions chan V2Session
    mu       sync.Mutex
}

func NewSessionPool(ctx context.Context, max int, opts ...SessionOption) (*SessionPool, error) {
    pool := &SessionPool{
        max:      max,
        sessions: make(chan V2Session, max),
    }

    // Pre-create sessions
    for i := 0; i < max; i++ {
        session, err := CreateSession(ctx, opts...)
        if err != nil {
            return nil, err
        }
        pool.sessions <- session
    }

    return pool, nil
}

func (p *SessionPool) Acquire(ctx context.Context) (V2Session, error) {
    select {
    case session := <-p.sessions:
        return session, nil
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (p *SessionPool) Release(session V2Session) {
    p.sessions <- session
}
```

**Step 2:** Usage
```go
pool, _ := NewSessionPool(ctx, 10)  // Max 10 concurrent sessions

session, _ := pool.Acquire(ctx)
defer pool.Release(session)

session.Send(ctx, "Hello")
// ...
```

**Effort:** 1 week
**Impact:** ✅ Prevents resource exhaustion
**Risk:** Low

---

### 15. Improve Configuration Validation

**Problem:** Invalid configs accepted silently.

**Solution:**

**Step 1:** Stricter validation
```go
// options.go
func (o *ClientOptions) Validate() error {
    // Validate model (against known models)
    validModels := []string{
        "claude-3-5-sonnet-20241022",
        "claude-3-opus-20240229",
        // ...
    }
    if !slices.Contains(validModels, o.Model) {
        return fmt.Errorf("invalid model: %s (valid: %v)", o.Model, validModels)
    }

    // Validate timeout
    if o.Timeout != "" {
        d, err := time.ParseDuration(o.Timeout)
        if err != nil {
            return fmt.Errorf("invalid timeout: %w", err)
        }
        if d < 0 {
            return fmt.Errorf("timeout cannot be negative: %s", o.Timeout)
        }
    }

    // Validate buffer sizes
    if o.BufferSize < 1 || o.BufferSize > 10000 {
        return fmt.Errorf("buffer size must be 1-10000, got %d", o.BufferSize)
    }

    return nil
}
```

**Effort:** 2 days
**Impact:** ✅ Fail fast on invalid config
**Risk:** Low

---

## P3 Recommendations (Future - Backlog)

### 16. HTTP API Fallback

**Problem:** Requires Claude CLI, cannot work in restricted environments.

**Solution:**

Implement direct HTTP API client (like TypeScript SDK):
```go
// internal/http/client.go
type HTTPClient struct {
    apiKey  string
    baseURL string
    client  *http.Client
}

func (c *HTTPClient) SendMessage(ctx context.Context, req MessageRequest) (*MessageResponse, error) {
    // Direct API call
    url := c.baseURL + "/v1/messages"
    // ...
}
```

**Effort:** 4-6 weeks (full HTTP client implementation)
**Impact:** ✅ Works without CLI, true TypeScript SDK port
**Risk:** High (large scope, requires API key handling)

---

### 17. MCP (Model Context Protocol) Support

**Problem:** Missing feature from enhancement plan.

**Solution:**

Integrate Model Context Protocol for tool execution:
```go
// sdk/mcp.go
type MCPServer interface {
    ListTools() ([]Tool, error)
    ExecuteTool(ctx context.Context, name string, args map[string]any) (any, error)
}

func (c *Client) RegisterMCPServer(server MCPServer) error {
    // Register MCP server
}
```

**Effort:** 8-12 weeks (requires MCP spec implementation)
**Impact:** ✅ Advanced tool support
**Risk:** High (complex protocol)

---

### 18. Structured Outputs with JSON Schema

**Problem:** No schema validation for responses.

**Solution:**

```go
// sdk/schema.go
type OutputSchema struct {
    Type       string                 `json:"type"`
    Properties map[string]Property    `json:"properties"`
    Required   []string               `json:"required"`
}

func (c *Client) QueryWithSchema(ctx context.Context, prompt string, schema OutputSchema) (*StructuredResponse, error) {
    // Validate response against schema
}
```

**Effort:** 3-4 weeks
**Impact:** ✅ Type-safe responses
**Risk:** Medium (depends on CLI support)

---

## Implementation Roadmap

### Sprint 1-2 (P0 - Critical)
- Week 1: Create public SDK package (#1)
- Week 1-2: Fix documentation (#2)
- Week 2: Remove deprecated code (#3)
- Week 2: Add CLI dependency notice (#4)

**Deliverable:** Usable library with accurate docs

---

### Sprint 3-6 (P1 - High Value)
- Week 3-4: Mock transport + client tests (#5)
- Week 4-5: Consolidate APIs + migration guide (#6)
- Week 5: Error handling improvements (#7)
- Week 6: Benchmarks (#8)
- Week 6: Session persistence (#9)

**Deliverable:** Production-ready SDK with testing

---

### Sprint 7-12 (P2 - Quality of Life)
- Week 7: CI/CD pipeline (#10)
- Week 7: Docker support (#11)
- Week 8-9: Middleware system (#12)
- Week 9-10: Retry logic (#13)
- Week 10-11: Resource limits (session pool) (#14)
- Week 11: Configuration validation (#15)

**Deliverable:** Hardened, production-grade SDK

---

### Backlog (P3 - Future)
- HTTP API fallback (#16) - 4-6 weeks
- MCP support (#17) - 8-12 weeks
- Structured outputs (#18) - 3-4 weeks

---

## Success Metrics

### After P0 (Week 2):
- ✅ Users can `import "agent-sdk-go/sdk"`
- ✅ Documentation matches reality
- ✅ Zero deprecated code in main branch

### After P1 (Week 6):
- ✅ 85%+ test coverage (without requiring CLI in CI)
- ✅ Clear API selection guide
- ✅ Zero lost errors in production

### After P2 (Week 12):
- ✅ Automated CI/CD
- ✅ Docker deployment ready
- ✅ Production-grade error handling

---

## Risk Mitigation

### Risk: Breaking Changes

**Mitigation:**
- Version all breaking changes
- Provide deprecation notices (6-month window)
- Maintain compatibility shims

### Risk: Claude CLI Changes

**Mitigation:**
- Pin CLI version in docs
- Add CLI version check in SDK
- Monitor CLI releases for breaking changes

### Risk: Resource Exhaustion

**Mitigation:**
- Implement session pooling (#14)
- Add resource limits by default
- Document resource requirements

---

## Summary

**Total Recommendations:** 18
**Critical Path:** P0 → P1 → P2
**Estimated Timeline:** 12 weeks for full implementation
**Highest ROI:** #1 (Public SDK package), #5 (Mock transport), #10 (CI/CD)

**Next Action:** Implement P0 recommendations in Sprint 1-2.
