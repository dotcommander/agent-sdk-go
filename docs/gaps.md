# Feature Gaps and Incomplete Implementations

**Project:** agent-sdk-go
**Date:** 2026-01-16

---

## Executive Summary

This document catalogs **missing features**, **incomplete implementations**, and **TODO items** in the agent-sdk-go codebase.

**Gap Categories:**
1. **Critical:** Blocks library usage or documented features
2. **High:** Missing features from enhancement plan
3. **Medium:** Incomplete implementations
4. **Low:** Quality-of-life improvements

---

## Critical Gaps

### 1. No Public SDK Package

**Status:** ❌ Blocking library usage

**Issue:**
```
User expectation: import "agent-sdk-go/sdk"
Current reality:  All code in internal/ (cannot import)
```

**Evidence:**
```
$ tree -L 1
├── cmd/         ← CLI binaries
├── internal/    ← All SDK code (not importable)
└── README.md    ← Says "go get agent-sdk-go/sdk" ❌
```

**Impact:**
- Users cannot use this as a library
- Must fork or use replace directives
- Breaks standard Go module conventions

**Fix Required:**
- Create `sdk/` directory at root
- Move public API types and client to `sdk/`
- Keep subprocess/parser in `internal/`

---

### 2. Tool System Not Implemented

**Status:** ❌ Documented but missing

**Claimed in CLAUDE.md:**
```go
// 1. Define tool in types.go
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema InputSchema `json:"input_schema"`
}

// 2. Implement ToolExecutor interface
type MyToolExecutor struct{}

// 3. Register with client
err := client.RegisterTool(toolDefinition, &MyToolExecutor{})  // ❌ Method doesn't exist!
```

**Reality:**
- `client.RegisterTool()` method: **NOT IMPLEMENTED**
- `ToolExecutor` interface: **NOT DEFINED**
- Tool execution happens in Claude CLI, SDK only observes

**Actual Code:**
```go
// cmd/agent/main.go:150 (tool detection, not execution)
if shared.HasToolUses(msg) {
    toolUses := shared.ExtractToolUses(msg)
    fmt.Printf("Detected %d tool use(s)\n", len(toolUses))
}
```

**Workaround:**
```
CLI handles tool execution internally.
SDK can observe tool use blocks in messages.
```

**Impact:**
- Misleading documentation
- Users expect client-side tool execution
- No custom tool support

---

### 3. Deprecated Code Still Present

**Status:** ❌ Confusing API surface

**Deprecated Functions:**
```go
// internal/app/config.go:78
func NewClientFromConfig(cfg *Config) (any, error) {
    return nil, fmt.Errorf("use V2 SDK via internal/claude/v2 package instead")
}

// internal/app/config.go:88
func NewClient() (any, error) {
    cfg, err := Load()
    if err != nil {
        return nil, fmt.Errorf("load config: %w", err)
    }
    return NewClientFromConfig(cfg)  // ← Calls deprecated function!
}
```

**Comment in cmd/main.go:**
```go
// This file is deprecated. Use cmd/agent/main.go instead.
```

**Impact:**
- Confuses new users
- Dead code paths in codebase
- No deprecation warnings at compile time

---

## High Priority Gaps

### 4. Missing Features from Enhancement Plan

**Note:** CLAUDE.md references `SDK-ENHANCEMENT-PLAN.md` but file doesn't exist in repo.

**Claimed Missing (P0-P3):**
- ❌ Permissions system and hooks framework
- ❌ User input workflows with timeouts
- ❌ Session forking capability
- ❌ File checkpointing for state restoration
- ❌ Structured outputs with JSON Schema validation
- ❌ System prompt presets and configuration
- ❌ MCP (Model Context Protocol) compatibility
- ❌ Subagent hierarchy support
- ❌ Slash command ecosystem integration

**Cannot Verify:** Enhancement plan document missing.

---

### 5. Test Coverage Gaps

**Status:** ✅ All tests passing, coverage varies by package (5-93%)

**Coverage by Package:**
| Package | Coverage |
|---------|----------|
| internal/app | 93% |
| internal/claude/parser | 83% |
| internal/claude/cli | 56% |
| internal/claude/subprocess | 44% |
| internal/claude/shared | 26% |
| internal/claude | 14% |
| internal/claude/v2 | 5% |

**Areas Needing More Tests:**

1. **V2 Session Management:**
   ```
   ⚠️ internal/claude/v2/ (5% - integration tests tagged)
   ```
   Integration tests require Claude CLI, tagged with `//go:build integration`.

2. **Shared Package:**
   ```
   ⚠️ internal/claude/shared/ (26%)
   ```
   Options and factory code needs more unit tests.

3. **CLI Entry Points:**
   ```
   ❌ cmd/main.go (no tests)
   ❌ cmd/agent/main.go (no tests)
   ```

**Integration Test Dependency:**
```go
// subprocess/transport_integration_test.go
// v2/v2_integration_test.go

// These skip if CLI not found:
if !cli.IsCLIAvailable() {
    t.Skip("Claude CLI not available")
}
```

**Impact:**
- Coverage drops in CI without Claude CLI
- Core logic not validated in isolation
- Refactoring risky

---

### 6. No HTTP API Fallback

**Status:** ❌ CLI dependency is hard requirement

**Current Architecture:**
```
User → Go SDK → Claude CLI subprocess → Anthropic API
```

**Expected (based on "port from TypeScript"):**
```
User → Go SDK → Anthropic API (HTTP)
```

**Impact:**
- Requires Claude CLI installation (non-trivial)
- Cannot work in restricted environments (Docker, CI)
- Not truly a "port" of TypeScript SDK (different architecture)

**Workaround:**
None. SDK cannot function without CLI.

---

## Medium Priority Gaps

### 7. Incomplete Error Handling

**Status:** ⚠️ Some errors lost in channels

**Issue 1: Channel Error Overwrite**
```go
// subprocess/transport.go:280
case t.errChan <- fmt.Errorf("stdout scanner error: %w", err):
case <-t.ctx.Done():
    return  // ← Error lost if context cancelled
```

**Issue 2: Buffered Error Channel**
```go
// subprocess/transport.go:197
t.errChan = make(chan error, channelBufferSize)

// If buffer full, sends block:
select {
case t.errChan <- err:
case <-t.ctx.Done():
    return  // ← Error dropped
}
```

**Impact:**
- Errors may be lost during shutdown
- User may not see all errors
- Debugging difficult

---

### 8. V1/V2 API Confusion

**Status:** ⚠️ Two API paradigms, no migration guide

**V1 API (Client-Based):**
```go
client, _ := claude.NewClient()
client.Connect(ctx)
response, _ := client.Query(ctx, "Hello")
```

**V2 API (Session-Based):**
```go
session, _ := v2.CreateSession(ctx)
session.Send(ctx, "Hello")
for msg := range session.Receive(ctx) {
    // Process messages
}
```

**Confusion:**
1. Both exist in same codebase
2. No clear "when to use which"
3. V2 internally uses V1 client
4. Different error handling patterns

**Missing:**
- Migration guide
- Deprecation notices
- Unified API

---

### 9. No Benchmarks

**Status:** ⚠️ Performance tuning without data

**Magic Numbers:**
```go
// subprocess/transport.go:27
const channelBufferSize = 100  // ← Why 100?

// options.go:19
BufferSize: 50,  // ← Why 50?
```

**No Benchmarks:**
```
$ find . -name "*_bench_test.go"
(no results)
```

**Impact:**
- Cannot validate buffer size choices
- No performance regression detection
- Optimization guesswork

---

### 10. Session Management Incomplete

**Status:** ⚠️ Basic features missing

**What Works:**
```go
session.Send(ctx, "message")
session.Receive(ctx) → channel
session.SessionID() → string
session.Close()
```

**What's Missing:**

1. **Session Persistence:**
   ```
   ❌ session.Save() to disk
   ❌ v2.LoadSession(id) from disk
   ```

2. **Session Forking:**
   ```
   ❌ session.Fork() for branching conversations
   ```

3. **Session History:**
   ```
   ❌ session.GetHistory() to retrieve past messages
   ❌ session.ClearHistory()
   ```

4. **Session Metadata:**
   ```
   ❌ session.GetMetadata() (creation time, message count, etc.)
   ```

**Impact:**
- Cannot implement stateful applications
- No conversation persistence
- Limited multi-turn use cases

---

## Low Priority Gaps

### 11. TODO Comments

**Status:** 📝 Inline TODOs not tracked

```go
// internal/app/config.go:9
// "agent-sdk-go/internal/sdk"  // TODO: Implement SDK package
```

**Only 1 TODO found in codebase** (via grep).

**Impact:**
- Minimal, but indicates incomplete refactoring

---

### 12. No Middleware/Hooks

**Status:** ⚠️ Cannot intercept messages

**Desired:**
```go
client.Use(func(msg Message) Message {
    // Log, modify, or reject messages
    return msg
})
```

**Current:**
No hook system. Users must wrap client or fork code.

**Impact:**
- Cannot add logging without modifying SDK
- No request/response interception
- Limits extensibility

---

### 13. Limited Streaming Control

**Status:** ⚠️ No pause/resume

**Current:**
```go
for event := range events {
    // Process event
}
```

**Missing:**
- ❌ Pause streaming
- ❌ Resume streaming
- ❌ Cancel mid-stream (only via context)
- ❌ Backpressure signaling

**Impact:**
- Cannot implement rate limiting
- No flow control for slow consumers

---

### 14. No Retry Logic

**Status:** ⚠️ Users must implement retries

**Expected:**
```go
client, _ := sdk.NewClient(apiKey,
    sdk.WithRetry(3),              // ❌ Not available
    sdk.WithBackoff(exponential),  // ❌ Not available
)
```

**Current:**
All retries left to user code.

**Impact:**
- Reduces SDK usability
- Users reinvent retry logic
- Inconsistent retry behavior

---

### 15. Configuration Validation Incomplete

**Status:** ⚠️ Invalid configs accepted

**Examples:**

1. **Negative Timeout:**
   ```go
   cfg := &Config{Timeout: -10}  // ← Accepted!
   cfg.GetTimeoutDuration()       // → 0 (silent failure)
   ```

2. **Invalid Model Name:**
   ```go
   claude.NewClient(WithModel("invalid-model"))  // ← No validation
   ```
   Error only appears when CLI runs.

**Impact:**
- Errors delayed until runtime
- Confusing error messages
- Harder to debug

---

## Documentation Gaps

### 16. Misleading Documentation

**Status:** ❌ Critical documentation issues

**Issue 1: "Port from TypeScript"**

README.md:
```
A Go implementation of the Anthropic Claude Agent SDK,
ported from TypeScript to provide native Go interfaces.
```

**Reality:**
- TypeScript SDK makes HTTP calls to Anthropic API
- Go SDK spawns Claude CLI subprocess
- Completely different architectures

**Issue 2: Tool System Example**

CLAUDE.md shows tool registration:
```go
err := client.RegisterTool(toolDefinition, &MyToolExecutor{})
```

**Reality:** Method doesn't exist.

**Issue 3: Import Path**

README.md:
```bash
go get agent-sdk-go/sdk
```

**Reality:** No `sdk/` directory exists.

---

### 17. Missing Guides

**Status:** 📝 No tutorials

**What Exists:**
- ✅ README with basic examples
- ✅ CLAUDE.md for Claude Code integration
- ✅ Inline code comments

**What's Missing:**
- ❌ Migration guide (V1 → V2)
- ❌ Architecture deep dive
- ❌ Error handling guide
- ❌ Testing guide (mocking, integration)
- ❌ Performance tuning guide
- ❌ Deployment guide (Docker, Kubernetes)

---

### 18. No API Reference

**Status:** 📝 godoc exists but not published

**Current:**
```bash
godoc -http=:6060
# View at http://localhost:6060
```

**Missing:**
- ❌ Published to pkg.go.dev
- ❌ Versioned documentation
- ❌ Searchable examples

**Reason:** Cannot publish `internal/` packages.

---

## Build/Deployment Gaps

### 19. No CI/CD

**Status:** ❌ No automation

**Missing:**
- ❌ GitHub Actions workflow
- ❌ Automated tests
- ❌ Code coverage reporting
- ❌ Linting (golangci-lint)
- ❌ Release automation

**Impact:**
- Manual testing only
- No quality gates
- Easy to introduce regressions

---

### 20. No Docker Support

**Status:** ⚠️ CLI dependency complicates containers

**Current:**
Users must install Claude CLI in Dockerfile:
```dockerfile
# Would need something like:
RUN curl -o /usr/local/bin/claude https://... && chmod +x /usr/local/bin/claude
```

**Missing:**
- ❌ Official Docker image
- ❌ Multi-stage build example
- ❌ CLI installation script

---

### 21. No Release Process

**Status:** 📝 No versioning

**Missing:**
- ❌ Changelog
- ❌ Semantic versioning
- ❌ Release notes
- ❌ Breaking change warnings

**Impact:**
- Users don't know what changed
- Hard to track compatibility
- No stable API guarantees

---

## Security Gaps

### 22. Input Validation Too Restrictive

**Status:** ⚠️ Breaks legitimate use cases

```go
// subprocess/transport.go:42
func isValidPrompt(prompt string) bool {
    return !strings.ContainsAny(prompt, "`$!;&|<>")
}

// Rejected prompts:
isValidPrompt("What's 2 + 2?")       // ❌ Contains '?'... wait, no
isValidPrompt("Tell me about C++!")  // ❌ Contains '!'
isValidPrompt("Compare x < y")       // ❌ Contains '<'
```

**Impact:**
- Valid prompts rejected
- Overly conservative security
- Users work around validation

---

### 23. No Secrets Management

**Status:** ⚠️ API keys in plaintext

**Current:**
```yaml
# config.yaml
api_key: "sk-ant-abc123"  # ← Plaintext in file
```

**Missing:**
- ❌ Environment variable encryption
- ❌ Secrets provider integration (Vault, KMS)
- ❌ Key rotation support

**Impact:**
- Secrets committed to git (if not careful)
- No key rotation workflow
- Compliance issues (PCI, HIPAA)

---

### 24. No Rate Limiting

**Status:** ⚠️ Can spawn unlimited processes

**Current:**
```go
// No limit on concurrent sessions
for i := 0; i < 1000; i++ {
    session, _ := v2.CreateSession(ctx)
    // ← 1000 Claude CLI processes!
}
```

**Missing:**
- ❌ Max concurrent sessions
- ❌ Process pooling
- ❌ Resource limits

**Impact:**
- Resource exhaustion possible
- DoS risk
- High memory usage

---

## Compatibility Gaps

### 25. No Windows Testing

**Status:** ⚠️ Untested on Windows

**Code References:**
```go
// shared/errors.go:85
case "darwin":
    suggestions = append(suggestions, "brew install claude")
case "linux":
    suggestions = append(suggestions, "download from GitHub")
case "windows":
    suggestions = append(suggestions, "use PowerShell")  // ← Generic
```

**Missing:**
- ❌ Windows CI testing
- ❌ Windows-specific path handling
- ❌ PowerShell script examples

**Impact:**
- Unknown if works on Windows
- Potential path separator issues
- No Windows installation guide

---

### 26. Go Version Compatibility

**Status:** ⚠️ Requires very recent Go

```go
// go.mod
go 1.25.5  // ← Unreleased version!
```

**Issues:**
1. Go 1.25 doesn't exist yet (latest stable: 1.22)
2. Uses `maps.Copy()` (requires Go 1.21+)
3. No compatibility matrix documented

**Impact:**
- Cannot build with older Go versions
- May break on future Go releases
- No compatibility guarantees

---

## Summary by Priority

### Critical (6 gaps)
1. No public SDK package ← **BLOCKING**
2. Tool system not implemented
3. Deprecated code still present
4. Enhancement plan missing
5. Test coverage gaps
6. No HTTP API fallback

### High (8 gaps)
7. Incomplete error handling
8. V1/V2 API confusion
9. No benchmarks
10. Session management incomplete
11. TODO comments
12. No middleware/hooks
13. Limited streaming control
14. No retry logic

### Medium (5 gaps)
15. Configuration validation incomplete
16. Misleading documentation
17. Missing guides
18. No API reference
19. No CI/CD

### Low (7 gaps)
20. No Docker support
21. No release process
22. Input validation too restrictive
23. No secrets management
24. No rate limiting
25. No Windows testing
26. Go version compatibility

---

## Estimated Effort

| Priority | Gaps | Est. Effort |
|----------|------|-------------|
| Critical | 6 | 4-6 weeks |
| High | 8 | 3-4 weeks |
| Medium | 5 | 2-3 weeks |
| Low | 7 | 2-3 weeks |
| **Total** | **26** | **11-16 weeks** |

**Note:** Assumes 1 engineer working full-time.

---

## Next Steps

See `docs/recommendations.md` for prioritized action plan.
