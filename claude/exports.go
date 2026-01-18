// Package claude provides the root-level API for the Claude Agent SDK.
// This file exports key types from the shared subpackage for convenient access.
//
// Note: For V2 session-based API (CreateSession, Query, etc.), import the v2 package directly:
//
//	import "github.com/dotcommander/agent-sdk-go/claude/v2"
//
// The v2 package provides the full session management API.
package claude

import (
	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

// SDKError is the base interface for all Claude Agent SDK errors.
// All error types in this package implement this interface.
//
// Example:
//
//	if sdkErr, ok := err.(claude.SDKError); ok {
//	    switch sdkErr.Type() {
//	    case "cli_not_found":
//	        // Handle CLI not found
//	    case "connection":
//	        // Handle connection error
//	    case "timeout":
//	        // Handle timeout
//	    }
//	}
type SDKError = shared.SDKError

// StreamValidator tracks tool requests and results to detect incomplete streams.
// It provides validation for stream integrity and collects statistics about message processing.
//
// Example:
//
//	validator := claude.NewStreamValidator()
//	for msg := range msgChan {
//	    validator.TrackMessage(msg)
//	}
//	validator.MarkStreamEnd()
//	issues := validator.GetIssues()
//	stats := validator.GetStats()
type StreamValidator = shared.StreamValidator

// NewStreamValidator creates a new stream validator.
var NewStreamValidator = shared.NewStreamValidator

// ValidateStreamEvent validates a single stream event.
// Use ValidateStreamEvent to validate individual events.
//
// Example:
//
//	issues := claude.ValidateStreamEvent(event)
//	if len(issues) > 0 {
//	    for _, issue := range issues {
//	        log.Printf("Stream issue: %s - %s", issue.Type, issue.Detail)
//	    }
//	}
var ValidateStreamEvent = shared.ValidateStreamEvent

// ParseStreamEvent parses a raw JSON message into a StreamEvent.
var ParseStreamEvent = shared.ParseStreamEvent

// ExtractDelta extracts the text content from a content_block_delta event.
var ExtractDelta = shared.ExtractDelta

// FormatStats formats StreamStats into a human-readable string.
var FormatStats = shared.FormatStats

// IsCriticalStreamEvent returns true if the event type is critical for response processing.
var IsCriticalStreamEvent = shared.IsCriticalStreamEvent

// IsDeltaStreamEvent returns true if the event type contains delta content.
var IsDeltaStreamEvent = shared.IsDeltaStreamEvent

// StreamEventTypeToString converts a stream event type to a human-readable string.
var StreamEventTypeToString = shared.StreamEventTypeToString

// Re-export error types from shared for convenience.
// These implement the SDKError interface.

// CLINotFoundError indicates the Claude CLI executable could not be found.
type CLINotFoundError = shared.CLINotFoundError

// ConnectionError indicates a failure to connect to or communicate with Claude CLI.
type ConnectionError = shared.ConnectionError

// TimeoutError indicates an operation timed out.
type TimeoutError = shared.TimeoutError

// ParserError indicates a JSON parsing failure.
type ParserError = shared.ParserError

// ProtocolError indicates a protocol violation or invalid message received.
type ProtocolError = shared.ProtocolError

// JSONDecodeError represents JSON parsing failures with detailed position info.
type JSONDecodeError = shared.JSONDecodeError

// MessageParseError represents message structure parsing failures.
type MessageParseError = shared.MessageParseError

// Re-export As*Error helpers for error extraction from wrapped chains.

// AsCLINotFoundError extracts a CLINotFoundError from the error chain.
var AsCLINotFoundError = shared.AsCLINotFoundError

// AsConnectionError extracts a ConnectionError from the error chain.
var AsConnectionError = shared.AsConnectionError

// AsTimeoutError extracts a TimeoutError from the error chain.
var AsTimeoutError = shared.AsTimeoutError

// AsParserError extracts a ParserError from the error chain.
var AsParserError = shared.AsParserError

// AsProtocolError extracts a ProtocolError from the error chain.
var AsProtocolError = shared.AsProtocolError

// AsProcessError extracts a ProcessError from the error chain.
var AsProcessError = shared.AsProcessError

// AsJSONDecodeError extracts a JSONDecodeError from the error chain.
var AsJSONDecodeError = shared.AsJSONDecodeError

// AsMessageParseError extracts a MessageParseError from the error chain.
var AsMessageParseError = shared.AsMessageParseError

// IsJSONDecodeError checks if an error is a JSONDecodeError.
var IsJSONDecodeError = shared.IsJSONDecodeError

// IsMessageParseError checks if an error is a MessageParseError.
var IsMessageParseError = shared.IsMessageParseError

// NewJSONDecodeError creates a new JSONDecodeError.
var NewJSONDecodeError = shared.NewJSONDecodeError

// NewMessageParseError creates a new MessageParseError.
var NewMessageParseError = shared.NewMessageParseError

// Re-export shared types commonly used with hooks and permissions.

// HookEvent represents the type of hook event.
type HookEvent = shared.HookEvent

// HookConfig defines a hook handler configuration.
type HookConfig = shared.HookConfig

// HookHandler is the function signature for hook handlers.
type HookHandler = shared.HookHandler

// SyncHookOutput is the output for synchronous hooks.
type SyncHookOutput = shared.SyncHookOutput

// BaseHookInput contains common fields for all hook inputs.
type BaseHookInput = shared.BaseHookInput

// PreToolUseHookInput is the input for PreToolUse hooks.
type PreToolUseHookInput = shared.PreToolUseHookInput

// PostToolUseHookInput is the input for PostToolUse hooks.
type PostToolUseHookInput = shared.PostToolUseHookInput

// Hook event constants.
const (
	HookEventPreToolUse         = shared.HookEventPreToolUse
	HookEventPostToolUse        = shared.HookEventPostToolUse
	HookEventPostToolUseFailure = shared.HookEventPostToolUseFailure
	HookEventNotification       = shared.HookEventNotification
	HookEventUserPromptSubmit   = shared.HookEventUserPromptSubmit
	HookEventSessionStart       = shared.HookEventSessionStart
	HookEventSessionEnd         = shared.HookEventSessionEnd
	HookEventStop               = shared.HookEventStop
	HookEventSubagentStart      = shared.HookEventSubagentStart
	HookEventSubagentStop       = shared.HookEventSubagentStop
	HookEventPreCompact         = shared.HookEventPreCompact
	HookEventPermissionRequest  = shared.HookEventPermissionRequest
)

// Permission types and constants.

// PermissionMode defines a permission mode.
type PermissionMode = shared.PermissionMode

// CanUseToolCallback is a callback for runtime tool permission checks.
type CanUseToolCallback = shared.CanUseToolCallback

// PermissionResult is the result of a permission check.
type PermissionResult = shared.PermissionResult

// CanUseToolOptions provides context for permission checks.
type CanUseToolOptions = shared.CanUseToolOptions

// PermissionBehavior constants.
const (
	PermissionBehaviorAllow = shared.PermissionBehaviorAllow
	PermissionBehaviorDeny  = shared.PermissionBehaviorDeny
	PermissionBehaviorAsk   = shared.PermissionBehaviorAsk
)

// Agent and MCP types.

// AgentDefinition defines a custom subagent.
type AgentDefinition = shared.AgentDefinition

// McpServerConfig defines MCP server configuration.
type McpServerConfig = shared.McpServerConfig

// OutputFormat defines structured output configuration.
type OutputFormat = shared.OutputFormat

// PluginConfig defines plugin configuration.
type PluginConfig = shared.PluginConfig

// SandboxSettings defines sandbox configuration.
type SandboxSettings = shared.SandboxSettings

// SettingSource defines which settings to load.
type SettingSource = shared.SettingSource

// ToolsConfig defines tools configuration (preset or explicit list).
type ToolsConfig = shared.ToolsConfig

// Tool preset helpers.

// ToolsPreset creates a ToolsConfig from a preset name.
var ToolsPreset = shared.ToolsPreset

// ToolsExplicit creates a ToolsConfig from an explicit list of tools.
var ToolsExplicit = shared.ToolsExplicit

// =============================================================================
// Permission Result Constructors (P1)
// =============================================================================

// NewPermissionResultAllow creates a permission result that allows tool execution.
var NewPermissionResultAllow = shared.NewPermissionResultAllow

// NewPermissionResultDeny creates a permission result that denies tool execution.
var NewPermissionResultDeny = shared.NewPermissionResultDeny

// WithUpdatedInput modifies the tool input before execution.
var WithUpdatedInput = shared.WithUpdatedInput

// WithPermissionUpdates suggests permission rule changes.
var WithPermissionUpdates = shared.WithPermissionUpdates

// WithInterrupt terminates the session on denial.
var WithInterrupt = shared.WithInterrupt

// WithToolUseID sets the tool use ID on the result.
var WithToolUseID = shared.WithToolUseID

// PermissionResultOption configures a PermissionResult.
type PermissionResultOption = shared.PermissionResultOption

// PermissionUpdate represents an update to permission configuration.
type PermissionUpdate = shared.PermissionUpdate

// PermissionRuleValue represents a permission rule.
type PermissionRuleValue = shared.PermissionRuleValue

// PermissionUpdateDestination specifies where permission updates are stored.
type PermissionUpdateDestination = shared.PermissionUpdateDestination

// Permission mode constants.
const (
	PermissionModeDefault           = shared.PermissionModeDefault
	PermissionModeAcceptEdits       = shared.PermissionModeAcceptEdits
	PermissionModeBypassPermissions = shared.PermissionModeBypassPermissions
	PermissionModePlan              = shared.PermissionModePlan
	PermissionModeDelegate          = shared.PermissionModeDelegate
	PermissionModeDontAsk           = shared.PermissionModeDontAsk
)

// Permission update destination constants.
const (
	PermissionDestUserSettings    = shared.PermissionDestUserSettings
	PermissionDestProjectSettings = shared.PermissionDestProjectSettings
	PermissionDestLocalSettings   = shared.PermissionDestLocalSettings
	PermissionDestSession         = shared.PermissionDestSession
	PermissionDestCLIArg          = shared.PermissionDestCLIArg
)

// =============================================================================
// Additional Hook Types
// =============================================================================

// PostToolUseFailureHookInput is the input for PostToolUseFailure hooks.
type PostToolUseFailureHookInput = shared.PostToolUseFailureHookInput

// NotificationHookInput is the input for Notification hooks.
type NotificationHookInput = shared.NotificationHookInput

// UserPromptSubmitHookInput is the input for UserPromptSubmit hooks.
type UserPromptSubmitHookInput = shared.UserPromptSubmitHookInput

// SessionStartHookInput is the input for SessionStart hooks.
type SessionStartHookInput = shared.SessionStartHookInput

// SessionEndHookInput is the input for SessionEnd hooks.
type SessionEndHookInput = shared.SessionEndHookInput

// StopHookInput is the input for Stop hooks.
type StopHookInput = shared.StopHookInput

// SubagentStartHookInput is the input for SubagentStart hooks.
type SubagentStartHookInput = shared.SubagentStartHookInput

// SubagentStopHookInput is the input for SubagentStop hooks.
type SubagentStopHookInput = shared.SubagentStopHookInput

// PreCompactHookInput is the input for PreCompact hooks.
type PreCompactHookInput = shared.PreCompactHookInput

// PermissionRequestHookInput is the input for PermissionRequest hooks.
type PermissionRequestHookInput = shared.PermissionRequestHookInput

// =============================================================================
// Additional Message Types
// =============================================================================

// CompactBoundaryMessage represents a conversation compaction boundary.
type CompactBoundaryMessage = shared.CompactBoundaryMessage

// ToolProgressMessage represents tool execution progress.
type ToolProgressMessage = shared.ToolProgressMessage

// StatusMessage represents a system status update.
type StatusMessage = shared.StatusMessage

// AuthStatusMessage represents authentication status.
type AuthStatusMessage = shared.AuthStatusMessage

// HookResponseMessage represents hook execution response.
type HookResponseMessage = shared.HookResponseMessage

// RawControlMessage wraps raw control protocol messages.
type RawControlMessage = shared.RawControlMessage

// =============================================================================
// Session Options (TypeScript SDK Parity)
// =============================================================================

// BaseOptions contains common options shared across different option types.
type BaseOptions = shared.BaseOptions

// OutputFormat represents structured output configuration.
// type OutputFormat = shared.OutputFormat (already exported above)

// =============================================================================
// Utility Functions
// =============================================================================

// AllHookEvents returns all valid hook event types.
var AllHookEvents = shared.AllHookEvents

// ExtractToolUses extracts tool use blocks from an assistant message.
var ExtractToolUses = shared.ExtractToolUses

// ExtractTextBlocks extracts text blocks from an assistant message.
var ExtractTextBlocks = shared.ExtractTextBlocks

// ExtractThinkingBlocks extracts thinking blocks from an assistant message.
var ExtractThinkingBlocks = shared.ExtractThinkingBlocks

// HasToolUses returns true if the message contains any tool use blocks.
var HasToolUses = shared.HasToolUses

// IsToolUseMessage returns true if the message is an assistant message containing tool uses.
var IsToolUseMessage = shared.IsToolUseMessage

// GetContentText concatenates all text content from a message.
var GetContentText = shared.GetContentText

// =============================================================================
// Model Usage and Result Types
// =============================================================================

// ModelUsage represents usage statistics for a specific model.
type ModelUsage = shared.ModelUsage

// SDKPermissionDenial represents a permission denial that occurred during execution.
type SDKPermissionDenial = shared.SDKPermissionDenial

// McpServerStatus represents the status of an MCP server.
type McpServerStatus = shared.McpServerStatus

// McpSetServersResult represents the result of setting MCP servers.
type McpSetServersResult = shared.McpSetServersResult

// =============================================================================
// Control Protocol Types (for advanced usage)
// =============================================================================

// Control protocol message type constants.
const (
	// ControlMessageTypeRequest is sent TO the CLI to request an action.
	ControlMessageTypeRequest = "control_request"
	// ControlMessageTypeResponse is received FROM the CLI as a response.
	ControlMessageTypeResponse = "control_response"
)

// Request subtype constants matching TypeScript SDK for 100% parity.
const (
	// ControlSubtypeInterrupt requests interruption of current operation.
	ControlSubtypeInterrupt = "interrupt"
	// ControlSubtypeCanUseTool requests permission to use a tool.
	ControlSubtypeCanUseTool = "can_use_tool"
	// ControlSubtypeInitialize performs the control protocol handshake.
	ControlSubtypeInitialize = "initialize"
	// ControlSubtypeSetPermissionMode changes the permission mode at runtime.
	ControlSubtypeSetPermissionMode = "set_permission_mode"
	// ControlSubtypeSetModel changes the AI model at runtime.
	ControlSubtypeSetModel = "set_model"
	// ControlSubtypeHookCallback invokes a registered hook callback.
	ControlSubtypeHookCallback = "hook_callback"
	// ControlSubtypeMcpMessage routes an MCP message to an SDK MCP server.
	ControlSubtypeMcpMessage = "mcp_message"
	// ControlSubtypeRewindFiles requests file rewind to a specific user message state.
	ControlSubtypeRewindFiles = "rewind_files"
)

// Response subtype constants for control responses.
const (
	// ControlResponseSubtypeSuccess indicates the request succeeded.
	ControlResponseSubtypeSuccess = "success"
	// ControlResponseSubtypeError indicates the request failed.
	ControlResponseSubtypeError = "error"
)
