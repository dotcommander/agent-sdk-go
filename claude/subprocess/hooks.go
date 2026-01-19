// Package subprocess provides subprocess communication with the Claude CLI.
package subprocess

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

const (
	// DefaultHookTimeout is the default timeout for hook execution.
	DefaultHookTimeout = 30 * time.Second
)

// HookExecutor manages hook registration and execution.
// It provides timeout protection, panic recovery, and tool name matching.
type HookExecutor struct {
	hooks   map[shared.HookEvent][]shared.HookConfig
	mu      sync.RWMutex
	timeout time.Duration
}

// NewHookExecutor creates a new HookExecutor with the given hooks.
func NewHookExecutor(hooks []shared.HookConfig) *HookExecutor {
	executor := &HookExecutor{
		hooks:   make(map[shared.HookEvent][]shared.HookConfig),
		timeout: DefaultHookTimeout,
	}

	for _, hook := range hooks {
		executor.hooks[hook.Event] = append(executor.hooks[hook.Event], hook)
	}

	return executor
}

// SetTimeout sets the default timeout for hook execution.
func (e *HookExecutor) SetTimeout(timeout time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.timeout = timeout
}

// RegisterHook adds a new hook configuration.
func (e *HookExecutor) RegisterHook(hook shared.HookConfig) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.hooks[hook.Event] = append(e.hooks[hook.Event], hook)
}

// HasHooks returns true if any hooks are registered for the given event.
func (e *HookExecutor) HasHooks(event shared.HookEvent) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.hooks[event]) > 0
}

// ExecuteHook executes all matching hooks for the given event.
// Returns the combined hook output and any error encountered.
// If any hook returns a block decision, execution stops and that decision is returned.
//
// Features:
// - Timeout protection: hooks that exceed timeout are cancelled
// - Panic recovery: panics in hooks are caught and converted to errors
// - Tool name matching: only hooks with matching patterns are executed
// - Fail-closed: errors result in block decisions for safety
func (e *HookExecutor) ExecuteHook(ctx context.Context, event shared.HookEvent, input any, toolName string) (*shared.SyncHookOutput, error) {
	e.mu.RLock()
	hooks := e.hooks[event]
	defaultTimeout := e.timeout
	e.mu.RUnlock()

	if len(hooks) == 0 {
		// No hooks registered - continue by default
		return &shared.SyncHookOutput{Continue: true}, nil
	}

	// Execute each matching hook in sequence
	for _, hook := range hooks {
		// Check tool name matcher for tool-related events
		if toolName != "" && !hook.MatchesToolName(toolName) {
			continue // Skip non-matching hooks
		}

		// Determine timeout for this hook
		timeout := hook.Timeout
		if timeout == 0 {
			timeout = defaultTimeout
		}

		// Execute hook with timeout and panic recovery
		output, err := e.executeWithProtection(ctx, hook, input, timeout)
		if err != nil {
			// Fail-closed: convert errors to block decisions
			return &shared.SyncHookOutput{
				Decision: "block",
				Reason:   fmt.Sprintf("hook error: %v", err),
			}, err
		}

		// Check for block decision
		if output != nil {
			if output.Decision == "block" || output.StopReason != "" {
				return output, nil // Short-circuit on block
			}
			if !output.Continue && output.Decision != "approve" {
				// Explicit continue=false without approve decision = block
				return output, nil
			}
		}
	}

	// All hooks passed - continue
	return &shared.SyncHookOutput{Continue: true}, nil
}

// executeWithProtection executes a single hook with timeout and panic recovery.
func (e *HookExecutor) executeWithProtection(ctx context.Context, hook shared.HookConfig, input any, timeout time.Duration) (output *shared.SyncHookOutput, err error) {
	// Create timeout context
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Channel for result
	type result struct {
		output *shared.SyncHookOutput
		err    error
	}
	resultChan := make(chan result, 1)

	// Execute in goroutine for panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				resultChan <- result{
					output: nil,
					err:    fmt.Errorf("hook panic: %v", r),
				}
			}
		}()

		out, execErr := hook.Handler(ctx, input)
		resultChan <- result{output: out, err: execErr}
	}()

	// Wait for result or timeout
	select {
	case res := <-resultChan:
		return res.output, res.err
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("hook timeout after %v", timeout)
		}
		return nil, ctx.Err()
	}
}

// BuildTypedInput constructs a typed hook input struct from a HookEventMessage.
// This is the canonical implementation for converting hook event data to typed inputs.
// Both streaming hooks (via HookExecutor) and control protocol hooks (via Protocol)
// should use this function to ensure consistent behavior.
func BuildTypedInput(msg *shared.HookEventMessage) (any, error) {
	base := shared.BaseHookInput{
		SessionID:      msg.SessionID,
		TranscriptPath: msg.TranscriptPath,
		Cwd:            msg.Cwd,
		PermissionMode: msg.PermissionMode,
	}

	switch shared.HookEvent(msg.HookEventName) {
	case shared.HookEventPreToolUse:
		return &shared.PreToolUseHookInput{
			BaseHookInput: base,
			HookEventName: msg.HookEventName,
			ToolName:      msg.ToolName,
			ToolInput:     msg.ToolInput,
			ToolUseID:     msg.ToolUseID,
		}, nil

	case shared.HookEventPostToolUse:
		return &shared.PostToolUseHookInput{
			BaseHookInput: base,
			HookEventName: msg.HookEventName,
			ToolName:      msg.ToolName,
			ToolInput:     msg.ToolInput,
			ToolResponse:  msg.ToolResponse,
			ToolUseID:     msg.ToolUseID,
		}, nil

	case shared.HookEventPostToolUseFailure:
		return &shared.PostToolUseFailureHookInput{
			BaseHookInput: base,
			HookEventName: msg.HookEventName,
			ToolName:      msg.ToolName,
			ToolInput:     msg.ToolInput,
			ToolUseID:     msg.ToolUseID,
			Error:         msg.Error,
		}, nil

	case shared.HookEventSessionStart:
		return &shared.SessionStartHookInput{
			BaseHookInput: base,
			HookEventName: msg.HookEventName,
			Source:        msg.Source,
			AgentType:     msg.AgentType,
			Model:         msg.Model,
		}, nil

	case shared.HookEventSessionEnd:
		return &shared.SessionEndHookInput{
			BaseHookInput: base,
			HookEventName: msg.HookEventName,
			Reason:        msg.Reason,
		}, nil

	case shared.HookEventPermissionRequest:
		return &shared.PermissionRequestHookInput{
			BaseHookInput: base,
			HookEventName: msg.HookEventName,
			ToolName:      msg.ToolName,
			ToolInput:     msg.ToolInput,
		}, nil

	case shared.HookEventNotification:
		return &shared.NotificationHookInput{
			BaseHookInput:    base,
			HookEventName:    msg.HookEventName,
			Message:          msg.Message,
			Title:            msg.Title,
			NotificationType: msg.NotificationType,
		}, nil

	case shared.HookEventUserPromptSubmit:
		return &shared.UserPromptSubmitHookInput{
			BaseHookInput: base,
			HookEventName: msg.HookEventName,
			Prompt:        msg.Prompt,
		}, nil

	case shared.HookEventStop:
		return &shared.StopHookInput{
			BaseHookInput:  base,
			HookEventName:  msg.HookEventName,
			StopHookActive: msg.StopHookActive,
		}, nil

	case shared.HookEventSubagentStart:
		return &shared.SubagentStartHookInput{
			BaseHookInput: base,
			HookEventName: msg.HookEventName,
			AgentID:       msg.AgentID,
			AgentType:     msg.AgentType,
		}, nil

	case shared.HookEventSubagentStop:
		return &shared.SubagentStopHookInput{
			BaseHookInput:       base,
			HookEventName:       msg.HookEventName,
			StopHookActive:      msg.StopHookActive,
			AgentID:             msg.AgentID,
			AgentTranscriptPath: msg.AgentTranscriptPath,
		}, nil

	case shared.HookEventPreCompact:
		return &shared.PreCompactHookInput{
			BaseHookInput:      base,
			HookEventName:      msg.HookEventName,
			Trigger:            msg.Trigger,
			CustomInstructions: msg.CustomInstructions,
		}, nil

	default:
		return nil, fmt.Errorf("unknown hook event type: %s", msg.HookEventName)
	}
}

// BuildHookResponse constructs a HookOutput from a SyncHookOutput.
func BuildHookResponse(output *shared.SyncHookOutput, toolUseID string) *shared.HookOutput {
	if output == nil {
		return &shared.HookOutput{
			Type:      "hook_response",
			ToolUseID: toolUseID,
			Continue:  true,
		}
	}

	return &shared.HookOutput{
		Type:           "hook_response",
		ToolUseID:      toolUseID,
		Continue:       output.Continue,
		SuppressOutput: output.SuppressOutput,
		Decision:       output.Decision,
		StopReason:     output.StopReason,
		SystemMessage:  output.SystemMessage,
		Reason:         output.Reason,
	}
}
