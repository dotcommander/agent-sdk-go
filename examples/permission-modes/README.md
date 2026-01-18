# Example: Permission Modes

## What This Demonstrates

This example shows the complete permission system in the Claude Agent SDK. It demonstrates:

- All 6 permission modes (default, acceptEdits, bypassPermissions, plan, delegate, dontAsk)
- Permission behaviors (allow, deny, ask)
- Permission update destinations (session, local, project, user, CLI arg)
- Permission rules and content filters
- Permission results and modified inputs
- Session configuration patterns for different use cases

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

```bash
cd examples/permission-modes
go run main.go
```

## Expected Output

```
=== Permission Modes Example ===
This demonstrates the permission system in the SDK.

--- Permission Modes ---

  default:
    Description: Standard permission checking with prompts
    Use case: Interactive sessions where user approves actions

  acceptEdits:
    Description: Automatically accept file edits, prompt for others
    Use case: Coding assistants where edits are expected

  bypassPermissions:
    Description: Skip all permission checks
    Use case: Fully automated pipelines, CI/CD environments
...
```

## Key Patterns

### Pattern 1: Default Permission Mode

Standard interactive mode with user prompts:

```go
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithPermissionMode("default"),
)
// User will be prompted for sensitive operations
```

### Pattern 2: Accept Edits Mode

Auto-approve file changes, prompt for shell commands:

```go
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithPermissionMode("acceptEdits"),
)
// File edits happen without prompting
// Shell commands still prompt
```

### Pattern 3: Bypass Permissions (Automated)

Skip all permission checks (use with caution):

```go
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithPermissionMode("bypassPermissions"),
    v2.WithAllowDangerouslySkipPermissions(true), // Required flag
)
// ALL operations execute without prompting
// Only use in sandboxed environments
```

### Pattern 4: Plan Mode (Dry Run)

See what would happen without executing:

```go
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithPermissionMode("plan"),
)
// Claude describes changes but doesn't execute them
// Great for reviewing proposed modifications
```

### Pattern 5: Delegate Mode (Custom Logic)

Use hooks for custom permission logic:

```go
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithPermissionMode("delegate"),
)
// Permission decisions delegated to PreToolUse hooks
```

### Pattern 6: Don't Ask Mode (Headless)

Non-interactive mode that fails safe:

```go
session, _ := v2.CreateSession(ctx,
    v2.WithModel("claude-sonnet-4-5-20250929"),
    v2.WithPermissionMode("dontAsk"),
)
// Never prompts - denies anything not pre-approved
// Safe for automated pipelines
```

### Pattern 7: Permission Rules

Configure fine-grained permissions:

```go
// Allow read-only tools
update := shared.PermissionUpdate{
    Type:     "addRules",
    Behavior: shared.PermissionBehaviorAllow,
    Rules: []shared.PermissionRuleValue{
        {ToolName: "Read"},
        {ToolName: "Glob"},
        {ToolName: "Grep"},
    },
    Destination: shared.PermissionDestSession,
}

// Allow editing only .go files
content := "*.go"
update := shared.PermissionUpdate{
    Type:     "addRules",
    Behavior: shared.PermissionBehaviorAllow,
    Rules: []shared.PermissionRuleValue{
        {ToolName: "Edit", RuleContent: &content},
    },
    Destination: shared.PermissionDestProjectSettings,
}
```

### Pattern 8: Permission Results

Handle permission check results:

```go
// Allow
result := shared.PermissionResult{
    Behavior:  shared.PermissionBehaviorAllow,
    ToolUseID: "tool-use-123",
}

// Deny with message
result := shared.PermissionResult{
    Behavior:  shared.PermissionBehaviorDeny,
    Message:   "Cannot write outside project directory",
    ToolUseID: "tool-use-456",
}

// Deny and interrupt session (critical)
result := shared.PermissionResult{
    Behavior:  shared.PermissionBehaviorDeny,
    Message:   "Security violation detected",
    ToolUseID: "tool-use-789",
    Interrupt: true, // Stops the entire session
}
```

## Permission Modes Reference

| Mode | Description | Risk Level |
|------|-------------|------------|
| `default` | Prompts for sensitive operations | Low |
| `acceptEdits` | Auto-approves file edits | Medium |
| `bypassPermissions` | Skips all checks | HIGH |
| `plan` | No execution, only planning | None |
| `delegate` | Custom hook-based decisions | Variable |
| `dontAsk` | Fails safe on unapproved ops | Low |

## Permission Behaviors Reference

| Behavior | Description |
|----------|-------------|
| `allow` | Permit without prompting |
| `deny` | Block the operation |
| `ask` | Prompt user for decision |

## Permission Destinations Reference

| Destination | Persistence | Scope |
|-------------|-------------|-------|
| `session` | Until session ends | This session only |
| `localSettings` | Local disk | Not shared with team |
| `projectSettings` | Project .claude/ | Shared in repo |
| `userSettings` | User home | All projects |
| `cliArg` | None | Single execution |

## TypeScript Equivalent

This ports permission patterns from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/permission-modes

The TypeScript version uses:
```typescript
const session = await createSession({
    permissionMode: 'acceptEdits',
    allowedTools: ['Read', 'Write', 'Edit'],
});
```

## Related Documentation

- [Permission System](../../docs/usage.md#permissions)
- [Hooks Reference](../../docs/usage.md#hooks)
- [Security Guide](../../docs/usage.md#security)
