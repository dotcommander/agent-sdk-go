// Package subprocess provides subprocess communication with the Claude CLI.
// This file contains control protocol message types and structures.
package subprocess

// Control protocol message type constants.
const (
	// MessageTypeControlRequest is sent TO the CLI to request an action.
	MessageTypeControlRequest = "control_request"
	// MessageTypeControlResponse is received FROM the CLI as a response.
	MessageTypeControlResponse = "control_response"
)

// Request subtype constants matching TypeScript SDK for 100% parity.
const (
	// SubtypeInterrupt requests interruption of current operation.
	SubtypeInterrupt = "interrupt"
	// SubtypeCanUseTool requests permission to use a tool.
	SubtypeCanUseTool = "can_use_tool"
	// SubtypeInitialize performs the control protocol handshake.
	SubtypeInitialize = "initialize"
	// SubtypeSetPermissionMode changes the permission mode at runtime.
	SubtypeSetPermissionMode = "set_permission_mode"
	// SubtypeSetModel changes the AI model at runtime.
	SubtypeSetModel = "set_model"
	// SubtypeHookCallback invokes a registered hook callback.
	SubtypeHookCallback = "hook_callback"
	// SubtypeMcpMessage routes an MCP message to an SDK MCP server.
	SubtypeMcpMessage = "mcp_message"
	// SubtypeRewindFiles requests file rewind to a specific user message state.
	SubtypeRewindFiles = "rewind_files"
	// SubtypeGetAccountInfo requests the current account information.
	SubtypeGetAccountInfo = "get_account_info"
	// SubtypeGetModels requests the list of available models.
	SubtypeGetModels = "get_models"
	// SubtypeGetCommands requests the list of available slash commands.
	SubtypeGetCommands = "get_commands"
	// SubtypeGetMcpServerStatus requests the status of MCP servers.
	SubtypeGetMcpServerStatus = "get_mcp_server_status"
	// SubtypeSetMcpServers sets the MCP server configuration.
	SubtypeSetMcpServers = "set_mcp_servers"
)

// Response subtype constants for control responses.
const (
	// ResponseSubtypeSuccess indicates the request succeeded.
	ResponseSubtypeSuccess = "success"
	// ResponseSubtypeError indicates the request failed.
	ResponseSubtypeError = "error"
)

// SDKControlRequest represents a control request sent TO the CLI.
// This is the envelope that wraps all control request types.
type SDKControlRequest struct {
	// Type is always MessageTypeControlRequest.
	Type string `json:"type"`
	// RequestID is a unique identifier for request/response correlation.
	// Format: req_{counter}_{random_hex}
	RequestID string `json:"request_id"`
	// Request contains the actual request payload (InterruptRequest, InitializeRequest, etc.).
	Request any `json:"request"`
}

// SDKControlResponse represents a control response received FROM the CLI.
// This is the envelope that wraps all control response types.
type SDKControlResponse struct {
	// Type is always MessageTypeControlResponse.
	Type string `json:"type"`
	// Response contains the actual response data.
	Response ControlResponse `json:"response"`
}

// ControlResponse is the inner response structure within SDKControlResponse.
type ControlResponse struct {
	// Subtype is either ResponseSubtypeSuccess or ResponseSubtypeError.
	Subtype string `json:"subtype"`
	// RequestID matches the request that this response is for.
	RequestID string `json:"request_id"`
	// Response contains the response data (only for success).
	Response any `json:"response,omitempty"`
	// Error contains the error message (only for error).
	Error string `json:"error,omitempty"`
}

// InterruptRequest requests interruption of the current operation.
type InterruptRequest struct {
	// Subtype is always SubtypeInterrupt.
	Subtype string `json:"subtype"`
}

// InitializeRequest performs the control protocol handshake.
// This must be sent before any other control requests in streaming mode.
type InitializeRequest struct {
	// Subtype is always SubtypeInitialize.
	Subtype string `json:"subtype"`
	// Hooks contains hook registrations keyed by event type.
	// Format: {"PreToolUse": [...], "PostToolUse": [...]}
	Hooks map[string][]HookMatcherConfig `json:"hooks,omitempty"`
}

// InitializeResponse contains the CLI's response to initialization.
type InitializeResponse struct {
	// SupportedCommands lists the control commands supported by this CLI version.
	SupportedCommands []string `json:"supported_commands,omitempty"`
}

// SetPermissionModeRequest changes the permission mode at runtime.
type SetPermissionModeRequest struct {
	// Subtype is always SubtypeSetPermissionMode.
	Subtype string `json:"subtype"`
	// Mode is the new permission mode to set.
	Mode string `json:"mode"`
}

// SetModelRequest changes the AI model at runtime.
// This matches TypeScript SDK's set_model() behavior exactly.
type SetModelRequest struct {
	// Subtype is always SubtypeSetModel.
	Subtype string `json:"subtype"`
	// Model is the new model to use. Use nil to reset to default.
	// Examples: "claude-sonnet-4-5", "claude-opus-4-1-20250805"
	Model *string `json:"model"`
}

// RewindFilesRequest requests rewinding files to a specific user message state.
// Matches TypeScript SDK's SDKControlRewindFilesRequest structure.
type RewindFilesRequest struct {
	// Subtype is always SubtypeRewindFiles ("rewind_files").
	Subtype string `json:"subtype"`
	// UserMessageID is the UUID of the user message to rewind to.
	// This should be obtained from UserMessage.UUID received during the session.
	UserMessageID string `json:"user_message_id"`
}

// HookMatcherConfig is the serializable format for the initialize request.
// This is what gets sent to the CLI during initialization.
type HookMatcherConfig struct {
	// Matcher is a tool name pattern.
	Matcher string `json:"matcher"`
	// HookCallbackIDs are the generated callback IDs for this matcher.
	HookCallbackIDs []string `json:"hookCallbackIds"`
	// Timeout is the maximum time in seconds.
	Timeout *float64 `json:"timeout,omitempty"`
}

// HookRegistration represents a hook registration for initialization.
type HookRegistration struct {
	// CallbackID is the unique identifier for this callback.
	CallbackID string `json:"callback_id"`
	// Matcher is the tool name pattern.
	Matcher string `json:"matcher"`
	// Timeout is the maximum time in seconds.
	Timeout *float64 `json:"timeout,omitempty"`
}

// CanUseToolRequest is a request from CLI to check tool permission.
// This is received as an incoming control request.
type CanUseToolRequest struct {
	// Subtype is always SubtypeCanUseTool.
	Subtype string `json:"subtype"`
	// ToolName is the name of the tool being requested.
	ToolName string `json:"tool_name"`
	// Input contains the tool's input parameters.
	Input map[string]any `json:"input"`
	// PermissionSuggestions from CLI (optional).
	PermissionSuggestions []any `json:"permission_suggestions,omitempty"`
}

// HookCallbackRequest is a request from CLI to invoke a hook callback.
type HookCallbackRequest struct {
	// Subtype is always SubtypeHookCallback.
	Subtype string `json:"subtype"`
	// CallbackID identifies which hook to invoke.
	CallbackID string `json:"callback_id"`
	// Input contains the hook input data.
	Input map[string]any `json:"input"`
	// ToolUseID is the tool use identifier (optional, for tool-related hooks).
	ToolUseID *string `json:"tool_use_id,omitempty"`
}

// McpMessageRequest is a request from CLI to route an MCP message.
type McpMessageRequest struct {
	// Subtype is always SubtypeMcpMessage.
	Subtype string `json:"subtype"`
	// ServerName identifies the target SDK MCP server.
	ServerName string `json:"server_name"`
	// Message is the JSONRPC message to route.
	Message map[string]any `json:"message"`
}

// GetAccountInfoRequest requests account information from the CLI.
type GetAccountInfoRequest struct {
	// Subtype is always SubtypeGetAccountInfo.
	Subtype string `json:"subtype"`
}

// GetModelsRequest requests the list of available models from the CLI.
type GetModelsRequest struct {
	// Subtype is always SubtypeGetModels.
	Subtype string `json:"subtype"`
}

// GetCommandsRequest requests the list of available slash commands from the CLI.
type GetCommandsRequest struct {
	// Subtype is always SubtypeGetCommands.
	Subtype string `json:"subtype"`
}

// GetMcpServerStatusRequest requests the status of MCP servers from the CLI.
type GetMcpServerStatusRequest struct {
	// Subtype is always SubtypeGetMcpServerStatus.
	Subtype string `json:"subtype"`
}

// SetMcpServersRequest sets the MCP server configuration.
type SetMcpServersRequest struct {
	// Subtype is always SubtypeSetMcpServers.
	Subtype string `json:"subtype"`
	// Servers is the map of server name to configuration.
	Servers map[string]any `json:"servers"`
}
