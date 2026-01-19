package v2

import (
	"context"
	"strings"
	"time"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

// V2Message represents any message in the V2 protocol.
// It's compatible with the existing shared.Message interface.
type V2Message interface {
	Type() string
}

// V2Session represents a V2 session with send/receive pattern.
// This mirrors the TypeScript SDK V2 Session interface.
type V2Session interface {
	// Send sends a message to Claude in this session.
	Send(ctx context.Context, message string) error

	// Receive returns a channel of V2Message responses.
	// The channel closes when the response is complete.
	Receive(ctx context.Context) <-chan V2Message

	// ReceiveIterator returns an iterator for receiving messages.
	// This provides an alternative to the channel-based Receive().
	ReceiveIterator(ctx context.Context) V2MessageIterator

	// Close closes the session and releases resources.
	Close() error

	// SessionID returns the unique session identifier.
	SessionID() string

	// Interrupt stops the current query execution.
	Interrupt(ctx context.Context) error

	// SetPermissionMode changes the permission mode.
	SetPermissionMode(ctx context.Context, mode shared.PermissionMode) error

	// SetModel changes the model used.
	SetModel(ctx context.Context, model string) error

	// SetMaxThinkingTokens adjusts the thinking token limit.
	// Pass nil to clear the limit.
	SetMaxThinkingTokens(ctx context.Context, tokens *int) error

	// SupportedCommands returns available slash commands.
	SupportedCommands(ctx context.Context) ([]shared.SlashCommand, error)

	// SupportedModels returns available models.
	SupportedModels(ctx context.Context) ([]shared.ModelInfo, error)

	// McpServerStatus returns MCP server statuses.
	McpServerStatus(ctx context.Context) ([]shared.McpServerStatus, error)

	// AccountInfo returns account information.
	AccountInfo(ctx context.Context) (*shared.AccountInfo, error)

	// RewindFiles rewinds files to a specific message state.
	RewindFiles(ctx context.Context, userMessageID string, opts *RewindFilesOptions) (*shared.RewindFilesResult, error)

	// SetMcpServers dynamically sets MCP servers.
	SetMcpServers(ctx context.Context, servers map[string]shared.McpServerConfig) (*shared.McpSetServersResult, error)
}

// RewindFilesOptions contains options for RewindFiles operation.
type RewindFilesOptions struct {
	DryRun bool `json:"dryRun,omitempty"`
}

// V2MessageIterator provides an iterator pattern for receiving messages.
// This is an alternative to the channel-based Receive() method.
type V2MessageIterator interface {
	// Next advances to the next message.
	// Returns ErrNoMoreMessages when iteration is complete.
	Next(ctx context.Context) (V2Message, error)

	// Close closes the iterator and releases resources.
	Close() error
}

// V2Result represents the result of a one-shot prompt operation.
// This mirrors the return value of unstable_v2_prompt() in TypeScript.
type V2Result struct {
	// Result contains the text response from Claude.
	Result string `json:"result"`

	// SessionID contains the unique session identifier.
	SessionID string `json:"session_id"`

	// StartTime is when the prompt was started.
	StartTime time.Time `json:"start_time"`

	// EndTime is when the prompt completed.
	EndTime time.Time `json:"end_time"`

	// Duration is the total time taken.
	Duration time.Duration `json:"duration"`
}

// V2SessionOptions contains configuration for creating V2 sessions.
// This mirrors the options parameter in unstable_v2_createSession().
// Embeds shared.BaseOptions for common fields (DRY).
type V2SessionOptions struct {
	shared.BaseOptions

	// Timeout specifies how long to wait for responses.
	Timeout time.Duration `json:"timeout"`

	// EnablePartialMessages enables streaming of partial messages.
	EnablePartialMessages bool `json:"enable_partial_messages,omitempty"`

	// Hooks contains registered hook handlers for session lifecycle events.
	// Hooks are executed in-process when the corresponding events occur.
	Hooks []shared.HookConfig `json:"hooks,omitempty"`

	// clientFactory provides a factory for creating clients (DIP compliance).
	// If nil, the default factory is used. This is unexported but settable via WithClientFactory.
	clientFactory ClientFactory

	// cliChecker provides a CLI availability checker (DIP compliance for testability).
	// If nil, the default cli.IsCLIAvailable is used.
	cliChecker shared.CLIChecker
}

// V2AssistantMessage represents an assistant response in V2 format.
type V2AssistantMessage struct {
	// TypeField is always "assistant" for assistant messages.
	TypeField string `json:"type"`

	// Message contains the assistant message content.
	Message AssistantMessageContent `json:"message"`

	// SessionID is the unique session identifier.
	SessionID string `json:"session_id"`
}

// Type returns the message type for V2AssistantMessage.
func (m *V2AssistantMessage) Type() string {
	return m.TypeField
}

// AssistantMessageContent represents the content of an assistant message.
type AssistantMessageContent struct {
	// Content is an array of content blocks.
	Content []shared.ContentBlock `json:"content"`

	// Model is the model that generated the response.
	Model string `json:"model"`
}

// V2ResultMessage represents the final result message in V2 format.
type V2ResultMessage struct {
	// TypeField is always "result" for result messages.
	TypeField string `json:"type"`

	// Result contains the final result text.
	Result string `json:"result"`

	// SessionID is the unique session identifier.
	SessionID string `json:"session_id"`

	// Usage contains token usage information.
	Usage *V2Usage `json:"usage,omitempty"`
}

// Type returns the message type for V2ResultMessage.
func (m *V2ResultMessage) Type() string {
	return m.TypeField
}

// V2Usage represents token usage information.
type V2Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// V2StreamDelta represents a streaming delta update.
type V2StreamDelta struct {
	// Type is the delta type (e.g., "text_delta").
	TypeField string `json:"type"`

	// Delta contains the delta content.
	Delta map[string]any `json:"delta"`

	// SessionID is the unique session identifier.
	SessionID string `json:"session_id"`
}

// Type returns the message type for V2StreamDelta.
func (m *V2StreamDelta) Type() string {
	return m.TypeField
}

// V2Error represents an error in the V2 protocol.
type V2Error struct {
	// Type is always "error" for error messages.
	TypeField string `json:"type"`

	// Error contains the error message.
	ErrorField string `json:"error"`

	// SessionID is the unique session identifier.
	SessionID string `json:"session_id,omitempty"`
}

// Type returns the message type for V2Error.
func (m *V2Error) Type() string {
	return m.TypeField
}

// ErrNoMoreMessages indicates the V2MessageIterator has no more messages.
var ErrNoMoreMessages = shared.NewProtocolError("no_more_messages", "iterator has no more messages")

// V2EventType constants for message type discrimination.
const (
	V2EventTypeAssistant   = "assistant"
	V2EventTypeResult      = "result"
	V2EventTypeStreamDelta = "stream_delta"
	V2EventTypeError       = "error"
)

// Helper functions for working with V2 messages.

// IsAssistantMessage checks if a message is an assistant response.
func IsAssistantMessage(msg V2Message) bool {
	return msg.Type() == V2EventTypeAssistant
}

// IsResultMessage checks if a message is a final result.
func IsResultMessage(msg V2Message) bool {
	return msg.Type() == V2EventTypeResult
}

// IsStreamDelta checks if a message is a streaming delta.
func IsStreamDelta(msg V2Message) bool {
	return msg.Type() == V2EventTypeStreamDelta
}

// IsErrorMessage checks if a message is an error.
func IsErrorMessage(msg V2Message) bool {
	return msg.Type() == V2EventTypeError
}

// ExtractAssistantText extracts text content from an assistant message.
// Returns empty string if the message is not an assistant message.
func ExtractAssistantText(msg V2Message) string {
	if !IsAssistantMessage(msg) {
		return ""
	}

	// Handle V2AssistantMessage
	if v2Msg, ok := msg.(*V2AssistantMessage); ok {
		var result strings.Builder
		for _, block := range v2Msg.Message.Content {
			if textBlock, ok := block.(*shared.TextBlock); ok {
				result.WriteString(textBlock.Text)
			}
		}
		return result.String()
	}

	// Handle shared.AssistantMessage (for compatibility)
	if assistantMsg, ok := msg.(*shared.AssistantMessage); ok {
		return shared.GetContentText(assistantMsg)
	}

	return ""
}

// ExtractResultText extracts the result text from a result message.
// Returns empty string if the message is not a result message.
func ExtractResultText(msg V2Message) string {
	if !IsResultMessage(msg) {
		return ""
	}

	// Handle V2ResultMessage
	if v2Msg, ok := msg.(*V2ResultMessage); ok {
		return v2Msg.Result
	}

	// Handle shared.ResultMessage (for compatibility)
	if resultMsg, ok := msg.(*shared.ResultMessage); ok {
		if resultMsg.Result != nil {
			return *resultMsg.Result
		}
	}

	return ""
}

// ExtractText extracts text from any message type that contains text.
// This is a convenience function that handles both assistant and result messages.
func ExtractText(msg V2Message) string {
	if text := ExtractAssistantText(msg); text != "" {
		return text
	}
	return ExtractResultText(msg)
}

// ExtractDeltaText extracts text from a stream delta message.
// Returns empty string if the message is not a stream delta or has no text.
func ExtractDeltaText(msg V2Message) string {
	if !IsStreamDelta(msg) {
		return ""
	}

	// Handle V2StreamDelta
	if v2Msg, ok := msg.(*V2StreamDelta); ok {
		if textDelta, ok := v2Msg.Delta["text"].(string); ok {
			return textDelta
		}
	}

	// Handle shared.StreamEvent (for compatibility)
	if streamEvent, ok := msg.(*shared.StreamEvent); ok {
		if textDelta, err := shared.ExtractDelta(streamEvent.Event); err == nil {
			return textDelta
		}
	}

	return ""
}

// ExtractErrorMessage extracts the error message from a V2Error message.
// Returns empty string if the message is not an error message.
func ExtractErrorMessage(msg V2Message) string {
	if !IsErrorMessage(msg) {
		return ""
	}

	// Handle V2Error
	if v2Msg, ok := msg.(*V2Error); ok {
		return v2Msg.ErrorField
	}

	return ""
}

// TextCarrier is an interface for types that carry text content.
type TextCarrier interface {
	GetText() string
}
