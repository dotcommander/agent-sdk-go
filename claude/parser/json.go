package parser

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

// Parser handles JSON message parsing with type discrimination.
type Parser struct {
	// Buffer for incomplete lines
	buffer string
	// Line number counter
	lineNumber int
	// Maximum buffer size to prevent memory exhaustion
	maxBufferSize int
	// Registry for message type parsers (OCP-compliant)
	registry *MessageParserRegistry
}

// NewParser creates a new Parser with the default registry.
func NewParser() *Parser {
	return &Parser{
		maxBufferSize: 1024 * 1024, // 1MB buffer limit
		registry:      DefaultRegistry(),
	}
}

// NewParserWithRegistry creates a new Parser with a custom registry.
// This allows for registering custom message types without modifying the parser.
func NewParserWithRegistry(registry *MessageParserRegistry) *Parser {
	return &Parser{
		maxBufferSize: 1024 * 1024, // 1MB buffer limit
		registry:      registry,
	}
}

// ParseMessage parses a raw JSON message into a Message interface.
// Returns the message type and the parsed message, or an error.
// If the JSON is incomplete, returns (nil, nil) and buffers the input for later completion.
func (p *Parser) ParseMessage(raw string) (shared.Message, error) {
	// Check buffer size
	if len(p.buffer)+len(raw) > p.maxBufferSize {
		return nil, shared.NewParserError(p.lineNumber, 0, raw, "buffer size exceeded")
	}

	// Append to buffer
	p.buffer += raw

	// Try to parse complete JSON objects
	messages, remaining, err := p.parseBuffer()
	if err != nil {
		return nil, err
	}

	// Update buffer and line number
	p.buffer = remaining
	p.lineNumber += countLines(raw)

	// Return the first message (there should be only one)
	// If no complete JSON object found, return nil (buffered for later)
	if len(messages) == 0 {
		return nil, nil
	}

	if len(messages) > 1 {
		// This could happen if there are multiple JSON objects in one message
		// For now, just return the first one and log the rest
		slog.Warn("multiple JSON objects in single message, returning first", "count", len(messages))
	}

	return messages[0], nil
}

// ParseMessages parses a raw JSON string that may contain multiple messages.
func (p *Parser) ParseMessages(raw string) ([]shared.Message, error) {
	// Check buffer size
	if len(p.buffer)+len(raw) > p.maxBufferSize {
		return nil, shared.NewParserError(p.lineNumber, 0, raw, "buffer size exceeded")
	}

	// Append to buffer
	p.buffer += raw

	// Parse buffer
	messages, remaining, err := p.parseBuffer()
	if err != nil {
		return nil, err
	}

	// Update buffer and line number
	p.buffer = remaining
	// Increment line number by newlines or by message count (whichever is greater)
	lineIncrement := countLines(raw)
	if len(messages) > lineIncrement {
		lineIncrement = len(messages)
	}
	p.lineNumber += lineIncrement

	return messages, nil
}

// parseBuffer tries to parse complete JSON objects from the buffer.
func (p *Parser) parseBuffer() ([]shared.Message, string, error) {
	var messages []shared.Message
	var buffer = p.buffer

	// Keep trying to parse JSON objects until we can't anymore
	for {
		// Skip whitespace
		buffer = skipWhitespace(buffer)

		if len(buffer) == 0 {
			break
		}

		// Check if it starts with '{' (JSON object)
		if buffer[0] != '{' {
			// Not a JSON object, can't parse
			break
		}

		// Find the end of the JSON object
		end, err := findJSONObjectEnd(buffer)
		if err != nil {
			// Incomplete JSON, stop parsing
			break
		}

		// Extract the JSON object
		jsonStr := buffer[:end]
		remaining := buffer[end:]

		// Parse the JSON object
		msg, err := p.parseJSONObject(jsonStr)
		if err != nil {
			return messages, remaining, err
		}

		messages = append(messages, msg)
		buffer = remaining
	}

	return messages, buffer, nil
}

// parseJSONObject parses a JSON object into a Message interface.
// Uses the registry pattern (OCP-compliant) - new message types can be added
// by registering them with the parser's registry without modifying this code.
func (p *Parser) parseJSONObject(jsonStr string) (shared.Message, error) {
	// First, get the type field
	var typeData struct {
		Type string `json:"type"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &typeData); err != nil {
		return nil, shared.NewParserError(p.lineNumber, 0, jsonStr, fmt.Sprintf("failed to parse type: %v", err))
	}

	// Parse using the registry (OCP: open for extension, closed for modification)
	return p.registry.Parse(typeData.Type, jsonStr, p.lineNumber)
}

// Registry returns the parser's message type registry.
// Use this to register custom message types.
func (p *Parser) Registry() *MessageParserRegistry {
	return p.registry
}

// findJSONObjectEnd finds the end of a JSON object starting at index 0.
func findJSONObjectEnd(s string) (int, error) {
	var depth int
	var inString bool
	var escape bool

	for i, r := range s {
		if escape {
			escape = false
			continue
		}

		switch r {
		case '\\':
			escape = true
		case '"':
			inString = !inString
		case '{':
			if !inString {
				depth++
			}
		case '}':
			if !inString {
				depth--
				if depth == 0 {
					return i + 1, nil
				}
			}
		}
	}

	return 0, fmt.Errorf("incomplete JSON object")
}

// skipWhitespace skips whitespace characters and commas from the beginning of a string.
// Commas are skipped to support comma-separated JSON objects in streams.
func skipWhitespace(s string) string {
	for i, r := range s {
		if !isWhitespaceOrComma(r) {
			return s[i:]
		}
	}
	return ""
}

// isWhitespaceOrComma checks if a character is whitespace or a comma separator.
func isWhitespaceOrComma(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == ','
}

// countLines counts the number of newline characters in a string.
func countLines(s string) int {
	count := 0
	for _, r := range s {
		if r == '\n' {
			count++
		}
	}
	return count
}

// Reset resets the parser state.
func (p *Parser) Reset() {
	p.buffer = ""
	p.lineNumber = 0
}

// GetBufferSize returns the current buffer size.
func (p *Parser) GetBufferSize() int {
	return len(p.buffer)
}

// GetLineNumber returns the current line number.
func (p *Parser) GetLineNumber() int {
	return p.lineNumber
}