package shared

// HookEvent represents the type of hook event.
type HookEvent string

const (
	HookEventPreToolUse         HookEvent = "PreToolUse"
	HookEventPostToolUse        HookEvent = "PostToolUse"
	HookEventPostToolUseFailure HookEvent = "PostToolUseFailure"
	HookEventNotification       HookEvent = "Notification"
	HookEventUserPromptSubmit   HookEvent = "UserPromptSubmit"
	HookEventSessionStart       HookEvent = "SessionStart"
	HookEventSessionEnd         HookEvent = "SessionEnd"
	HookEventStop               HookEvent = "Stop"
	HookEventSubagentStart      HookEvent = "SubagentStart"
	HookEventSubagentStop       HookEvent = "SubagentStop"
	HookEventPreCompact         HookEvent = "PreCompact"
	HookEventPermissionRequest  HookEvent = "PermissionRequest"
)

// AllHookEvents returns all valid hook event types.
func AllHookEvents() []HookEvent {
	return []HookEvent{
		HookEventPreToolUse, HookEventPostToolUse, HookEventPostToolUseFailure,
		HookEventNotification, HookEventUserPromptSubmit, HookEventSessionStart,
		HookEventSessionEnd, HookEventStop, HookEventSubagentStart,
		HookEventSubagentStop, HookEventPreCompact, HookEventPermissionRequest,
	}
}

// BaseHookInput contains common fields for all hook inputs.
type BaseHookInput struct {
	SessionID      string `json:"session_id"`
	TranscriptPath string `json:"transcript_path"`
	Cwd            string `json:"cwd"`
	PermissionMode string `json:"permission_mode,omitempty"`
}

// PreToolUseHookInput is the input for PreToolUse hooks.
type PreToolUseHookInput struct {
	BaseHookInput
	HookEventName string         `json:"hook_event_name"` // always "PreToolUse"
	ToolName      string         `json:"tool_name"`
	ToolInput     map[string]any `json:"tool_input"`
	ToolUseID     string         `json:"tool_use_id"`
}

// PostToolUseHookInput is the input for PostToolUse hooks.
type PostToolUseHookInput struct {
	BaseHookInput
	HookEventName string         `json:"hook_event_name"` // always "PostToolUse"
	ToolName      string         `json:"tool_name"`
	ToolInput     map[string]any `json:"tool_input"`
	ToolResponse  any            `json:"tool_response"`
	ToolUseID     string         `json:"tool_use_id"`
}

// PostToolUseFailureHookInput is the input for PostToolUseFailure hooks.
type PostToolUseFailureHookInput struct {
	BaseHookInput
	HookEventName string         `json:"hook_event_name"` // always "PostToolUseFailure"
	ToolName      string         `json:"tool_name"`
	ToolInput     map[string]any `json:"tool_input"`
	ToolUseID     string         `json:"tool_use_id"`
	Error         string         `json:"error"`
	IsInterrupt   bool           `json:"is_interrupt,omitempty"`
}

// NotificationHookInput is the input for Notification hooks.
type NotificationHookInput struct {
	BaseHookInput
	HookEventName    string `json:"hook_event_name"` // always "Notification"
	Message          string `json:"message"`
	Title            string `json:"title,omitempty"`
	NotificationType string `json:"notification_type"`
}

// UserPromptSubmitHookInput is the input for UserPromptSubmit hooks.
type UserPromptSubmitHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // always "UserPromptSubmit"
	Prompt        string `json:"prompt"`
}

// SessionStartHookInput is the input for SessionStart hooks.
type SessionStartHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // always "SessionStart"
	Source        string `json:"source"`          // "startup" | "resume" | "clear" | "compact"
	AgentType     string `json:"agent_type,omitempty"`
	Model         string `json:"model,omitempty"`
}

// SessionEndHookInput is the input for SessionEnd hooks.
type SessionEndHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // always "SessionEnd"
	Reason        string `json:"reason"`
}

// StopHookInput is the input for Stop hooks.
type StopHookInput struct {
	BaseHookInput
	HookEventName  string `json:"hook_event_name"` // always "Stop"
	StopHookActive bool   `json:"stop_hook_active"`
}

// SubagentStartHookInput is the input for SubagentStart hooks.
type SubagentStartHookInput struct {
	BaseHookInput
	HookEventName string `json:"hook_event_name"` // always "SubagentStart"
	AgentID       string `json:"agent_id"`
	AgentType     string `json:"agent_type"`
}

// SubagentStopHookInput is the input for SubagentStop hooks.
type SubagentStopHookInput struct {
	BaseHookInput
	HookEventName       string `json:"hook_event_name"` // always "SubagentStop"
	StopHookActive      bool   `json:"stop_hook_active"`
	AgentID             string `json:"agent_id"`
	AgentTranscriptPath string `json:"agent_transcript_path"`
}

// PreCompactHookInput is the input for PreCompact hooks.
type PreCompactHookInput struct {
	BaseHookInput
	HookEventName      string  `json:"hook_event_name"` // always "PreCompact"
	Trigger            string  `json:"trigger"`         // "manual" | "auto"
	CustomInstructions *string `json:"custom_instructions"`
}

// PermissionRequestHookInput is the input for PermissionRequest hooks.
type PermissionRequestHookInput struct {
	BaseHookInput
	HookEventName         string             `json:"hook_event_name"` // always "PermissionRequest"
	ToolName              string             `json:"tool_name"`
	ToolInput             any                `json:"tool_input"`
	PermissionSuggestions []PermissionUpdate `json:"permission_suggestions,omitempty"`
}

// AsyncHookOutput indicates an async hook response.
type AsyncHookOutput struct {
	Async        bool `json:"async"`
	AsyncTimeout int  `json:"asyncTimeout,omitempty"` // seconds
}

// SyncHookOutput is the output for synchronous hooks.
type SyncHookOutput struct {
	Continue           bool           `json:"continue,omitempty"`
	SuppressOutput     bool           `json:"suppressOutput,omitempty"`
	StopReason         string         `json:"stopReason,omitempty"`
	Decision           string         `json:"decision,omitempty"` // "approve" | "block"
	SystemMessage      string         `json:"systemMessage,omitempty"`
	Reason             string         `json:"reason,omitempty"`
	HookSpecificOutput map[string]any `json:"hookSpecificOutput,omitempty"`
}
