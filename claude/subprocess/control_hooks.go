// Package subprocess provides subprocess communication with the Claude CLI.
// This file handles hook callback processing and registration.
package subprocess

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

// ProtocolHookCallback is the function signature for protocol-level hook callbacks.
// This is similar to shared.HookHandler but with additional toolUseID parameter.
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - input: Hook input (shared.PreToolUseHookInput, shared.PostToolUseHookInput, etc.)
//   - toolUseID: Optional tool use identifier (only for tool-related hooks)
//
// Returns:
//   - *shared.SyncHookOutput: The hook's response
//   - error: Non-nil if the callback encounters an error
type ProtocolHookCallback func(
	ctx context.Context,
	input any,
	toolUseID *string,
) (*shared.SyncHookOutput, error)

// ProtocolHookMatcher defines which hooks to trigger for a given pattern.
type ProtocolHookMatcher struct {
	// Matcher is a tool name pattern (e.g., "Bash", "Write|Edit|MultiEdit").
	// Empty string matches all tools.
	Matcher string
	// Hooks are the callbacks to execute when the pattern matches.
	Hooks []ProtocolHookCallback
	// Timeout is the maximum time in seconds for all hooks in this matcher.
	// Default is 60 seconds.
	Timeout *float64
}

// handleHookCallbackRequest processes a hook callback request from CLI.
// Follows the same pattern as handleCanUseToolRequest with panic recovery.
func (p *Protocol) handleHookCallbackRequest(ctx context.Context, requestID string, request map[string]any) error {
	// Parse callback ID
	callbackID, _ := request["callback_id"].(string)
	if callbackID == "" {
		return p.sendErrorResponse(ctx, requestID, "missing callback_id")
	}

	// Parse hook event name from input
	inputData, _ := request["input"].(map[string]any)
	if inputData == nil {
		inputData = make(map[string]any)
	}

	eventName, _ := inputData["hook_event_name"].(string)
	event := shared.HookEvent(eventName)

	// Parse input based on event type
	input := p.parseHookInput(event, inputData)

	// Parse tool_use_id if present
	var toolUseID *string
	if id, ok := request["tool_use_id"].(string); ok {
		toolUseID = &id
	}

	// Get callback (thread-safe read)
	p.hookCallbacksMu.RLock()
	callback, exists := p.hookCallbacks[callbackID]
	p.hookCallbacksMu.RUnlock()

	if !exists {
		return p.sendErrorResponse(ctx, requestID, fmt.Sprintf("callback not found: %s", callbackID))
	}

	// Invoke callback with panic recovery (matches permission callback pattern)
	var result *shared.SyncHookOutput
	var callbackErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				callbackErr = fmt.Errorf("hook callback panicked: %v", r)
			}
		}()
		result, callbackErr = callback(ctx, input, toolUseID)
	}()

	if callbackErr != nil {
		return p.sendErrorResponse(ctx, requestID, fmt.Sprintf("callback error: %v", callbackErr))
	}

	return p.sendHookResponse(ctx, requestID, result)
}

// parseHookInput creates the appropriate typed input based on event type.
// Returns the strongly-typed input struct for the callback.
func (p *Protocol) parseHookInput(event shared.HookEvent, inputData map[string]any) any {
	// Parse base fields
	base := shared.BaseHookInput{
		SessionID:      getString(inputData, "session_id"),
		TranscriptPath: getString(inputData, "transcript_path"),
		Cwd:            getString(inputData, "cwd"),
		PermissionMode: getString(inputData, "permission_mode"),
	}

	switch event {
	case shared.HookEventPreToolUse:
		return &shared.PreToolUseHookInput{
			BaseHookInput: base,
			HookEventName: "PreToolUse",
			ToolName:      getString(inputData, "tool_name"),
			ToolInput:     getMap(inputData, "tool_input"),
			ToolUseID:     getString(inputData, "tool_use_id"),
		}
	case shared.HookEventPostToolUse:
		return &shared.PostToolUseHookInput{
			BaseHookInput: base,
			HookEventName: "PostToolUse",
			ToolName:      getString(inputData, "tool_name"),
			ToolInput:     getMap(inputData, "tool_input"),
			ToolResponse:  inputData["tool_response"],
			ToolUseID:     getString(inputData, "tool_use_id"),
		}
	case shared.HookEventPostToolUseFailure:
		return &shared.PostToolUseFailureHookInput{
			BaseHookInput: base,
			HookEventName: "PostToolUseFailure",
			ToolName:      getString(inputData, "tool_name"),
			ToolInput:     getMap(inputData, "tool_input"),
			ToolUseID:     getString(inputData, "tool_use_id"),
			Error:         getString(inputData, "error"),
		}
	case shared.HookEventUserPromptSubmit:
		return &shared.UserPromptSubmitHookInput{
			BaseHookInput: base,
			HookEventName: "UserPromptSubmit",
			Prompt:        getString(inputData, "prompt"),
		}
	case shared.HookEventStop:
		return &shared.StopHookInput{
			BaseHookInput:  base,
			HookEventName:  "Stop",
			StopHookActive: getBool(inputData, "stop_hook_active"),
		}
	case shared.HookEventSubagentStop:
		return &shared.SubagentStopHookInput{
			BaseHookInput:  base,
			HookEventName:  "SubagentStop",
			StopHookActive: getBool(inputData, "stop_hook_active"),
		}
	case shared.HookEventPreCompact:
		return &shared.PreCompactHookInput{
			BaseHookInput:      base,
			HookEventName:      "PreCompact",
			Trigger:            getString(inputData, "trigger"),
			CustomInstructions: getStringPtr(inputData, "custom_instructions"),
		}
	case shared.HookEventSessionStart:
		return &shared.SessionStartHookInput{
			BaseHookInput: base,
			HookEventName: "SessionStart",
			Source:        getString(inputData, "source"),
			AgentType:     getString(inputData, "agent_type"),
			Model:         getString(inputData, "model"),
		}
	case shared.HookEventSessionEnd:
		return &shared.SessionEndHookInput{
			BaseHookInput: base,
			HookEventName: "SessionEnd",
			Reason:        getString(inputData, "reason"),
		}
	case shared.HookEventNotification:
		return &shared.NotificationHookInput{
			BaseHookInput:    base,
			HookEventName:    "Notification",
			Message:          getString(inputData, "message"),
			Title:            getString(inputData, "title"),
			NotificationType: getString(inputData, "notification_type"),
		}
	case shared.HookEventSubagentStart:
		return &shared.SubagentStartHookInput{
			BaseHookInput: base,
			HookEventName: "SubagentStart",
			AgentID:       getString(inputData, "agent_id"),
			AgentType:     getString(inputData, "agent_type"),
		}
	case shared.HookEventPermissionRequest:
		return &shared.PermissionRequestHookInput{
			BaseHookInput: base,
			HookEventName: "PermissionRequest",
			ToolName:      getString(inputData, "tool_name"),
			ToolInput:     inputData["tool_input"],
		}
	default:
		// Forward compatibility - return raw input for unknown events
		return inputData
	}
}

// sendHookResponse sends a hook callback response back to CLI.
func (p *Protocol) sendHookResponse(ctx context.Context, requestID string, result *shared.SyncHookOutput) error {
	// Build response data from SyncHookOutput
	responseData := make(map[string]any)

	if result != nil {
		if result.Continue {
			responseData["continue"] = result.Continue
		}
		if result.SuppressOutput {
			responseData["suppressOutput"] = result.SuppressOutput
		}
		if result.StopReason != "" {
			responseData["stopReason"] = result.StopReason
		}
		if result.Decision != "" {
			responseData["decision"] = result.Decision
		}
		if result.SystemMessage != "" {
			responseData["systemMessage"] = result.SystemMessage
		}
		if result.Reason != "" {
			responseData["reason"] = result.Reason
		}
		if result.HookSpecificOutput != nil {
			responseData["hookSpecificOutput"] = result.HookSpecificOutput
		}
	} else {
		// Default: continue
		responseData["continue"] = true
	}

	response := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: ControlResponse{
			Subtype:   ResponseSubtypeSuccess,
			RequestID: requestID,
			Response:  responseData,
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal hook response: %w", err)
	}

	return p.transport.Write(ctx, append(data, '\n'))
}

// buildHooksConfig creates the hooks config for the initialize request.
// Format: {"PreToolUse": [{"matcher": "Bash", "hookCallbackIds": ["hook_0"]}], ...}
// This matches the TypeScript SDK's format exactly for CLI compatibility.
func (p *Protocol) buildHooksConfig() map[string][]HookMatcherConfig {
	if p.hooks == nil {
		return nil
	}

	config := make(map[string][]HookMatcherConfig)

	// Initialize callback map if needed
	p.hookCallbacksMu.Lock()
	if p.hookCallbacks == nil {
		p.hookCallbacks = make(map[string]ProtocolHookCallback)
	}

	for event, matchers := range p.hooks {
		eventName := string(event)
		var matcherConfigs []HookMatcherConfig

		for _, matcher := range matchers {
			// Generate callback IDs for each callback in this matcher
			var callbackIDs []string
			for _, callback := range matcher.Hooks {
				callbackID := fmt.Sprintf("hook_%d", p.nextHookCallback)
				p.nextHookCallback++

				// Store callback for later lookup
				p.hookCallbacks[callbackID] = callback
				callbackIDs = append(callbackIDs, callbackID)
			}

			matcherConfigs = append(matcherConfigs, HookMatcherConfig{
				Matcher:         matcher.Matcher,
				HookCallbackIDs: callbackIDs,
				Timeout:         matcher.Timeout,
			})
		}

		if len(matcherConfigs) > 0 {
			config[eventName] = matcherConfigs
		}
	}
	p.hookCallbacksMu.Unlock()

	return config
}

// Helper functions for parsing hook input fields

func getString(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getStringPtr(m map[string]any, key string) *string {
	if v, ok := m[key].(string); ok {
		return &v
	}
	return nil
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getMap(m map[string]any, key string) map[string]any {
	if v, ok := m[key].(map[string]any); ok {
		return v
	}
	return make(map[string]any)
}
