package parser

import (
	"encoding/json"
	"fmt"
	"sync"

	"agent-sdk-go/internal/claude/shared"
)

// MessageParserFunc is a function that parses a JSON string into a Message.
// The lineNumber parameter is provided for error reporting.
type MessageParserFunc func(jsonStr string, lineNumber int) (shared.Message, error)

// MessageParserRegistry provides a registry for message type parsers.
// New message types can be registered without modifying the parser code (OCP).
type MessageParserRegistry struct {
	mu      sync.RWMutex
	parsers map[string]MessageParserFunc
}

// NewMessageParserRegistry creates a new registry with default parsers registered.
func NewMessageParserRegistry() *MessageParserRegistry {
	r := &MessageParserRegistry{
		parsers: make(map[string]MessageParserFunc),
	}
	r.registerDefaults()
	return r
}

// Register registers a parser function for a message type.
// If a parser for this type already exists, it will be replaced.
func (r *MessageParserRegistry) Register(messageType string, parser MessageParserFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.parsers[messageType] = parser
}

// Unregister removes a parser for a message type.
func (r *MessageParserRegistry) Unregister(messageType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.parsers, messageType)
}

// Parse parses a JSON string into a Message using the registered parser.
// Returns an error if no parser is registered for the message type.
func (r *MessageParserRegistry) Parse(messageType, jsonStr string, lineNumber int) (shared.Message, error) {
	r.mu.RLock()
	parser, ok := r.parsers[messageType]
	r.mu.RUnlock()

	if !ok {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("unknown message type: %s", messageType))
	}

	return parser(jsonStr, lineNumber)
}

// HasParser returns true if a parser is registered for the message type.
func (r *MessageParserRegistry) HasParser(messageType string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.parsers[messageType]
	return ok
}

// RegisteredTypes returns a slice of all registered message types.
func (r *MessageParserRegistry) RegisteredTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	types := make([]string, 0, len(r.parsers))
	for t := range r.parsers {
		types = append(types, t)
	}
	return types
}

// registerDefaults registers the default message type parsers.
func (r *MessageParserRegistry) registerDefaults() {
	r.Register(shared.MessageTypeUser, parseUserMessage)
	r.Register(shared.MessageTypeAssistant, parseAssistantMessage)
	r.Register(shared.MessageTypeSystem, parseSystemMessage)
	r.Register(shared.MessageTypeResult, parseResultMessage)
	r.Register(shared.MessageTypeStreamEvent, parseStreamEvent)
	r.Register(shared.MessageTypeControlRequest, parseControlRequest)
	r.Register(shared.MessageTypeControlResponse, parseControlResponse)
	r.Register(shared.MessageTypeToolProgress, parseToolProgressMessage)
	r.Register(shared.MessageTypeAuthStatus, parseAuthStatusMessage)
}

// Default parser functions

func parseUserMessage(jsonStr string, lineNumber int) (shared.Message, error) {
	var msg shared.UserMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse UserMessage: %v", err))
	}
	return &msg, nil
}

func parseAssistantMessage(jsonStr string, lineNumber int) (shared.Message, error) {
	var msg shared.AssistantMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse AssistantMessage: %v", err))
	}
	return &msg, nil
}

func parseSystemMessage(jsonStr string, lineNumber int) (shared.Message, error) {
	// First unmarshal into a raw map to capture all fields
	var rawData map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse SystemMessage: %v", err))
	}

	msg := &shared.SystemMessage{
		Data: make(map[string]any),
	}

	// Extract known fields and collect the rest into Data
	for k, v := range rawData {
		switch k {
		case "type":
			// Skip, this is handled by the parser
		case "subtype":
			var subtype string
			if err := json.Unmarshal(v, &subtype); err == nil {
				msg.Subtype = subtype
			}
		case "agents":
			var agents []string
			if err := json.Unmarshal(v, &agents); err == nil {
				msg.Agents = agents
			}
			// Also store in Data for backward compatibility
			var data any
			if err := json.Unmarshal(v, &data); err == nil {
				msg.Data[k] = data
			}
		case "betas":
			var betas []string
			if err := json.Unmarshal(v, &betas); err == nil {
				msg.Betas = betas
			}
			var data any
			if err := json.Unmarshal(v, &data); err == nil {
				msg.Data[k] = data
			}
		case "claudeCodeVersion":
			var version string
			if err := json.Unmarshal(v, &version); err == nil {
				msg.ClaudeCodeVersion = version
			}
			var data any
			if err := json.Unmarshal(v, &data); err == nil {
				msg.Data[k] = data
			}
		case "skills":
			var skills []string
			if err := json.Unmarshal(v, &skills); err == nil {
				msg.Skills = skills
			}
			var data any
			if err := json.Unmarshal(v, &data); err == nil {
				msg.Data[k] = data
			}
		case "plugins":
			var plugins []shared.PluginInfo
			if err := json.Unmarshal(v, &plugins); err == nil {
				msg.Plugins = plugins
			}
			var data any
			if err := json.Unmarshal(v, &data); err == nil {
				msg.Data[k] = data
			}
		default:
			// All other fields go into Data
			var data any
			if err := json.Unmarshal(v, &data); err != nil {
				// If we can't parse, keep it as raw JSON string
				data = string(v)
			}
			msg.Data[k] = data
		}
	}

	return msg, nil
}

func parseResultMessage(jsonStr string, lineNumber int) (shared.Message, error) {
	var msg shared.ResultMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse ResultMessage: %v", err))
	}
	return &msg, nil
}

func parseStreamEvent(jsonStr string, lineNumber int) (shared.Message, error) {
	var msg shared.StreamEvent
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse StreamEvent: %v", err))
	}
	return &msg, nil
}

func parseControlRequest(jsonStr string, lineNumber int) (shared.Message, error) {
	var msg shared.RawControlMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse ControlRequest: %v", err))
	}
	return &msg, nil
}

func parseControlResponse(jsonStr string, lineNumber int) (shared.Message, error) {
	var msg shared.RawControlMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse ControlResponse: %v", err))
	}
	return &msg, nil
}

func parseToolProgressMessage(jsonStr string, lineNumber int) (shared.Message, error) {
	var msg shared.ToolProgressMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse ToolProgressMessage: %v", err))
	}
	return &msg, nil
}

func parseAuthStatusMessage(jsonStr string, lineNumber int) (shared.Message, error) {
	var msg shared.AuthStatusMessage
	if err := json.Unmarshal([]byte(jsonStr), &msg); err != nil {
		return nil, shared.NewParserError(lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse AuthStatusMessage: %v", err))
	}
	return &msg, nil
}

// defaultRegistry is the global default registry instance.
var defaultRegistry = NewMessageParserRegistry()

// DefaultRegistry returns the default message parser registry.
// This allows external packages to register custom message types.
func DefaultRegistry() *MessageParserRegistry {
	return defaultRegistry
}
