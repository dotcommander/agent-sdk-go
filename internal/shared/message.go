package shared

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strings"
)

// ErrNoMoreMessages is returned by MessageIterator.Next when there are no more messages.
var ErrNoMoreMessages = errors.New("no more messages")

// MessageIterator provides an iterator pattern for streaming messages.
// This is the recommended way to consume messages from a streaming session.
//
// Example:
//
//	iter := client.ReceiveResponse(ctx)
//	defer iter.Close()
//	for {
//	    msg, err := iter.Next(ctx)
//	    if errors.Is(err, shared.ErrNoMoreMessages) {
//	        break
//	    }
//	    if err != nil {
//	        return err
//	    }
//	    // Process message
//	}
type MessageIterator interface {
	// Next returns the next message in the stream.
	// Returns ErrNoMoreMessages when the stream is exhausted.
	// Returns context.Canceled or context.DeadlineExceeded if the context is cancelled.
	Next(ctx context.Context) (Message, error)
	// Close releases any resources held by the iterator.
	Close() error
}

// Message type constants
const (
	MessageTypeUser      = "user"
	MessageTypeAssistant = "assistant"
	MessageTypeSystem    = "system"
	MessageTypeResult    = "result"

	// Control protocol message types
	MessageTypeControlRequest  = "control_request"
	MessageTypeControlResponse = "control_response"

	// Partial message streaming type
	MessageTypeStreamEvent = "stream_event"

	// SDK message types (from TypeScript SDK)
	MessageTypeToolProgress = "tool_progress"
	MessageTypeAuthStatus   = "auth_status"
)

// System message subtype constants
const (
	SystemSubtypeInit           = "init"
	SystemSubtypeCompactBoundary = "compact_boundary"
	SystemSubtypeStatus         = "status"
	SystemSubtypeHookResponse   = "hook_response"
)

// Content block type constants
const (
	ContentBlockTypeText       = "text"
	ContentBlockTypeThinking   = "thinking"
	ContentBlockTypeToolUse    = "tool_use"
	ContentBlockTypeToolResult = "tool_result"
)

// AssistantMessageError represents error types in assistant messages.
type AssistantMessageError string

// AssistantMessageError constants for error type identification.
const (
	AssistantMessageErrorAuthFailed     AssistantMessageError = "authentication_failed"
	AssistantMessageErrorBilling        AssistantMessageError = "billing_error"
	AssistantMessageErrorRateLimit      AssistantMessageError = "rate_limit"
	AssistantMessageErrorInvalidRequest AssistantMessageError = "invalid_request"
	AssistantMessageErrorServer         AssistantMessageError = "server_error"
	AssistantMessageErrorUnknown        AssistantMessageError = "unknown"
)

// Message represents any message type in the Claude Code protocol.
type Message interface {
	Type() string
}

// ContentBlock represents any content block within a message.
type ContentBlock interface {
	BlockType() string
}

// UserMessage represents a message from the user.
type UserMessage struct {
	MessageType     string  `json:"type"`
	Content         any     `json:"content"` // string or []ContentBlock
	UUID            *string `json:"uuid,omitempty"`
	ParentToolUseID *string `json:"parent_tool_use_id,omitempty"`
}

// Type returns the message type for UserMessage.
func (m *UserMessage) Type() string {
	return MessageTypeUser
}

// GetUUID returns the UUID or empty string if nil.
func (m *UserMessage) GetUUID() string {
	if m.UUID != nil {
		return *m.UUID
	}
	return ""
}

// GetParentToolUseID returns the parent tool use ID or empty string if nil.
func (m *UserMessage) GetParentToolUseID() string {
	if m.ParentToolUseID != nil {
		return *m.ParentToolUseID
	}
	return ""
}

// MarshalJSON implements custom JSON marshaling for UserMessage
func (m *UserMessage) MarshalJSON() ([]byte, error) {
	return MarshalWithType(m, MessageTypeUser)
}

// AssistantMessage represents a message from the assistant.
type AssistantMessage struct {
	MessageType string                 `json:"type"`
	Content     []ContentBlock         `json:"content"`
	Model       string                 `json:"model"`
	Error       *AssistantMessageError `json:"error,omitempty"`
}

// Type returns the message type for AssistantMessage.
func (m *AssistantMessage) Type() string {
	return MessageTypeAssistant
}

// HasError returns true if the message contains an error.
func (m *AssistantMessage) HasError() bool {
	return m.Error != nil
}

// GetError returns the error type or empty string if nil.
func (m *AssistantMessage) GetError() AssistantMessageError {
	if m.Error != nil {
		return *m.Error
	}
	return ""
}

// IsRateLimited returns true if the error is a rate limit error.
func (m *AssistantMessage) IsRateLimited() bool {
	return m.Error != nil && *m.Error == AssistantMessageErrorRateLimit
}

// MarshalJSON implements custom JSON marshaling for AssistantMessage
func (m *AssistantMessage) MarshalJSON() ([]byte, error) {
	return MarshalWithType(m, MessageTypeAssistant)
}

// UnmarshalJSON implements custom JSON unmarshaling for AssistantMessage
func (m *AssistantMessage) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary struct with content as raw messages
	type Alias AssistantMessage
	aux := &struct {
		Content []json.RawMessage `json:"content"`
		*Alias
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Parse each content block based on its type
	m.Content = make([]ContentBlock, 0, len(aux.Content))
	for _, rawBlock := range aux.Content {
		// First parse to get the type field
		var typeHolder struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(rawBlock, &typeHolder); err != nil {
			return fmt.Errorf("parse content block type: %w", err)
		}

		// Create the appropriate concrete type based on the type field
		var block ContentBlock
		switch typeHolder.Type {
		case ContentBlockTypeText:
			var tb TextBlock
			if err := json.Unmarshal(rawBlock, &tb); err != nil {
				return fmt.Errorf("parse text block: %w", err)
			}
			block = &tb
		case ContentBlockTypeThinking:
			var tb ThinkingBlock
			if err := json.Unmarshal(rawBlock, &tb); err != nil {
				return fmt.Errorf("parse thinking block: %w", err)
			}
			block = &tb
		case ContentBlockTypeToolUse:
			var tb ToolUseBlock
			if err := json.Unmarshal(rawBlock, &tb); err != nil {
				return fmt.Errorf("parse tool use block: %w", err)
			}
			block = &tb
		case ContentBlockTypeToolResult:
			var tb ToolResultBlock
			if err := json.Unmarshal(rawBlock, &tb); err != nil {
				return fmt.Errorf("parse tool result block: %w", err)
			}
			block = &tb
		default:
			return fmt.Errorf("unknown content block type: %s", typeHolder.Type)
		}

		m.Content = append(m.Content, block)
	}

	return nil
}

// PluginInfo contains information about an active plugin.
type PluginInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// SystemMessage represents a system message.
// When Subtype is "init", the following fields are populated:
// - Agents: list of active agent names
// - Betas: list of active beta features
// - ClaudeCodeVersion: version string of Claude Code
// - Skills: list of active skills
// - Plugins: list of active plugins
type SystemMessage struct {
	MessageType      string         `json:"type"`
	Subtype          string         `json:"subtype"`
	Data             map[string]any `json:"-"` // Preserve all original data
	Agents           []string       `json:"-"` // Populated when subtype is "init"
	Betas            []string       `json:"-"` // Populated when subtype is "init"
	ClaudeCodeVersion string        `json:"-"` // Populated when subtype is "init"
	Skills           []string       `json:"-"` // Populated when subtype is "init"
	Plugins          []PluginInfo   `json:"-"` // Populated when subtype is "init"
}

// Type returns the message type for SystemMessage.
func (m *SystemMessage) Type() string {
	return MessageTypeSystem
}

// MarshalJSON implements custom JSON marshaling for SystemMessage
func (m *SystemMessage) MarshalJSON() ([]byte, error) {
	data := make(map[string]any, len(m.Data)+2)
	maps.Copy(data, m.Data)
	data["type"] = MessageTypeSystem
	data["subtype"] = m.Subtype
	return json.Marshal(data)
}

// ResultMessage represents the final result of a conversation turn.
// The Subtype field indicates the result status:
// - "success": normal completion
// - "error_during_execution": error occurred during execution
// - "error_max_turns": maximum turns reached
// - "error_max_budget_usd": budget limit exceeded
// - "error_max_structured_output_retries": structured output retry limit exceeded
type ResultMessage struct {
	MessageType       string                 `json:"type"`
	Subtype           string                 `json:"subtype"`
	DurationMs        int                    `json:"duration_ms"`
	DurationAPIMs     int                    `json:"duration_api_ms"`
	IsError           bool                   `json:"is_error"`
	NumTurns          int                    `json:"num_turns"`
	SessionID         string                 `json:"session_id"`
	TotalCostUSD      *float64               `json:"total_cost_usd,omitempty"`
	Usage             *map[string]any        `json:"usage,omitempty"`
	Result            *string                `json:"result,omitempty"`
	StructuredOutput  any                    `json:"structured_output,omitempty"`
	ModelUsage        map[string]ModelUsage  `json:"modelUsage,omitempty"`
	PermissionDenials []SDKPermissionDenial  `json:"permission_denials,omitempty"`
	Errors            []string               `json:"errors,omitempty"` // for error subtypes
}

// Type returns the message type for ResultMessage.
func (m *ResultMessage) Type() string {
	return MessageTypeResult
}

// MarshalJSON implements custom JSON marshaling for ResultMessage
func (m *ResultMessage) MarshalJSON() ([]byte, error) {
	return MarshalWithType(m, MessageTypeResult)
}

// TextBlock represents text content.
type TextBlock struct {
	MessageType string `json:"type"`
	Text        string `json:"text"`
}

// BlockType returns the content block type for TextBlock.
func (b *TextBlock) BlockType() string {
	return ContentBlockTypeText
}

// MarshalJSON implements custom JSON marshaling for TextBlock
func (b *TextBlock) MarshalJSON() ([]byte, error) {
	return MarshalWithType(b, ContentBlockTypeText)
}

// ThinkingBlock represents thinking content with signature.
type ThinkingBlock struct {
	MessageType string `json:"type"`
	Thinking    string `json:"thinking"`
	Signature   string `json:"signature"`
}

// BlockType returns the content block type for ThinkingBlock.
func (b *ThinkingBlock) BlockType() string {
	return ContentBlockTypeThinking
}

// MarshalJSON implements custom JSON marshaling for ThinkingBlock
func (b *ThinkingBlock) MarshalJSON() ([]byte, error) {
	return MarshalWithType(b, ContentBlockTypeThinking)
}

// ToolUseBlock represents a tool use request.
type ToolUseBlock struct {
	MessageType string         `json:"type"`
	ToolUseID   string         `json:"tool_use_id"`
	Name        string         `json:"name"`
	Input       map[string]any `json:"input"`
}

// BlockType returns the content block type for ToolUseBlock.
func (b *ToolUseBlock) BlockType() string {
	return ContentBlockTypeToolUse
}

// MarshalJSON implements custom JSON marshaling for ToolUseBlock
func (b *ToolUseBlock) MarshalJSON() ([]byte, error) {
	return MarshalWithType(b, ContentBlockTypeToolUse)
}

// ToolResultBlock represents the result of a tool use.
type ToolResultBlock struct {
	MessageType string `json:"type"`
	ToolUseID   string `json:"tool_use_id"`
	Content     any    `json:"content"` // string or structured data
	IsError     *bool  `json:"is_error,omitempty"`
}

// BlockType returns the content block type for ToolResultBlock.
func (b *ToolResultBlock) BlockType() string {
	return ContentBlockTypeToolResult
}

// MarshalJSON implements custom JSON marshaling for ToolResultBlock
func (b *ToolResultBlock) MarshalJSON() ([]byte, error) {
	return MarshalWithType(b, ContentBlockTypeToolResult)
}

// RawControlMessage wraps raw control protocol messages for passthrough to the control handler.
// Control messages are not parsed into typed structs by the parser - they are routed directly
// to the control protocol handler which performs its own parsing.
type RawControlMessage struct {
	MessageType string
	Data        map[string]any
}

// Type returns the message type for RawControlMessage.
func (m *RawControlMessage) Type() string {
	return m.MessageType
}

// Stream event type constants for Event["type"] discrimination.
// Use these when type-switching on StreamEvent.Event to handle different event types.
const (
	StreamEventTypeContentBlockStart = "content_block_start"
	StreamEventTypeContentBlockDelta = "content_block_delta"
	StreamEventTypeContentBlockStop  = "content_block_stop"
	StreamEventTypeMessageStart      = "message_start"
	StreamEventTypeMessageDelta      = "message_delta"
	StreamEventTypeMessageStop       = "message_stop"
)

// StreamEvent represents a partial message update during streaming.
// Emitted when IncludePartialMessages is enabled in Options.
//
// The Event field contains varying structure depending on event type:
//   - content_block_start: {"type": "content_block_start", "index": <int>, "content_block": {...}}
//   - content_block_delta: {"type": "content_block_delta", "index": <int>, "delta": {...}}
//   - content_block_stop: {"type": "content_block_stop", "index": <int>}
//   - message_start: {"type": "message_start", "message": {...}}
//   - message_delta: {"type": "message_delta", "delta": {...}, "usage": {...}}
//   - message_stop: {"type": "message_stop"}
//
// Consumer code should type-switch on Event["type"] to handle different event types:
//
//	switch event.Event["type"] {
//	case shared.StreamEventTypeContentBlockDelta:
//	    // Handle content delta
//	case shared.StreamEventTypeMessageStop:
//	    // Handle message completion
//	}
type StreamEvent struct {
	UUID            string         `json:"uuid"`
	SessionID       string         `json:"session_id"`
	Event           map[string]any `json:"event"`
	ParentToolUseID *string        `json:"parent_tool_use_id,omitempty"`
}

// Type returns the message type for StreamEvent.
func (m *StreamEvent) Type() string {
	return MessageTypeStreamEvent
}

// MarshalJSON implements custom JSON marshaling for StreamEvent
func (m *StreamEvent) MarshalJSON() ([]byte, error) {
	return MarshalWithType(m, MessageTypeStreamEvent)
}

// CompactMetadata contains metadata about conversation compaction.
type CompactMetadata struct {
	Trigger   string `json:"trigger"` // "manual" | "auto"
	PreTokens int    `json:"pre_tokens"`
}

// CompactBoundaryMessage represents a conversation compaction boundary.
// This is a system message with subtype "compact_boundary".
type CompactBoundaryMessage struct {
	MessageType     string           `json:"type"`    // always "system"
	Subtype         string           `json:"subtype"` // always "compact_boundary"
	CompactMetadata *CompactMetadata `json:"compact_metadata,omitempty"`
	UUID            string           `json:"uuid"`
	SessionID       string           `json:"session_id"`
}

// Type returns the message type for CompactBoundaryMessage.
func (m *CompactBoundaryMessage) Type() string {
	return MessageTypeSystem
}

// MarshalJSON implements custom JSON marshaling for CompactBoundaryMessage
func (m *CompactBoundaryMessage) MarshalJSON() ([]byte, error) {
	return MarshalWithTypeAndSubtype(m, MessageTypeSystem, SystemSubtypeCompactBoundary)
}

// ToolProgressMessage represents tool execution progress.
type ToolProgressMessage struct {
	MessageType        string  `json:"type"` // always "tool_progress"
	ToolUseID          string  `json:"tool_use_id"`
	ToolName           string  `json:"tool_name"`
	ParentToolUseID    *string `json:"parent_tool_use_id,omitempty"`
	ElapsedTimeSeconds float64 `json:"elapsed_time_seconds"`
	UUID               string  `json:"uuid"`
	SessionID          string  `json:"session_id"`
}

// Type returns the message type for ToolProgressMessage.
func (m *ToolProgressMessage) Type() string {
	return MessageTypeToolProgress
}

// MarshalJSON implements custom JSON marshaling for ToolProgressMessage
func (m *ToolProgressMessage) MarshalJSON() ([]byte, error) {
	return MarshalWithType(m, MessageTypeToolProgress)
}

// StatusMessage represents a system status update.
// This is a system message with subtype "status".
type StatusMessage struct {
	MessageType string  `json:"type"`    // always "system"
	Subtype     string  `json:"subtype"` // always "status"
	Status      *string `json:"status"`  // "compacting" | null
	UUID        string  `json:"uuid"`
	SessionID   string  `json:"session_id"`
}

// Type returns the message type for StatusMessage.
func (m *StatusMessage) Type() string {
	return MessageTypeSystem
}

// MarshalJSON implements custom JSON marshaling for StatusMessage
func (m *StatusMessage) MarshalJSON() ([]byte, error) {
	return MarshalWithTypeAndSubtype(m, MessageTypeSystem, SystemSubtypeStatus)
}

// AuthStatusMessage represents authentication status.
type AuthStatusMessage struct {
	MessageType      string   `json:"type"` // always "auth_status"
	IsAuthenticating bool     `json:"isAuthenticating"`
	Output           []string `json:"output"`
	Error            *string  `json:"error,omitempty"`
	UUID             string   `json:"uuid"`
	SessionID        string   `json:"session_id"`
}

// Type returns the message type for AuthStatusMessage.
func (m *AuthStatusMessage) Type() string {
	return MessageTypeAuthStatus
}

// MarshalJSON implements custom JSON marshaling for AuthStatusMessage
func (m *AuthStatusMessage) MarshalJSON() ([]byte, error) {
	return MarshalWithType(m, MessageTypeAuthStatus)
}

// HookResponseMessage represents hook execution response.
// This is a system message with subtype "hook_response".
type HookResponseMessage struct {
	MessageType string  `json:"type"`      // always "system"
	Subtype     string  `json:"subtype"`   // always "hook_response"
	HookName    string  `json:"hook_name"`
	HookEvent   string  `json:"hook_event"`
	Stdout      string  `json:"stdout"`
	Stderr      string  `json:"stderr"`
	ExitCode    *int    `json:"exit_code,omitempty"`
	UUID        string  `json:"uuid"`
	SessionID   string  `json:"session_id"`
}

// Type returns the message type for HookResponseMessage.
func (m *HookResponseMessage) Type() string {
	return MessageTypeSystem
}

// MarshalJSON implements custom JSON marshaling for HookResponseMessage
func (m *HookResponseMessage) MarshalJSON() ([]byte, error) {
	return MarshalWithTypeAndSubtype(m, MessageTypeSystem, SystemSubtypeHookResponse)
}

// ExtractBlocks extracts content blocks of type T from an assistant message.
// Returns nil if the message is not an AssistantMessage or has no matching blocks.
//
// Example:
//
//	toolUses := ExtractBlocks[*ToolUseBlock](msg)
//	textBlocks := ExtractBlocks[*TextBlock](msg)
func ExtractBlocks[T ContentBlock](msg Message) []T {
	assistant, ok := msg.(*AssistantMessage)
	if !ok || assistant == nil {
		return nil
	}
	var result []T
	for _, block := range assistant.Content {
		if typed, ok := block.(T); ok {
			result = append(result, typed)
		}
	}
	return result
}

// ExtractToolUses extracts tool use blocks from an assistant message.
// Returns a slice of ToolUseBlock found in the message content.
func ExtractToolUses(msg Message) []*ToolUseBlock {
	return ExtractBlocks[*ToolUseBlock](msg)
}

// ExtractTextBlocks extracts text blocks from an assistant message.
// Returns a slice of TextBlock found in the message content.
func ExtractTextBlocks(msg Message) []*TextBlock {
	return ExtractBlocks[*TextBlock](msg)
}

// ExtractThinkingBlocks extracts thinking blocks from an assistant message.
// Returns a slice of ThinkingBlock found in the message content.
func ExtractThinkingBlocks(msg Message) []*ThinkingBlock {
	return ExtractBlocks[*ThinkingBlock](msg)
}

// HasToolUses returns true if the message contains any tool use blocks.
func HasToolUses(msg Message) bool {
	return len(ExtractToolUses(msg)) > 0
}

// IsToolUseMessage returns true if the message is an assistant message containing tool uses.
func IsToolUseMessage(msg Message) bool {
	_, ok := msg.(*AssistantMessage)
	return ok && HasToolUses(msg)
}

// GetContentText concatenates all text content from a message.
// Returns empty string if the message is not an assistant message or contains no text blocks.
func GetContentText(msg Message) string {
	assistantMsg, ok := msg.(*AssistantMessage)
	if !ok {
		return ""
	}

	var sb strings.Builder
	for _, block := range assistantMsg.Content {
		if text, ok := block.(*TextBlock); ok {
			sb.WriteString(text.Text)
		}
	}

	return sb.String()
}