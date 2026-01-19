// Package subprocess provides subprocess communication with the Claude CLI.
// This file handles permission callback processing for can_use_tool requests.
package subprocess

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

// handleCanUseToolRequest processes a permission check request from CLI.
// Follows StderrCallback pattern: synchronous with panic recovery.
func (p *Protocol) handleCanUseToolRequest(ctx context.Context, requestID string, request map[string]any) error {
	// Parse request fields
	toolName, _ := request["tool_name"].(string)
	if toolName == "" {
		return p.sendErrorResponse(ctx, requestID, "missing tool_name")
	}

	input, _ := request["input"].(map[string]any)
	if input == nil {
		input = make(map[string]any)
	}

	// Parse optional fields for CanUseToolOptions
	opts := shared.CanUseToolOptions{}
	if suggestions, ok := request["permission_suggestions"].([]any); ok {
		opts.Suggestions = parsePermissionSuggestions(suggestions)
	}
	if blockedPath, ok := request["blocked_path"].(string); ok {
		opts.BlockedPath = blockedPath
	}
	if reason, ok := request["decision_reason"].(string); ok {
		opts.DecisionReason = reason
	}
	if toolUseID, ok := request["tool_use_id"].(string); ok {
		opts.ToolUseID = toolUseID
	}
	if agentID, ok := request["agent_id"].(string); ok {
		opts.AgentID = agentID
	}

	// Get callback (thread-safe read)
	p.mu.Lock()
	callback := p.canUseToolCallback
	p.mu.Unlock()

	// No callback = deny (secure default)
	if callback == nil {
		return p.sendPermissionResponse(ctx, requestID, shared.PermissionResult{
			Behavior: shared.PermissionBehaviorDeny,
			Message:  "no permission callback registered",
		})
	}

	// Invoke callback synchronously with panic recovery (matches StderrCallback pattern)
	var result shared.PermissionResult
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("permission callback panicked: %v", r)
			}
		}()
		result, err = callback(ctx, toolName, input, opts)
	}()

	if err != nil {
		return p.sendErrorResponse(ctx, requestID, fmt.Sprintf("callback error: %v", err))
	}

	return p.sendPermissionResponse(ctx, requestID, result)
}

// sendPermissionResponse sends a permission result back to CLI.
func (p *Protocol) sendPermissionResponse(ctx context.Context, requestID string, result shared.PermissionResult) error {
	// Build response based on result behavior
	responseData := map[string]any{"behavior": string(result.Behavior)}

	if result.Behavior == shared.PermissionBehaviorAllow {
		if result.UpdatedInput != nil {
			responseData["updatedInput"] = result.UpdatedInput
		}
		if len(result.UpdatedPermissions) > 0 {
			responseData["updatedPermissions"] = result.UpdatedPermissions
		}
	} else if result.Behavior == shared.PermissionBehaviorDeny {
		if result.Message != "" {
			responseData["message"] = result.Message
		}
		if result.Interrupt {
			responseData["interrupt"] = result.Interrupt
		}
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
		return fmt.Errorf("marshal permission response: %w", err)
	}

	return p.transport.Write(ctx, append(data, '\n'))
}

// parsePermissionSuggestions converts raw JSON to PermissionUpdate slice.
// Invalid or unrecognized items are silently skipped for forward compatibility
// with future CLI versions that may introduce new fields or formats.
func parsePermissionSuggestions(raw []any) []shared.PermissionUpdate {
	var suggestions []shared.PermissionUpdate
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		update := shared.PermissionUpdate{}

		if t, ok := m["type"].(string); ok {
			update.Type = t
		}

		if rules, ok := m["rules"].([]any); ok {
			for _, rule := range rules {
				ruleMap, ok := rule.(map[string]any)
				if !ok {
					continue
				}
				rv := shared.PermissionRuleValue{}
				if tn, ok := ruleMap["toolName"].(string); ok {
					rv.ToolName = tn
				}
				if rc, ok := ruleMap["ruleContent"].(string); ok {
					rv.RuleContent = &rc
				}
				update.Rules = append(update.Rules, rv)
			}
		}

		if b, ok := m["behavior"].(string); ok {
			update.Behavior = shared.PermissionBehavior(b)
		}
		if mode, ok := m["mode"].(string); ok {
			update.Mode = shared.PermissionMode(mode)
		}
		if dirs, ok := m["directories"].([]any); ok {
			for _, d := range dirs {
				if ds, ok := d.(string); ok {
					update.Directories = append(update.Directories, ds)
				}
			}
		}
		if dest, ok := m["destination"].(string); ok {
			update.Destination = shared.PermissionUpdateDestination(dest)
		}

		suggestions = append(suggestions, update)
	}
	return suggestions
}
