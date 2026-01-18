# Example: Hooks Lifecycle

## What This Demonstrates

This example shows the complete hooks system in the Claude Agent SDK. It demonstrates:

- All 12 hook event types and their purposes
- Hook input structures for each event type
- Pre and post tool execution hooks
- Session lifecycle hooks (start, end, stop)
- Subagent lifecycle hooks
- Permission request handling
- Sync and async hook output patterns

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

```bash
cd examples/hooks-lifecycle
go run main.go
```

## Expected Output

```
=== Hook Lifecycle Example ===
This demonstrates all 12 hook event types in the SDK.

--- Available Hook Events ---
   1. PreToolUse
   2. PostToolUse
   3. PostToolUseFailure
   4. Notification
   5. UserPromptSubmit
   6. SessionStart
   7. SessionEnd
   8. Stop
   9. SubagentStart
  10. SubagentStop
  11. PreCompact
  12. PermissionRequest

--- PreToolUse Hook ---
Called BEFORE a tool executes. Can block, modify, or approve.
  Input:
  {
    "session_id": "session-123",
    "tool_name": "Write",
    "tool_input": {"file_path": "/home/user/project/main.go", ...}
  }
...
```

## Key Patterns

### Pattern 1: PreToolUse - Block/Allow Tool Execution

```go
// Validate file paths before write operations
func preToolUseHandler(input shared.PreToolUseHookInput) shared.SyncHookOutput {
    if input.ToolName == "Write" {
        filePath := input.ToolInput["file_path"].(string)
        if !strings.HasPrefix(filePath, input.Cwd) {
            return shared.SyncHookOutput{
                Decision:   "block",
                StopReason: "Cannot write outside project directory",
            }
        }
    }
    return shared.SyncHookOutput{Continue: true}
}
```

### Pattern 2: PostToolUse - Audit Logging

```go
// Log all successful tool executions
func postToolUseHandler(input shared.PostToolUseHookInput) shared.SyncHookOutput {
    log.Printf("Tool %s executed: %v -> %v",
        input.ToolName,
        input.ToolInput,
        input.ToolResponse,
    )
    return shared.SyncHookOutput{Continue: true}
}
```

### Pattern 3: PostToolUseFailure - Error Handling

```go
// Handle tool failures with recovery logic
func postToolUseFailureHandler(input shared.PostToolUseFailureHookInput) shared.SyncHookOutput {
    // Log the failure
    log.Printf("Tool %s failed: %s", input.ToolName, input.Error)

    // Send alert for critical failures
    if input.ToolName == "Bash" && strings.Contains(input.Error, "permission denied") {
        alertSecurityTeam(input)
    }

    return shared.SyncHookOutput{Continue: true}
}
```

### Pattern 4: Session Lifecycle Tracking

```go
// Track session metrics
func sessionStartHandler(input shared.SessionStartHookInput) shared.SyncHookOutput {
    metrics.RecordSessionStart(input.SessionID, input.Model)
    return shared.SyncHookOutput{Continue: true}
}

func sessionEndHandler(input shared.SessionEndHookInput) shared.SyncHookOutput {
    metrics.RecordSessionEnd(input.SessionID, input.Reason)
    return shared.SyncHookOutput{Continue: true}
}
```

### Pattern 5: Permission Request Handling

```go
// Auto-approve safe commands, block others
func permissionRequestHandler(input shared.PermissionRequestHookInput) shared.SyncHookOutput {
    if input.ToolName == "Bash" {
        cmd := input.ToolInput["command"].(string)
        safeCommands := []string{"ls", "cat", "head", "tail", "grep"}

        for _, safe := range safeCommands {
            if strings.HasPrefix(cmd, safe) {
                return shared.SyncHookOutput{Decision: "approve"}
            }
        }

        return shared.SyncHookOutput{
            Decision: "block",
            Reason:   "Command requires manual approval",
        }
    }
    return shared.SyncHookOutput{Decision: "approve"}
}
```

### Pattern 6: Async Hook for External Validation

```go
// Use async hooks for slow external validation
func asyncValidationHandler(input shared.PreToolUseHookInput) shared.AsyncHookOutput {
    // Trigger external validation asynchronously
    go validateExternally(input)

    return shared.AsyncHookOutput{
        Async:        true,
        AsyncTimeout: 30, // Wait up to 30 seconds
    }
}
```

## Hook Events Reference

| Event | When Triggered | Use Cases |
|-------|----------------|-----------|
| `PreToolUse` | Before tool execution | Validation, blocking, input modification |
| `PostToolUse` | After successful tool | Logging, metrics, side effects |
| `PostToolUseFailure` | After failed tool | Error handling, alerts, recovery |
| `Notification` | System notifications | Alerting, logging |
| `UserPromptSubmit` | User sends prompt | Input validation, rate limiting |
| `SessionStart` | Session begins | Initialization, metrics |
| `SessionEnd` | Session ends | Cleanup, reporting |
| `Stop` | Stop requested | Graceful shutdown |
| `SubagentStart` | Subagent spawns | Resource tracking |
| `SubagentStop` | Subagent finishes | Cleanup, aggregation |
| `PreCompact` | Before compaction | Context preservation |
| `PermissionRequest` | Permission needed | Custom approval flows |

## Hook Output Reference

| Output Type | Fields | Description |
|-------------|--------|-------------|
| `SyncHookOutput` | `Continue`, `Decision`, `StopReason`, `SystemMessage`, `SuppressOutput` | Synchronous response |
| `AsyncHookOutput` | `Async`, `AsyncTimeout` | Async validation |

### Decision Values

- `"approve"` - Allow the operation
- `"block"` - Prevent the operation

## TypeScript Equivalent

This ports hook patterns from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/hooks-lifecycle

The TypeScript version uses:
```typescript
const hooks = {
    onPreToolUse: async (input) => {
        if (input.toolName === 'Write') {
            return { decision: 'approve' };
        }
        return { continue: true };
    },
    onPostToolUse: async (input) => {
        console.log(`Tool ${input.toolName} completed`);
    },
};
```

## Related Documentation

- [Hooks Reference](../../docs/usage.md#hooks)
- [Permission System](../../docs/usage.md#permissions)
- [Session Management](../../docs/usage.md#sessions)
