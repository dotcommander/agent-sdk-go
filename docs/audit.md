# Codebase Audit - agent-sdk-go

**Date:** 2026-01-16
**Version:** Analyzed current state
**Test Coverage:** 5-93% by package (see breakdown below)

---

## Executive Summary

The `agent-sdk-go` project is a **subprocess-based Go wrapper** around the Claude CLI, NOT a direct HTTP API SDK. This is a critical architectural distinction:

- **Architecture:** Spawns `claude` CLI as subprocess, communicates via stdin/stdout JSON streams
- **Scope:** Dual-purpose (SDK library + CLI tool)
- **Maturity:** Core functionality implemented, advanced features missing
- **Code Quality:** Good structure, comprehensive error handling, strong typing
- **Test Coverage:** 5-93% by package, all tests passing
- **Documentation:** Excellent (README, CLAUDE.md, inline comments)

**Health Score: 72/100**

| Category | Score | Notes |
|----------|-------|-------|
| Architecture | 80 | Clean separation, good interfaces, subprocess dependency |
| Code Quality | 85 | Strong typing, error handling, Go idioms |
| Test Coverage | 55 | 5-93% by package, integration tests require CLI |
| Documentation | 90 | Comprehensive README, CLAUDE.md, examples |
| Completeness | 50 | Core done, advanced features missing |
| Maintainability | 75 | Clear structure, some complexity in transport layer |

**Test Coverage by Package:**
| Package | Coverage | Status |
|---------|----------|--------|
| internal/app | 93% | PASS |
| internal/claude/parser | 83% | PASS |
| internal/claude/cli | 56% | PASS |
| internal/claude/subprocess | 44% | PASS |
| internal/claude/shared | 26% | PASS |
| internal/claude | 14% | PASS |
| internal/claude/v2 | 5% | PASS (integration tests tagged) |

---

## Architecture Analysis

### Architectural Model: Subprocess Wrapper

```
┌─────────────────────────────────────────────┐
│          Application Code                   │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│      internal/claude/client.go              │  ← High-level client interface
│      (ClientImpl)                           │
└──────────────┬──────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────┐
│   internal/claude/subprocess/transport.go   │  ← Process management + I/O
│   (Transport)                               │
└──────────────┬──────────────────────────────┘
               │ spawn + stdio pipes
               ▼
┌─────────────────────────────────────────────┐
│      claude CLI subprocess                  │  ← Anthropic's official CLI
│      (--output-format stream-json)          │
└─────────────────────────────────────────────┘
```

**Key Insight:** This is NOT the TypeScript SDK port mentioned in CLAUDE.md. The TypeScript SDK makes direct HTTP calls to Anthropic's API. This Go SDK wraps the CLI binary.

### Directory Structure

```
agent-sdk-go/
├── cmd/
│   ├── main.go              ← Simple demo (V2 SDK one-shot)
│   └── agent/main.go        ← Full CLI with subcommands
├── internal/
│   ├── app/                 ← Config loading (YAML + env vars)
│   │   └── config.go
│   ├── claude/              ← Core SDK implementation
│   │   ├── client.go        ← Client interface (Query, Connect, etc.)
│   │   ├── types.go         ← Interface definitions
│   │   ├── options.go       ← Functional options pattern
│   │   ├── shared/          ← Message types, errors, validation
│   │   ├── subprocess/      ← Process spawning + I/O handling
│   │   ├── parser/          ← JSON message parsing
│   │   ├── cli/             ← CLI discovery
│   │   └── v2/              ← V2-style session API
└── docs/                    ← Audit outputs (this file)
```

**Strengths:**
- Clean separation of concerns (transport, parsing, client, session)
- Proper use of Go's `internal/` package (prevents external import)
- Interfaces for testability (Client, Transport, Parser)

**Weaknesses:**
- No public SDK package (all code in `internal/`)
- Mixing of V1 (client-based) and V2 (session-based) APIs
- `internal/app/config.go` has deprecated methods

---

## Code Quality Assessment

### Strengths

1. **Error Handling:**
   - Custom error types with `errors.Is`/`errors.As` support
   - Contextual error wrapping (`fmt.Errorf("%w")`)
   - Domain-specific errors (CLINotFoundError, TimeoutError, etc.)

2. **Concurrency Safety:**
   - Proper mutex usage (`sync.RWMutex`)
   - Context-aware goroutines
   - Channel cleanup on shutdown

3. **Go Idioms:**
   - Functional options pattern
   - Interface-based design
   - Table-driven tests (where present)

4. **Type Safety:**
   - Strong typing with `shared.Message` interface
   - Type discrimination in message parsing
   - Generic parsing helpers (`parseMessage[T]`)

### Issues

#### Critical

1. **No Public SDK Package**
   ```
   ❌ import "agent-sdk-go/sdk"  // DOES NOT EXIST
   ✅ import "agent-sdk-go/internal/claude"  // But can't import internal/
   ```
   Users cannot import this as a library without forking or using replace directives.

2. **CLI Dependency**
   - Entire SDK requires Claude CLI installed
   - No fallback to HTTP API
   - Integration tests can't run in CI without CLI

3. **Mixed API Paradigms**
   - V1 API: `client.Connect()`, `client.Query()`
   - V2 API: `v2.CreateSession()`, `session.Send()`/`Receive()`
   - No clear migration path documented

#### Major

1. **Incomplete Tool System**
   ```go
   // From CLAUDE.md - claims this works:
   client.RegisterTool(tool, executor)  // ❌ Method doesn't exist!
   ```
   Tool execution happens in CLI, SDK only observes.

2. **Session Management Confusion**
   - `internal/claude/client.go` has session ID methods
   - `internal/claude/v2/session.go` implements full sessions
   - Unclear which to use when

3. **Deprecated Code**
   ```go
   // internal/app/config.go:78
   func NewClientFromConfig(cfg *Config) (any, error) {
       return nil, fmt.Errorf("use V2 SDK via internal/claude/v2 package instead")
   }
   ```

#### Minor

1. **Magic Numbers**
   - `channelBufferSize = 100` (subprocess/transport.go:27)
   - `BufferSize: 50` (options.go:19)
   - No rationale documented

2. **Input Validation Security**
   ```go
   // subprocess/transport.go:42
   func isValidPrompt(prompt string) bool {
       return !strings.ContainsAny(prompt, "`$!;&|<>")  // Too restrictive?
   }
   ```
   Rejects legitimate prompts with `!` or `<`.

3. **TODO Comments**
   ```go
   // internal/app/config.go:9
   // "agent-sdk-go/internal/sdk"  // TODO: Implement SDK package
   ```

---

## Test Coverage Analysis

### Coverage Claims

- CLAUDE.md: "94.1% coverage"
- README.md: "94.1% coverage (exceeds 85% spec requirement)"

### Test Distribution

| Package | Test Files | Notes |
|---------|-----------|-------|
| `internal/app` | 2 | Config loading, examples |
| `internal/claude` | 1 | Client error tests |
| `internal/claude/shared` | 4 | Message, context, errors, validator |
| `internal/claude/subprocess` | 2 | Transport + integration (requires CLI) |
| `internal/claude/parser` | 2 | JSON parsing, registry |
| `internal/claude/cli` | 1 | CLI discovery |
| `internal/claude/v2` | 2 | Factory, integration (requires CLI) |

**Total: 14 test files for 25 source files (56% file coverage)**

### Test Quality Issues

1. **Integration Tests Require CLI**
   ```go
   // subprocess/transport_integration_test.go
   // v2/v2_integration_test.go
   ```
   These skip if Claude CLI not installed → Coverage drops in CI.

2. **Missing Unit Tests**
   - No tests for `internal/claude/client.go` (core client)
   - No tests for `internal/claude/v2/session.go` (session management)
   - No tests for `cmd/main.go` or `cmd/agent/main.go`

3. **Mock Dependency**
   - No mock Transport for testing client without subprocess
   - No test doubles for subprocess scenarios

---

## Security Assessment

### Strengths

1. **Input Sanitization**
   - `isValidEnvVar()` validates environment variable format
   - `isValidPrompt()` checks for shell injection characters

2. **Subprocess Safety**
   - Uses `exec.CommandContext()` for timeout support
   - Validates environment variables before passing to subprocess
   - Closes pipes properly to prevent resource leaks

### Concerns

1. **Overly Restrictive Validation**
   ```go
   isValidPrompt("`Hello!`")  // ❌ Rejects valid prompt
   ```
   May break legitimate use cases.

2. **No Rate Limiting**
   - Can spawn unlimited Claude CLI processes
   - No connection pooling or reuse

3. **Error Messages Expose Paths**
   ```go
   fmt.Sprintf("Claude CLI not found at %s", e.Path)
   ```
   Leaks system paths in error messages.

---

## Performance Characteristics

### Bottlenecks

1. **Process Spawn Overhead**
   - Each session spawns new `claude` CLI process
   - ~100-500ms startup time per process
   - No process pooling

2. **JSON Parsing on Hot Path**
   ```go
   // subprocess/transport.go:296
   json.Unmarshal([]byte(line), &rawMsg)  // Every line
   ```
   Allocates for every message.

3. **Channel Buffering**
   - Fixed buffer sizes (50-100 messages)
   - May block if consumer slow

### Optimizations

1. **Good: Streaming Architecture**
   - Goroutines for stdout/stderr prevent blocking
   - Context-aware cancellation

2. **Good: Reusable Parsers**
   - Parser registry avoids repeated reflection

---

## Dependency Analysis

### External Dependencies

```go
require (
    github.com/stretchr/testify v1.11.1  // Test assertions
    gopkg.in/yaml.v3 v3.0.1             // Config parsing
)
```

**Minimal dependency footprint** ✓

### Internal Dependencies

- Entire SDK depends on Claude CLI being installed
- No fallback to HTTP API
- Tight coupling to CLI's JSON protocol

---

## Maintainability

### Positive Factors

1. **Clear Separation**
   - Transport layer isolated in `subprocess/`
   - Message types in `shared/`
   - Business logic in `client.go`

2. **Interface-Driven**
   - Easy to add mock implementations
   - Testable without subprocess

3. **Documentation**
   - Excellent README
   - CLAUDE.md for Claude Code integration
   - Inline comments on complex logic

### Negative Factors

1. **API Surface Confusion**
   - V1 vs V2 APIs
   - Deprecated methods still present
   - No migration guide

2. **Implicit Knowledge**
   - Must know CLI flags to understand transport layer
   - JSON protocol not documented in code

3. **No Changelog**
   - No versioning beyond git tags
   - Breaking changes not tracked

---

## Anti-Patterns Detected

1. **God Object:** `ClientOptions` has 15 fields (options.go:13-60)
2. **Magic Strings:** Model names hardcoded in multiple places
3. **Error Shadowing:** Some error paths return generic errors
4. **Premature Optimization:** Buffer size tuning without benchmarks
5. **Dead Code:** Deprecated functions not removed

---

## Technical Debt

### High Priority

1. **Create public SDK package** (prevents library usage)
2. **Document CLI dependency prominently** (users expect HTTP SDK)
3. **Remove deprecated code** or mark with deprecation notices
4. **Add mock Transport** for unit testing

### Medium Priority

1. **Consolidate V1/V2 APIs** or document clearly
2. **Add benchmarks** for buffer size optimization
3. **Implement tool executor registry** (claimed but missing)
4. **Add changelog/versioning**

### Low Priority

1. **Extract magic numbers to constants**
2. **Add more inline examples**
3. **Create godoc-friendly package comments**

---

## Recommendations Summary

See `docs/recommendations.md` for prioritized action items.

Key takeaways:
- ✅ **Solid foundation** for subprocess wrapper
- ⚠️ **Misleading documentation** claims direct SDK port
- ❌ **Cannot be used as library** without public package
- 🔧 **Technical debt manageable** with focused effort

---

## Files Analyzed

- **Source files:** 25 Go files (~9,300 lines)
- **Test files:** 14 Go files
- **Documentation:** README.md, CLAUDE.md, config.example.yaml
- **Configuration:** go.mod, go.sum

**Analysis Tool:** Manual code review + static analysis
**Reviewer:** Claude (Sonnet 4.5)
