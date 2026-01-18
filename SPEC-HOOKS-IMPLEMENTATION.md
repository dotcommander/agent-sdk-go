# Feature: Full Hooks Support for Agent SDK Go

**Author**: Claude Code (Spec Creator)
**Date**: 2026-01-17
**Status**: Draft

---

## TL;DR

| Aspect | Detail |
|--------|--------|
| What | Complete hooks implementation for Claude CLI subprocess lifecycle events |
| Why | Enable validation, logging, metrics, and side effects for tool executions and session lifecycle |
| Who | Go SDK users building production agents with governance, auditing, or safety requirements |
| When | Before/after tool execution, session start/end, permission requests, subagent lifecycle |

---

## Problem Statement

**Current state**:
- Hook types exist in `claude/shared/hooks.go` (12 event types with typed inputs/outputs)
- NO mechanism to register hooks with sessions
- NO `WithHooks()` option function
- NO wire-up to Claude CLI subprocess invocation
- Example code in `examples/hooks-lifecycle/main.go` only demonstrates types, not actual execution

**Pain point**:
- Users cannot validate tool inputs before execution (e.g., block writes outside project directory)
- No audit trail for tool executions, session lifecycle events
- No way to inject custom logging, metrics collection, or side effects
- TypeScript SDK demos show hooks as critical feature for production agents - Go SDK lacks parity

**Impact**:
- Go SDK cannot be used for production agents requiring governance
- No safety guardrails for destructive operations
- Missing feature blocks adoption for enterprise use cases

---

## Proposed Solution

### Overview

Implement hooks as Go function handlers that execute at Claude CLI lifecycle events. Hooks receive typed input structs and return typed output structs controlling execution flow. The implementation wires hooks to the Claude CLI via external script invocation (matching TypeScript SDK pattern) or embedded handler execution.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  User Code (Session Creation)                              │
│                                                             │
│  v2.CreateSession(ctx,                                      │
│    v2.WithHooks(map[shared.HookEvent][]HookHandler{        │
│      shared.HookEventPreToolUse: []HookHandler{            │
│        func(input any) (any, error) { ... }                │
│      },                                                     │
│    }),                                                      │
│  )                                                          │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  V2SessionOptions (claude/v2/options.go)                    │
│  + Hooks map[HookEvent][]HookHandler                        │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Session Factory (claude/v2/session.go)                     │
│  - Serializes hooks to CLI arguments OR                     │
│  - Registers hooks with transport for embedded execution    │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Transport Layer (claude/subprocess/transport.go)           │
│  - Receives hook events from CLI stdout (JSON messages)     │
│  - Executes registered Go hook handlers                     │
│  - Returns hook output to CLI via stdin                     │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│  Claude CLI Subprocess                                      │
│  - Emits hook events as JSON to stdout                      │
│  - Reads hook responses from stdin                          │
│  - Continues/blocks tool execution based on response        │
└─────────────────────────────────────────────────────────────┘
```

**Components affected:**

| Component | Change Type | Description |
|-----------|-------------|-------------|
| `claude/shared/hooks.go` | Modified | Add `HookHandler` function type, matcher pattern support |
| `claude/v2/options.go` | Modified | Add `WithHooks()`, `WithPromptHooks()` option functions |
| `claude/v2/types.go` | Modified | Add `Hooks` field to `V2SessionOptions` |
| `claude/subprocess/transport.go` | Modified | Add hook execution engine, stdin/stdout hook protocol |
| `claude/subprocess/hooks.go` | New | Hook executor, matcher logic, message framing |
| `examples/hooks-validation/main.go` | New | Example: file path validation hook |
| `examples/hooks-metrics/main.go` | New | Example: metrics collection hook |

---

## User Stories

### US-1: File Safety Validation

**As a** developer deploying an autonomous agent
**I want** to block file writes outside project directory
**So that** the agent cannot corrupt system files or access sensitive data

**Acceptance Criteria:**
- [ ] Given a PreToolUse hook registered, when agent attempts Write to `/etc/passwd`, then hook blocks with "block" decision
- [ ] Given a PreToolUse hook registered, when agent writes to `./project/file.go`, then hook returns Continue=true
- [ ] Given hook blocks execution, when response reaches CLI, then tool execution is skipped and error returned to agent

### US-2: Audit Logging

**As a** compliance officer
**I want** all tool executions logged with timestamp, user, and result
**So that** we can audit agent actions for regulatory compliance

**Acceptance Criteria:**
- [ ] Given PostToolUse hook registered, when any tool succeeds, then structured log entry written to audit.jsonl
- [ ] Given PostToolUseFailure hook registered, when tool fails, then failure logged with error details
- [ ] Given SessionStart/SessionEnd hooks, when session lifecycle events occur, then session metadata logged

### US-3: Metrics Collection

**As a** DevOps engineer
**I want** to track tool execution latency and error rates
**So that** I can monitor agent performance and reliability

**Acceptance Criteria:**
- [ ] Given PreToolUse/PostToolUse hooks registered, when tool executes, then duration measured and sent to metrics backend
- [ ] Given PostToolUseFailure hook, when tool fails, then error counter incremented by tool name
- [ ] Given hooks return within timeout, when CLI processes response, then metrics do not block agent execution

### US-4: Permission Auto-Approval

**As a** developer
**I want** to auto-approve safe bash commands
**So that** agent doesn't interrupt flow for benign operations

**Acceptance Criteria:**
- [ ] Given PermissionRequest hook with safelist, when agent runs `ls -la`, then hook returns Decision="approve"
- [ ] Given PermissionRequest hook, when agent runs `rm -rf /`, then hook returns Decision="block"
- [ ] Given hook approves permission, when CLI receives response, then tool executes without user prompt

---

## Interface Design

### Go API

```go
// Hook handler function signature
type HookHandler func(input any) (any, error)

// Hook configuration with optional matcher
type HookConfig struct {
    Matcher string        // Regex pattern to match tool names (e.g., "Write|Edit")
    Handler HookHandler   // Function to execute
}

// Usage pattern
hooks := map[shared.HookEvent][]shared.HookConfig{
    shared.HookEventPreToolUse: []shared.HookConfig{
        {
            Matcher: "Write|Edit|MultiEdit",
            Handler: func(input any) (any, error) {
                preToolInput := input.(*shared.PreToolUseHookInput)

                // Validate file path
                filePath := preToolInput.ToolInput["file_path"].(string)
                if !strings.HasPrefix(filePath, preToolInput.Cwd) {
                    return shared.SyncHookOutput{
                        Decision:   "block",
                        StopReason: "Cannot write outside project directory",
                    }, nil
                }

                return shared.SyncHookOutput{Continue: true}, nil
            },
        },
    },
    shared.HookEventPostToolUse: []shared.HookConfig{
        {
            Matcher: ".*", // All tools
            Handler: func(input any) (any, error) {
                postToolInput := input.(*shared.PostToolUseHookInput)
                log.Printf("Tool executed: %s\n", postToolInput.ToolName)
                return shared.SyncHookOutput{Continue: true}, nil
            },
        },
    },
}

session, err := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithHooks(hooks),
)
```

### Hook Message Protocol

**CLI → SDK (Hook Event):**
```json
{
  "type": "hook_event",
  "hook_event_name": "PreToolUse",
  "session_id": "session-123",
  "transcript_path": "/tmp/transcript.json",
  "cwd": "/home/user/project",
  "tool_name": "Write",
  "tool_input": {
    "file_path": "/home/user/project/main.go",
    "content": "package main\n"
  },
  "tool_use_id": "tool-use-abc123"
}
```

**SDK → CLI (Hook Response):**
```json
{
  "type": "hook_response",
  "tool_use_id": "tool-use-abc123",
  "continue": true,
  "suppress_output": false,
  "decision": "approve",
  "system_message": "",
  "reason": ""
}
```

---

## Implementation Notes

### Approach

**Two-phase implementation:**

1. **Phase 1: External Script Hooks** (matches TypeScript SDK pattern)
   - Serialize hooks to temporary scripts in `/tmp/hooks-{session-id}/`
   - Pass script paths to CLI via `--hooks` flag
   - CLI invokes scripts as subprocesses, reads JSON from stdout
   - Pro: No protocol changes, CLI already supports external hooks
   - Con: Extra subprocess overhead, serialization complexity

2. **Phase 2: Embedded Hooks** (optimized for Go)
   - Transport registers hook handlers in memory
   - CLI emits hook events as JSON messages to stdout
   - Transport intercepts hook events, executes Go functions, returns response via stdin
   - Pro: No subprocess overhead, type-safe Go functions
   - Con: Requires bidirectional protocol, may need CLI changes

**Recommendation: Start with Phase 2 (embedded) if CLI supports hook event messages, fallback to Phase 1 if not.**

### Hook Execution Flow

```
1. User registers hooks via WithHooks()
2. Session factory stores hooks in V2SessionOptions
3. Transport.Connect() checks for registered hooks
4. If hooks exist:
   a. Register hook event message parser
   b. Set up hook response writer
5. During session:
   a. CLI emits hook event (e.g., PreToolUse) to stdout
   b. Transport receives message, parses into typed input struct
   c. Transport looks up matching handlers (by event type + tool name matcher)
   d. Transport executes handler(s) sequentially
   e. Transport serializes response to JSON
   f. Transport sends response to CLI via stdin
   g. CLI processes response (continue/block tool execution)
6. Session.Close() cleans up hook resources
```

### Matcher Pattern Logic

```go
// MatchesToolName checks if hook should execute for given tool
func (cfg HookConfig) MatchesToolName(toolName string) bool {
    if cfg.Matcher == "" {
        return true // No matcher = match all
    }
    matched, _ := regexp.MatchString(cfg.Matcher, toolName)
    return matched
}

// Execute hook chain (multiple hooks for same event)
for _, hookCfg := range hooks[eventType] {
    if !hookCfg.MatchesToolName(toolName) {
        continue // Skip non-matching hooks
    }

    output, err := hookCfg.Handler(input)
    if err != nil {
        return hookError(err)
    }

    // If any hook blocks, stop execution
    if syncOut, ok := output.(shared.SyncHookOutput); ok {
        if syncOut.Decision == "block" || !syncOut.Continue {
            return syncOut // Short-circuit on block
        }
    }
}
```

### Dependencies

| Dependency | Type | Notes |
|------------|------|-------|
| `regexp` | Stdlib | Tool name matcher patterns |
| Existing parser registry | Internal | Reuse for hook event messages |
| Claude CLI v2.x+ | External | Must support hook events or external script hooks |

### Alternatives Considered

| Approach | Pros | Cons | Why Not |
|----------|------|------|---------|
| External script hooks only | CLI already supports, no protocol changes | Extra subprocess overhead, slower | Phase 2 embedded is faster if CLI supports hook events |
| HTTP server for hooks | Language-agnostic, remote hooks possible | Complex setup, network overhead | Overkill for local agent use case |
| Modify CLI to accept Go plugin .so files | Native Go execution, no serialization | Requires CGO, platform-specific builds, fragile ABI | Too complex, breaks cross-platform goal |

---

## Test Plan

### Unit Tests

| Test | Input | Expected Output |
|------|-------|-----------------|
| `TestHookConfig_MatchesToolName` | Matcher="Write\|Edit", toolName="Write" | true |
| `TestHookConfig_MatchesToolName` | Matcher="Write\|Edit", toolName="Read" | false |
| `TestHookConfig_MatchesToolName` | Matcher="", toolName="Bash" | true (match all) |
| `TestHookExecutor_BlockDecision` | Hook returns Decision="block" | Tool execution skipped, error returned |
| `TestHookExecutor_ContinueDecision` | Hook returns Continue=true | Tool execution proceeds |
| `TestHookExecutor_ChainExecution` | Multiple hooks registered | All matching hooks execute sequentially |
| `TestHookSerialization` | PreToolUseHookInput | Valid JSON matches CLI expected format |

### Integration Tests

| Scenario | Steps | Expected |
|----------|-------|----------|
| Happy path: file validation passes | 1. Register PreToolUse hook, 2. Agent writes to project dir | Hook approves, file written |
| Block: file validation fails | 1. Register PreToolUse hook, 2. Agent writes to /etc | Hook blocks, error "Cannot write outside project directory" |
| Audit logging | 1. Register PostToolUse hook, 2. Agent executes 3 tools | 3 log entries in audit.jsonl |
| Permission auto-approve | 1. Register PermissionRequest hook, 2. Agent runs `ls` | Hook approves, no user prompt |
| Permission block unsafe | 1. Register PermissionRequest hook, 2. Agent runs `rm -rf /` | Hook blocks, tool not executed |
| Session lifecycle | 1. Register SessionStart/SessionEnd hooks | Hooks execute at session boundaries |
| Matcher filtering | 1. Register hook for "Write\|Edit", 2. Agent uses Read tool | Hook NOT executed |

### Performance Criteria

| Metric | Target | Measurement |
|--------|--------|-------------|
| Hook execution latency (p95) | < 10ms | Time from hook event receipt to response sent |
| Hook execution latency (p99) | < 50ms | Includes user-defined validation logic |
| Memory overhead per hook | < 1MB | Measured via runtime.ReadMemStats |
| Throughput degradation | < 10% | Tools/sec with hooks vs without hooks |

---

## Security Considerations

| Risk | Mitigation |
|------|------------|
| Malicious hook blocks all operations | Hook timeout (default 5s), if exceeded, default to deny |
| Hook leaks sensitive data in logs | Hook handlers are user code - document best practices in examples |
| Hook modifies tool input to bypass validation | Hook output is decision only, cannot modify tool input (by design) |
| Panic in hook handler crashes subprocess | Recover from panics in hook executor, return error to CLI |

---

## Rollback Plan

**If hooks cause issues in production:**

1. **Disable hooks**: Remove `v2.WithHooks()` from session creation, deploy without hooks
2. **Feature flag**: Add `AGENT_SDK_HOOKS_ENABLED=false` env var check in `WithHooks()`
3. **CLI fallback**: If CLI doesn't support hook events, implementation gracefully degrades (logs warning, hooks not executed)

**Data implications:**
- No persistent data created by hooks feature itself
- User-defined hooks may write logs/metrics - cleanup is user responsibility

---

## Open Questions

| Question | Owner | Status |
|----------|-------|--------|
| Does Claude CLI emit hook events as JSON messages to stdout? | SDK developer (check CLI docs/source) | **OPEN** - Need to verify CLI protocol |
| If not, does CLI support external script hooks via `--hooks` flag? | SDK developer | **OPEN** - Fallback implementation path |
| Should hooks be global (all sessions) or per-session? | User feedback | **RESOLVED**: Per-session (via `WithHooks()`) for flexibility |
| Should hook errors fail-closed (deny) or fail-open (allow)? | Security review | **RESOLVED**: Fail-closed (deny) for safety |
| Async hooks support needed in MVP? | User requirements | **RESOLVED**: Defer async to post-MVP, implement sync only |

---

## Implementation Tasks

Tasks sized for 10-20 minute completion.

### Milestone 1: Foundation (Hook Types & Options)

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M1.1 | Add `HookHandler` type to `claude/shared/hooks.go` | - | `claude/shared/hooks.go` |
| M1.2 | Add `HookConfig` struct with `Matcher` + `Handler` fields | M1.1 | `claude/shared/hooks.go` |
| M1.3 | Add `MatchesToolName()` method to `HookConfig` | M1.2 | `claude/shared/hooks.go` |
| M1.4 | Write unit tests for `MatchesToolName()` | M1.3 | `claude/shared/hooks_test.go` |
| M1.5 | Add `Hooks` field to `V2SessionOptions` struct | - | `claude/v2/types.go` |
| M1.6 | Add `WithHooks()` option function | M1.5 | `claude/v2/options.go` |
| M1.7 | Add `WithPromptHooks()` option function | M1.5 | `claude/v2/options.go` |
| M1.8 | Write unit tests for hook options | M1.6, M1.7 | `claude/v2/options_test.go` |

### Milestone 2: Hook Message Parsing

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M2.1 | Create `HookEventMessage` struct in `claude/shared/hooks.go` | M1.2 | `claude/shared/hooks.go` |
| M2.2 | Add parser for `hook_event` message type | M2.1 | `claude/parser/registry.go` |
| M2.3 | Create `HookResponseMessage` struct | M2.1 | `claude/shared/hooks.go` |
| M2.4 | Add serializer for hook response messages | M2.3 | `claude/shared/hooks.go` |
| M2.5 | Write unit tests for hook message parsing | M2.2, M2.4 | `claude/parser/hooks_test.go` |

### Milestone 3: Hook Executor Engine

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M3.1 | Create `claude/subprocess/hooks.go` file | - | `claude/subprocess/hooks.go` |
| M3.2 | Add `HookExecutor` struct with registered hooks map | M3.1 | `claude/subprocess/hooks.go` |
| M3.3 | Implement `RegisterHooks()` method | M3.2 | `claude/subprocess/hooks.go` |
| M3.4 | Implement `ExecuteHook()` method with matcher filtering | M3.3, M1.3 | `claude/subprocess/hooks.go` |
| M3.5 | Add panic recovery in hook execution | M3.4 | `claude/subprocess/hooks.go` |
| M3.6 | Add hook timeout (5s default) with context | M3.4 | `claude/subprocess/hooks.go` |
| M3.7 | Implement hook chain execution (all matching hooks) | M3.4 | `claude/subprocess/hooks.go` |
| M3.8 | Add short-circuit on "block" decision | M3.7 | `claude/subprocess/hooks.go` |
| M3.9 | Write unit tests for `HookExecutor` | M3.2-M3.8 | `claude/subprocess/hooks_test.go` |

### Milestone 4: Transport Integration

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M4.1 | Add `hookExecutor *HookExecutor` field to `Transport` struct | M3.2 | `claude/subprocess/transport.go` |
| M4.2 | Initialize `HookExecutor` in `Connect()` if hooks registered | M4.1, M3.3 | `claude/subprocess/transport.go` |
| M4.3 | Modify `handleStdout()` to detect hook event messages | M2.1 | `claude/subprocess/transport.go` |
| M4.4 | Add `handleHookEvent()` method to execute hooks and send response | M4.3, M3.4 | `claude/subprocess/transport.go` |
| M4.5 | Add stdin write logic for hook responses | M4.4, M2.4 | `claude/subprocess/transport.go` |
| M4.6 | Add error handling for hook execution failures | M4.4 | `claude/subprocess/transport.go` |
| M4.7 | Write integration test for hook event flow | M4.1-M4.6 | `claude/subprocess/transport_test.go` |

### Milestone 5: Session Plumbing

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M5.1 | Pass hooks from `V2SessionOptions` to transport config | M1.6, M4.1 | `claude/v2/session.go` |
| M5.2 | Add hooks to `TransportConfig` struct | M5.1 | `claude/subprocess/transport.go` |
| M5.3 | Wire hooks to subprocess args (if external script mode) | M5.2 | `claude/subprocess/transport.go` |
| M5.4 | Add session-level hook registration in factory | M5.1 | `claude/v2/factory.go` |
| M5.5 | Write integration test for session with hooks | M5.1-M5.4 | `claude/v2/session_test.go` |

### Milestone 6: Examples & Documentation

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M6.1 | Create `examples/hooks-validation/main.go` | M5.4 | `examples/hooks-validation/main.go` |
| M6.2 | Implement file path validation example | M6.1 | `examples/hooks-validation/main.go` |
| M6.3 | Create `examples/hooks-metrics/main.go` | M5.4 | `examples/hooks-metrics/main.go` |
| M6.4 | Implement metrics collection example | M6.3 | `examples/hooks-metrics/main.go` |
| M6.5 | Create `examples/hooks-audit/main.go` | M5.4 | `examples/hooks-audit/main.go` |
| M6.6 | Implement audit logging example (JSONL) | M6.5 | `examples/hooks-audit/main.go` |
| M6.7 | Update `examples/hooks-lifecycle/main.go` to show actual execution | M5.4 | `examples/hooks-lifecycle/main.go` |
| M6.8 | Create `docs/HOOKS.md` with complete guide | M6.1-M6.7 | `docs/HOOKS.md` |
| M6.9 | Add hooks section to `README.md` | M6.8 | `README.md` |

### Milestone 7: Testing & Validation

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M7.1 | Write table-driven tests for all 12 hook event types | M3.9 | `claude/subprocess/hooks_test.go` |
| M7.2 | Add integration test: PreToolUse blocks file write | M5.5 | `claude/v2/hooks_integration_test.go` |
| M7.3 | Add integration test: PostToolUse audit logging | M5.5 | `claude/v2/hooks_integration_test.go` |
| M7.4 | Add integration test: PermissionRequest auto-approve | M5.5 | `claude/v2/hooks_integration_test.go` |
| M7.5 | Add benchmark for hook execution latency | M3.9 | `claude/subprocess/hooks_bench_test.go` |
| M7.6 | Run full test suite with coverage report | M7.1-M7.5 | Terminal output |
| M7.7 | Verify coverage >80% for new hook code | M7.6 | Terminal output |

### Summary

| Milestone | Tasks | Focus |
|-----------|-------|-------|
| M1 | 8 | Hook types, options, matcher logic |
| M2 | 5 | Message parsing and serialization |
| M3 | 9 | Hook executor engine with timeout/recovery |
| M4 | 7 | Transport integration and stdio protocol |
| M5 | 5 | Session factory plumbing |
| M6 | 9 | Examples and documentation |
| M7 | 7 | Testing and validation |

**Total: 7 milestones, 50 tasks**

### MVP Cutoff

For minimal working feature (file validation use case):
- Complete: M1-M5 (34 tasks)
- Deferred: M6-M7 (examples/docs can be added after core works)

**Critical path:** M1 → M2 → M3 → M4 → M5 (types → parsing → execution → transport → session)

---

## References

- [TypeScript SDK Hooks Demo](https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hello-world)
- [Existing Hook Types](claude/shared/hooks.go)
- [Hook Lifecycle Example](examples/hooks-lifecycle/main.go)
- [V2 Session Options](claude/v2/options.go)
- [Transport Layer](claude/subprocess/transport.go)
