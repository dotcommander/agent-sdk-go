package claude

import (
	"context"
	"github.com/dotcommander/agent-sdk-go/claude/parser"
	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

// ClientOption is a function type for configuring the client.
type ClientOption func(*ClientOptions)

// ClientOptions provides configuration for the Claude client.
// Composes focused option structs for Single Responsibility adherence.
type ClientOptions struct {
	// Connection options (CLI path, command, timeout, env)
	shared.ConnectionOptions

	// Buffer options (buffer size, max messages)
	shared.BufferOptions

	// Model options (model name, permission mode, context files, custom args)
	shared.ModelOptions

	// Debug options (trace, cache settings, logger, metrics)
	shared.DebugOptions

	// IncludePartialMessages enables streaming of partial messages.
	IncludePartialMessages bool

	// EnableStructuredOutput enables structured output mode.
	EnableStructuredOutput bool

	// McpServers are MCP server configurations.
	McpServers map[string]shared.McpServerConfig
}

// BasicTransport provides core transport functionality.
type BasicTransport interface {
	// Connect establishes a connection to Claude CLI.
	Connect(ctx context.Context) error

	// SendMessage sends a message to Claude CLI.
	// The message should be formatted as per the Claude Code protocol.
	SendMessage(ctx context.Context, message string) error

	// ReceiveMessages receives messages from Claude CLI.
	// Returns a channel of messages and a channel for errors.
	ReceiveMessages(ctx context.Context) (<-chan string, <-chan error)

	// Close gracefully shuts down the transport.
	Close() error

	// Interrupt forcibly interrupts the transport.
	Interrupt() error

	// IsConnected returns whether the transport is currently connected.
	IsConnected() bool
}

// ProcessTransport extends BasicTransport with process-specific methods.
type ProcessTransport interface {
	BasicTransport
	// GetCommand returns the command used to start Claude CLI.
	GetCommand() string

	// GetPID returns the process ID of Claude CLI.
	GetPID() int
}

// Transport is the interface for communicating with Claude CLI.
// It combines BasicTransport and ProcessTransport for backward compatibility.
type Transport interface {
	ProcessTransport
}

// Connector handles connection lifecycle.
type Connector interface {
	// Connect establishes a connection to Claude CLI.
	Connect(ctx context.Context) error

	// Disconnect closes the connection to Claude CLI.
	Disconnect() error
}

// Querier handles query operations.
type Querier interface {
	// Query sends a one-shot query to Claude CLI and returns the response.
	Query(ctx context.Context, prompt string) (string, error)

	// QueryWithSession sends a query with session context.
	QueryWithSession(ctx context.Context, sessionID string, prompt string) (string, error)

	// QueryStream sends a one-shot query and streams the response.
	QueryStream(ctx context.Context, prompt string) (<-chan Message, <-chan error)
}

// Receiver handles message reception.
type Receiver interface {
	// ReceiveMessages receives messages from Claude CLI.
	ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error)

	// ReceiveResponse receives a single response from Claude CLI.
	ReceiveResponse(ctx context.Context) (Message, error)
}

// Controller handles runtime configuration.
type Controller interface {
	// Interrupt forcibly interrupts the current operation (process-level).
	Interrupt() error

	// InterruptGraceful sends interrupt via control protocol (if active).
	InterruptGraceful(ctx context.Context) error

	// SetModel changes the AI model during a streaming session.
	// Pass nil to reset to the default model.
	// Only works when control protocol is active (connected streaming session).
	SetModel(ctx context.Context, model *string) error

	// SetPermissionMode changes the permission mode during a streaming session.
	// Valid modes: "default", "acceptEdits", "plan", "bypassPermissions", "delegate", "dontAsk"
	// Only works when control protocol is active.
	SetPermissionMode(ctx context.Context, mode string) error

	// RewindFiles reverts tracked files to their state at a specific user message.
	// The messageUUID should be the UUID from a UserMessage received during the session.
	// Requires EnableFileCheckpointing option.
	RewindFiles(ctx context.Context, messageUUID string) error

	// GetStreamIssues returns validation issues found in the message stream.
	GetStreamIssues() []shared.StreamIssue

	// GetStreamStats returns statistics about the message stream.
	GetStreamStats() shared.StreamStats

	// GetServerInfo returns diagnostic information about the client connection.
	GetServerInfo(ctx context.Context) (map[string]any, error)

	// IsProtocolActive returns whether the control protocol is active.
	IsProtocolActive() bool

	// McpServerStatus returns MCP server statuses.
	McpServerStatus(ctx context.Context) ([]shared.McpServerStatus, error)

	// SetMcpServers dynamically sets MCP servers.
	SetMcpServers(ctx context.Context, servers map[string]shared.McpServerConfig) (*shared.McpSetServersResult, error)
}

// ContextManager handles context file operations.
type ContextManager interface {
	// AddContextFiles adds files to the context for the next query.
	AddContextFiles(ctx context.Context, files []string) error

	// GetOptions returns a copy of the client options.
	GetOptions() *ClientOptions
}

// Client represents the full Claude client interface.
// It composes all sub-interfaces for backward compatibility.
type Client interface {
	Connector
	Querier
	Receiver
	Controller
	ContextManager
}

// Message represents any message type in the Claude Code protocol.
type Message = shared.Message

// ContentBlock represents any content block within a message.
type ContentBlock = shared.ContentBlock

// UserMessage represents a message from the user.
type UserMessage = shared.UserMessage

// AssistantMessage represents a message from the assistant.
type AssistantMessage = shared.AssistantMessage

// SystemMessage represents a system message.
type SystemMessage = shared.SystemMessage

// ResultMessage represents the final result of a conversation turn.
type ResultMessage = shared.ResultMessage

// TextBlock represents text content.
type TextBlock = shared.TextBlock

// ThinkingBlock represents thinking content with signature.
type ThinkingBlock = shared.ThinkingBlock

// ToolUseBlock represents a tool use request.
type ToolUseBlock = shared.ToolUseBlock

// ToolResultBlock represents the result of a tool use.
type ToolResultBlock = shared.ToolResultBlock

// StreamEvent represents a partial message update during streaming.
type StreamEvent = shared.StreamEvent

// AssistantMessageError represents error types in assistant messages.
type AssistantMessageError = shared.AssistantMessageError

// StreamIssue represents a validation issue with a stream message.
type StreamIssue = shared.StreamIssue

// StreamStats collects statistics about stream processing.
type StreamStats = shared.StreamStats

// Parser represents the interface for parsing JSON messages.
type Parser interface {
	// ParseMessage parses a raw JSON message into a Message interface.
	ParseMessage(raw string) (Message, error)

	// ParseMessages parses a raw JSON string that may contain multiple messages.
	ParseMessages(raw string) ([]Message, error)

	// Reset resets the parser state.
	Reset()

	// GetBufferSize returns the current buffer size.
	GetBufferSize() int

	// GetLineNumber returns the current line number.
	GetLineNumber() int
}

// Check functions from shared package
func IsCLINotFound(err error) bool {
	return shared.IsCLINotFound(err)
}

func IsConnectionError(err error) bool {
	return shared.IsConnectionError(err)
}

func IsTimeoutError(err error) bool {
	return shared.IsTimeoutError(err)
}

func IsParserError(err error) bool {
	return shared.IsParserError(err)
}

func IsProtocolError(err error) bool {
	return shared.IsProtocolError(err)
}

func NewCLINotFoundError(path, command string) *shared.CLINotFoundError {
	return shared.NewCLINotFoundError(path, command)
}

func NewConnectionError(reason string, inner error) *shared.ConnectionError {
	return shared.NewConnectionError(reason, inner)
}

func NewTimeoutError(operation, timeout string) *shared.TimeoutError {
	return shared.NewTimeoutError(operation, timeout)
}

func NewParserError(line, offset int, data, reason string) *shared.ParserError {
	return shared.NewParserError(line, offset, data, reason)
}

func NewProtocolError(messageType, reason string) *shared.ProtocolError {
	return shared.NewProtocolError(messageType, reason)
}


func ProcessError(pid int, command, reason, signal string) *shared.ProcessError {
	return shared.NewProcessError(pid, command, reason, signal)
}

// Default command function
func GetDefaultCommand() string {
	return shared.GetDefaultCommand()
}

// Parser creation function
func NewParser() Parser {
	return parser.NewParser()
}

// MessageParserRegistry provides a registry for message type parsers.
// Exported to allow external packages to register custom message types.
type MessageParserRegistry = parser.MessageParserRegistry

// MessageParserFunc is a function that parses a JSON string into a Message.
type MessageParserFunc = parser.MessageParserFunc

// NewMessageParserRegistry creates a new registry with default parsers registered.
func NewMessageParserRegistry() *MessageParserRegistry {
	return parser.NewMessageParserRegistry()
}

// DefaultParserRegistry returns the default message parser registry.
// This allows external packages to register custom message types.
func DefaultParserRegistry() *MessageParserRegistry {
	return parser.DefaultRegistry()
}

// NewParserWithRegistry creates a new Parser with a custom registry.
func NewParserWithRegistry(registry *MessageParserRegistry) Parser {
	return parser.NewParserWithRegistry(registry)
}