# Feature: TypeScript SDK Demos Port to Go

**Author**: Claude Code (Spec Creator Agent)
**Date**: 2026-01-17
**Status**: Draft

---

## TL;DR

| Aspect | Detail |
|--------|--------|
| What | Port all TypeScript examples from claude-agent-sdk-demos to Go equivalents |
| Why | Provide reference implementations for agent-sdk-go users, demonstrate SDK capabilities |
| Who | Go developers building Claude Code agents |
| When | After cloning demos repo, for each example directory |

---

## Problem Statement

**Current state**: The official `claude-agent-sdk-demos` repository contains TypeScript examples demonstrating SDK patterns. Our `agent-sdk-go` has the underlying SDK implementation but lacks comprehensive examples.

**Pain point**:
- Go users have no reference implementations
- TypeScript patterns don't directly translate without guidance
- Missing demonstrations of subprocess transport, streaming, tools, and session management

**Impact**:
- Reduced adoption of agent-sdk-go
- Users must reverse-engineer patterns from TypeScript examples
- No validation that Go SDK achieves feature parity

---

## Proposed Solution

### Overview

Clone the TypeScript demos repository, analyze each example's purpose and implementation, then create equivalent Go implementations using our `internal/claude/` packages. Each port maintains the same demonstration goal while adapting to Go idioms (interfaces, goroutines, error handling).

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│ claude-agent-sdk-demos (TypeScript)                         │
│ https://github.com/anthropics/claude-agent-sdk-demos        │
└────────────────┬────────────────────────────────────────────┘
                 │ Clone to /tmp
                 │ Analyze each example/
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ Port Strategy (Per Example)                                 │
│ 1. Identify demo purpose (tools/streaming/sessions/etc)     │
│ 2. Map TypeScript SDK calls → Go SDK equivalents            │
│ 3. Adapt async/promise patterns → goroutines/channels       │
│ 4. Implement using internal/claude/* packages               │
└────────────────┬────────────────────────────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────────────────────────────┐
│ agent-sdk-go/examples/                                      │
│ ├── basic-query/           # Simple one-shot query          │
│ ├── streaming/             # Real-time response handling    │
│ ├── custom-tools/          # Tool registration & execution  │
│ ├── multi-turn-session/    # Interactive conversations      │
│ ├── mcp-integration/       # MCP server usage               │
│ ├── hooks-lifecycle/       # Hook event handling            │
│ ├── permission-modes/      # Permission system demo         │
│ └── subprocess-advanced/   # Low-level transport control    │
└─────────────────────────────────────────────────────────────┘
```

**Components affected:**

| Component | Change Type | Description |
|-----------|-------------|-------------|
| `examples/` directory | New | Create directory structure for ported demos |
| Each `examples/*/main.go` | New | Runnable Go program demonstrating pattern |
| Each `examples/*/README.md` | New | Port-specific documentation |
| `examples/README.md` | New | Index of all examples with quick start |
| `docs/examples-guide.md` | New | Developer guide for example patterns |

---

## User Stories

### US-1: Quick Start Example

**As a** new agent-sdk-go user
**I want** a simple "hello world" example
**So that** I can verify my setup and understand basic usage

**Acceptance Criteria:**
- [ ] Given Claude CLI is installed, when I run `go run examples/basic-query/main.go`, then I see a response
- [ ] Given I read `examples/basic-query/README.md`, when I follow the steps, then I understand initialization, query, and response handling
- [ ] Given no prior Go experience, when I copy-paste the example, then it runs without modification

### US-2: Streaming Response Example

**As a** developer building interactive UIs
**I want** to see how to handle streaming responses
**So that** I can display incremental output to users

**Acceptance Criteria:**
- [ ] Given a streaming query, when messages arrive, then I process `TextDelta`, `ThinkingDelta`, `ToolUse` events
- [ ] Given channel-based API, when errors occur, then I handle via error channel pattern
- [ ] Given incomplete stream, when context is cancelled, then cleanup happens gracefully

### US-3: Custom Tool Example

**As a** developer extending Claude's capabilities
**I want** to register and execute custom tools
**So that** Claude can interact with my application's domain

**Acceptance Criteria:**
- [ ] Given a tool definition, when I register via `RegisterTool`, then Claude can invoke it
- [ ] Given tool invocation, when my executor runs, then I return structured results
- [ ] Given tool errors, when execution fails, then error is propagated to Claude correctly

### US-4: Multi-Turn Session Example

**As a** developer building conversational agents
**I want** to maintain context across turns
**So that** I can create coherent multi-step interactions

**Acceptance Criteria:**
- [ ] Given an active session, when I send multiple messages, then context is preserved
- [ ] Given session state, when I need to persist, then I can serialize and restore
- [ ] Given session control, when I need to interrupt, then I can cancel gracefully

### US-5: MCP Integration Example

**As a** developer integrating MCP servers
**I want** to see MCP server configuration and usage
**So that** I can extend Claude with external capabilities

**Acceptance Criteria:**
- [ ] Given MCP server config (stdio/SSE/HTTP), when I configure client, then server is available
- [ ] Given MCP tools, when Claude invokes them, then requests route correctly
- [ ] Given MCP lifecycle, when sessions start/end, then servers connect/disconnect properly

---

## Demo Repository Analysis (Requires Clone)

**Step 1: Clone and explore**

```bash
cd /tmp
git clone https://github.com/anthropics/claude-agent-sdk-demos.git
cd claude-agent-sdk-demos
find . -type f -name "*.ts" -o -name "*.js" | head -20
```

**Step 2: Categorize examples**

For each example directory:
1. Identify purpose (based on directory name, README, main entry point)
2. List TypeScript SDK features used
3. Note special patterns (async/await, event emitters, decorators)
4. Document expected inputs/outputs

**Expected demo categories** (common in SDK demo repos):

| Category | Likely Examples | SDK Features Demonstrated |
|----------|-----------------|---------------------------|
| **Basics** | hello-world, simple-query | Client initialization, one-shot query |
| **Streaming** | streaming-chat, real-time | Message channels, event handling |
| **Tools** | calculator, web-search, custom-tool | Tool registration, ToolExecutor interface |
| **Sessions** | multi-turn, context-management | Session lifecycle, state persistence |
| **MCP** | mcp-stdio, mcp-sse, mcp-http | MCP server configs, tool routing |
| **Hooks** | lifecycle-events, permission-hooks | Hook registration, event types |
| **Advanced** | resume-session, fork-session, limits | Session control, budget management |

---

## TypeScript → Go Translation Patterns

### Pattern 1: Async/Await → Goroutines/Channels

**TypeScript:**
```typescript
const response = await client.query("Hello");
console.log(response.content);
```

**Go:**
```go
response, err := client.Query(ctx, "Hello")
if err != nil {
    log.Fatal(err)
}
fmt.Println(response.Content)
```

### Pattern 2: Event Emitters → Channel Selects

**TypeScript:**
```typescript
client.on('message', (msg) => console.log(msg));
client.on('error', (err) => console.error(err));
```

**Go:**
```go
msgChan, errChan := client.QueryStream(ctx, "Hello")
for {
    select {
    case msg, ok := <-msgChan:
        if !ok { return }
        fmt.Printf("%+v\n", msg)
    case err := <-errChan:
        log.Printf("Error: %v", err)
    }
}
```

### Pattern 3: Promise.all → sync.WaitGroup

**TypeScript:**
```typescript
const results = await Promise.all([
    client.query("Q1"),
    client.query("Q2"),
]);
```

**Go:**
```go
var wg sync.WaitGroup
results := make([]string, 2)

wg.Add(2)
go func() {
    defer wg.Done()
    resp, _ := client.Query(ctx, "Q1")
    results[0] = resp.Content
}()
go func() {
    defer wg.Done()
    resp, _ := client.Query(ctx, "Q2")
    results[1] = resp.Content
}()
wg.Wait()
```

### Pattern 4: Class Methods → Interface Implementation

**TypeScript:**
```typescript
class MyTool implements ToolExecutor {
    async execute(args: any): Promise<any> {
        return { result: "done" };
    }
}
```

**Go:**
```go
type MyTool struct{}

func (t *MyTool) Execute(ctx context.Context, toolName string, args map[string]any) (any, error) {
    return map[string]any{"result": "done"}, nil
}
```

### Pattern 5: Decorators → Functional Options

**TypeScript:**
```typescript
@WithTimeout(30000)
@WithModel("claude-sonnet-4")
class MySession extends Session {}
```

**Go:**
```go
session, err := v2.CreateSession(ctx,
    v2.WithTimeout(30*time.Second),
    v2.WithModel("claude-sonnet-4-20250514"),
)
```

---

## Implementation Plan

### Phase 1: Repository Clone & Analysis

**Prerequisites:**
- Claude Code CLI authenticated
- Git installed
- Go 1.21+ installed

**Steps:**
```bash
# Clone demos to temporary location
cd /tmp
rm -rf claude-agent-sdk-demos  # Clean previous clone
git clone https://github.com/anthropics/claude-agent-sdk-demos.git

# Analyze structure
cd claude-agent-sdk-demos
tree -L 2  # View directory structure
find . -name "package.json" -exec cat {} \;  # Check dependencies
find . -name "README.md" -exec echo "=== {} ===" \; -exec cat {} \;  # Read docs
```

**Outputs:**
- Demo inventory table (path, purpose, SDK features)
- Dependency list (which TypeScript SDK features are used)
- Priority ranking (which demos to port first)

### Phase 2: Create Examples Directory Structure

```bash
cd /Users/vampire/go/src/agent-sdk-go
mkdir -p examples/{basic-query,streaming,custom-tools,multi-turn-session,mcp-integration,hooks-lifecycle,permission-modes,subprocess-advanced}
```

### Phase 3: Port Each Demo (Iterative)

For each demo in priority order:

1. **Read TypeScript source**
   - Identify main entry point (`index.ts`, `main.ts`)
   - Note imports from `@anthropic-ai/claude-agent-sdk`
   - Document program flow (init → configure → execute → output)

2. **Map to Go equivalents**
   - TypeScript SDK types → `internal/claude/shared/types.go`
   - Client methods → `claude/client.go` or `claude/v2/session.go`
   - Tool interfaces → `shared/tools/` schemas + ToolExecutor
   - MCP configs → `shared/mcp.go` types
   - Hooks → `shared/hooks.go` event types

3. **Implement Go version**
   - Create `examples/{name}/main.go`
   - Add error handling (no unchecked errors)
   - Use context.Context for cancellation
   - Follow Go idioms (interfaces, composition, explicit errors)

4. **Document**
   - Create `examples/{name}/README.md` with:
     - What this example demonstrates
     - Prerequisites (CLI installed, authenticated)
     - How to run (`go run main.go`)
     - Expected output
     - Key SDK patterns shown
     - Links to relevant docs

5. **Verify**
   - Run: `go build ./examples/{name}`
   - Test: Execute and verify output matches expectations
   - Lint: `go vet ./examples/{name}`
   - Document any differences from TypeScript version

---

## Example Template Structure

Each ported example follows this structure:

```
examples/{demo-name}/
├── main.go              # Runnable Go program
├── README.md            # Demo-specific documentation
├── go.mod               # If demo needs specific dependencies
└── testdata/            # Sample input files if needed
    └── input.txt
```

**main.go template:**
```go
// Package main demonstrates [specific SDK feature]
//
// This example shows:
// - [Key pattern 1]
// - [Key pattern 2]
// - [Key pattern 3]
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/[path]
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/dotcommander/agent-sdk-go/claude"
)

func main() {
    // Initialize client with configuration
    client, err := claude.NewClient(
        claude.WithModel("claude-sonnet-4-20250514"),
        claude.WithTimeout("60s"),
    )
    if err != nil {
        log.Fatalf("Failed to create client: %v", err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
    defer cancel()

    // Example-specific logic here
    runExample(ctx, client)
}

func runExample(ctx context.Context, client *claude.Client) {
    // Implementation
}
```

**README.md template:**
```markdown
# Example: [Demo Name]

## What This Demonstrates

[2-3 sentences about what SDK features this shows]

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

\`\`\`bash
cd examples/[demo-name]
go run main.go
\`\`\`

## Expected Output

\`\`\`
[Sample output]
\`\`\`

## Key Patterns

### Pattern 1: [Name]

[Code snippet with explanation]

### Pattern 2: [Name]

[Code snippet with explanation]

## TypeScript Equivalent

This ports the TypeScript example from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/[path]

## Related Documentation

- [Link to relevant SDK doc]
- [Link to concept guide]
```

---

## Demo Priority Ranking

**Tier 1: Foundation (port first)**
1. **basic-query** - One-shot query/response
2. **streaming** - Real-time message handling
3. **custom-tools** - Tool registration and execution

**Tier 2: Core Features**
4. **multi-turn-session** - Session lifecycle
5. **error-handling** - Graceful failure modes
6. **configuration** - All option patterns

**Tier 3: Advanced**
7. **mcp-integration** - MCP server configs
8. **hooks-lifecycle** - Hook events
9. **permission-modes** - Permission system
10. **session-control** - Resume, fork, limits

---

## Test Plan

### Per-Example Validation

| Test | Steps | Expected |
|------|-------|----------|
| Build succeeds | `go build ./examples/{name}` | No errors, binary created |
| Run succeeds | `./examples/{name}/{name}` | Output matches documented behavior |
| No warnings | `go vet ./examples/{name}` | Clean output |
| Formatting | `gofmt -l examples/{name}` | No changes needed |
| Demonstrates feature | Manual review | Key SDK pattern is clear |

### Integration Test

| Scenario | Steps | Expected |
|----------|-------|----------|
| Fresh clone | Clone repo, run example | Works without modification |
| Multiple models | Swap model via flag/env | Works with sonnet/opus/haiku |
| Context cancellation | CTRL+C during execution | Graceful cleanup |
| Missing CLI | Unset PATH to claude | Clear error message |
| CLI not authenticated | Revoke auth | Clear error message |

### Documentation Test

| Check | Verification |
|-------|--------------|
| README accuracy | Follow steps → matches expected output |
| Links valid | All docs/ references resolve |
| Code comments | Every exported symbol documented |
| Examples index | `examples/README.md` lists all demos |

---

## Open Questions

| Question | Owner | Status |
|----------|-------|--------|
| What demos actually exist in anthropics/claude-agent-sdk-demos? | Investigation | Open - requires clone |
| Do all TypeScript features have Go equivalents? | SDK comparison | Partially resolved - see CLAUDE.md feature status |
| Should examples use public SDK or internal packages? | Architecture decision | **Decision**: Examples use public `claude` package, not `internal/` |
| Where to document TypeScript → Go migration patterns? | Documentation | Open - could be `docs/typescript-migration.md` |
| Should we create integration tests for examples? | Testing strategy | Open - consider `examples_test.go` |

---

## Implementation Tasks

Tasks sized for 10-20 minute completion.

### Milestone 1: Discovery & Foundation

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M1.1 | Clone anthropics/claude-agent-sdk-demos to /tmp | - | `/tmp/claude-agent-sdk-demos/` |
| M1.2 | Generate demo inventory (path, purpose, features) | M1.1 | `DEMO-INVENTORY.md` |
| M1.3 | Create examples/ directory structure | M1.2 | `examples/` dirs |
| M1.4 | Create examples/README.md index | M1.2 | `examples/README.md` |
| M1.5 | Document TypeScript→Go patterns | - | `docs/typescript-migration.md` |

### Milestone 2: Tier 1 Examples (Foundation)

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M2.1 | Port basic-query example | M1.2 | `examples/basic-query/main.go` |
| M2.2 | Write basic-query README | M2.1 | `examples/basic-query/README.md` |
| M2.3 | Test basic-query | M2.1 | Verified working |
| M2.4 | Port streaming example | M1.2 | `examples/streaming/main.go` |
| M2.5 | Write streaming README | M2.4 | `examples/streaming/README.md` |
| M2.6 | Test streaming | M2.4 | Verified working |
| M2.7 | Port custom-tools example | M1.2 | `examples/custom-tools/main.go` |
| M2.8 | Write custom-tools README | M2.7 | `examples/custom-tools/README.md` |
| M2.9 | Test custom-tools | M2.7 | Verified working |

### Milestone 3: Tier 2 Examples (Core Features)

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M3.1 | Port multi-turn-session example | M1.2 | `examples/multi-turn-session/main.go` |
| M3.2 | Write multi-turn-session README | M3.1 | `examples/multi-turn-session/README.md` |
| M3.3 | Test multi-turn-session | M3.1 | Verified working |
| M3.4 | Port error-handling example | M1.2 | `examples/error-handling/main.go` |
| M3.5 | Write error-handling README | M3.4 | `examples/error-handling/README.md` |
| M3.6 | Test error-handling | M3.4 | Verified working |
| M3.7 | Port configuration example | M1.2 | `examples/configuration/main.go` |
| M3.8 | Write configuration README | M3.7 | `examples/configuration/README.md` |
| M3.9 | Test configuration | M3.7 | Verified working |

### Milestone 4: Tier 3 Examples (Advanced)

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M4.1 | Port mcp-integration example | M1.2 | `examples/mcp-integration/main.go` |
| M4.2 | Port hooks-lifecycle example | M1.2 | `examples/hooks-lifecycle/main.go` |
| M4.3 | Port permission-modes example | M1.2 | `examples/permission-modes/main.go` |
| M4.4 | Port session-control example | M1.2 | `examples/session-control/main.go` |
| M4.5 | Write READMEs for all Tier 3 | M4.1-M4.4 | 4 README files |
| M4.6 | Test all Tier 3 examples | M4.1-M4.4 | Verified working |

### Milestone 5: Documentation & Polish

| ID | Task | Inputs | Outputs |
|----|------|--------|---------|
| M5.1 | Update main README with examples section | M2-M4 | `README.md` |
| M5.2 | Create examples integration tests | M2-M4 | `examples/examples_test.go` |
| M5.3 | Add examples to CI pipeline | M5.2 | `.github/workflows/examples.yml` |
| M5.4 | Review all example docs for consistency | M2-M4 | Updated READMEs |
| M5.5 | Add troubleshooting section to examples/README | Common issues | `examples/README.md` |

### Summary

| Milestone | Tasks | Focus |
|-----------|-------|-------|
| M1 | 5 | Discovery, inventory, structure |
| M2 | 9 | Foundation examples (basic, streaming, tools) |
| M3 | 9 | Core features (sessions, errors, config) |
| M4 | 6 | Advanced features (MCP, hooks, permissions) |
| M5 | 5 | Documentation and testing |

**Total: 5 milestones, 34 tasks**

### MVP Cutoff

For minimal working example suite:
- Complete: M1-M2 (14 tasks)
- Demonstrates: Basic usage, streaming, tools
- Deferred: M3-M5 (advanced features, comprehensive docs)

---

## Security Considerations

| Risk | Mitigation |
|------|------------|
| Examples execute arbitrary code | Document that examples are for development only, not production |
| Tool execution in examples | Clearly mark which examples execute external tools |
| MCP server examples | Warn about MCP server security, validate server sources |
| Hardcoded API calls in tools | Use environment variables for any external service credentials |

---

## Rollback Plan

**If ported examples don't work:**

1. **Verification**: Check which examples fail via `make test-examples`
2. **Isolation**: Move failing examples to `examples/wip/`
3. **Documentation**: Update `examples/README.md` to mark experimental status
4. **Communication**: Document known issues in each example's README

**Data implications:**
- Examples don't persist state, so no cleanup needed
- If examples create files, ensure cleanup in defer statements

---

## Next Steps

### Immediate (Do First)

1. **Clone repository**:
   ```bash
   cd /tmp
   git clone https://github.com/anthropics/claude-agent-sdk-demos.git
   ```

2. **Create inventory**:
   ```bash
   cd /tmp/claude-agent-sdk-demos
   # List all examples with descriptions
   find . -type d -maxdepth 1 | tail -n +2 | while read dir; do
       echo "## $(basename $dir)"
       [ -f "$dir/README.md" ] && head -5 "$dir/README.md"
       [ -f "$dir/package.json" ] && jq -r '.description // .name' "$dir/package.json"
       echo ""
   done > /tmp/DEMO-INVENTORY.md
   ```

3. **Begin porting** (use this spec as reference)

### Follow-Up (After Initial Ports)

- Add examples to CI/CD pipeline
- Create video tutorials for complex examples
- Gather user feedback on example clarity
- Expand examples based on common SDK questions

---

## References

- **TypeScript SDK Demos**: https://github.com/anthropics/claude-agent-sdk-demos
- **Claude Code CLI Docs**: https://docs.anthropic.com/en/docs/claude-code
- **agent-sdk-go Docs**: `/Users/vampire/go/src/agent-sdk-go/docs/`
- **SDK Enhancement Plan**: `/Users/vampire/go/src/agent-sdk-go/SDK-ENHANCEMENT-PLAN.md`
- **Original Spec**: `/Users/vampire/go/src/agent-sdk-go/SPEC-ANTHROPIC-SDK-PORT.md`

---

## Appendix A: TypeScript SDK → Go SDK Mapping

| TypeScript SDK | Go SDK Equivalent | Notes |
|----------------|-------------------|-------|
| `import { Client } from '@anthropic-ai/claude-agent-sdk'` | `import "github.com/dotcommander/agent-sdk-go/claude"` | Package name differs |
| `new Client({ model: "..." })` | `claude.NewClient(claude.WithModel("..."))` | Functional options pattern |
| `await client.query(prompt)` | `client.Query(ctx, prompt)` | Context parameter required |
| `client.on('message', handler)` | `msgChan, errChan := client.QueryStream(ctx, prompt)` | Channels instead of events |
| `session.send(message)` | `session.Send(message); resp, err := session.SendMessage(ctx)` | Separate send/receive |
| `ToolExecutor` interface | `shared.ToolExecutor` interface | Same concept, different import |
| `McpServerConfig` | `shared.McpServerConfig` | Same types, from `shared/mcp.go` |
| `HookEvent` types | `shared.HookInput*` / `HookOutput*` | 12 hook event types in `shared/hooks.go` |
| Async iterators | Channel loops with `select` | Go concurrency model |

---

## Appendix B: Demo Complexity Estimation

**Without cloning, estimated demo types and complexity:**

| Demo Type | Estimated Complexity | Port Time | Reason |
|-----------|---------------------|-----------|--------|
| Basic query | Low | 1 hour | Simple client init + query |
| Streaming | Medium | 2 hours | Channel handling, goroutines |
| Custom tools | Medium | 3 hours | Interface implementation, registration |
| Multi-turn session | Medium | 2 hours | Session lifecycle, state management |
| MCP integration | High | 4 hours | Server config, protocol handling |
| Hooks | Medium | 2 hours | Event registration, typed inputs/outputs |
| Permission modes | Low | 1 hour | Configuration options |
| Session control | Medium | 2 hours | Resume, fork, limits |

**Total estimated time**: 17 hours for 8 example categories

**Parallelization**: Can port examples independently after M1 complete

---

## DAWN Validation Checklist

- [x] **D**iagram: Architecture showing TypeScript → Go flow
- [x] **A**ction: Clone commands, port workflow, example templates
- [x] **W**hy-not: TypeScript SDK → Go SDK mapping table, pattern differences
- [x] **N**ext: Immediate steps (clone, inventory, port), verification tests

**Spec Quality Checks:**
- [x] Problem statement quantified (no reference examples in Go SDK)
- [x] Solution architecture diagram included
- [x] All user stories have acceptance criteria
- [x] Interface design covered (example structure, templates)
- [x] Test plan covers build, run, documentation accuracy
- [x] Security risks identified (arbitrary code execution in examples)
- [x] Implementation tasks are 10-20 minute sized
- [x] Tasks have clear inputs and outputs
- [x] MVP cutoff defined (M1-M2 = foundation examples)
- [x] TypeScript → Go translation patterns documented
