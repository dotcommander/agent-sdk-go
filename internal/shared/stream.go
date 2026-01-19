package shared

import (
	"encoding/json"
	"fmt"
	"strings"
)

// StreamIssue represents a validation issue with a stream message.
type StreamIssue struct {
	UUID   string
	Type   string
	Detail string
}

// StreamStats collects statistics about stream processing.
type StreamStats struct {
	TotalMessages    int
	PartialMessages int
	Errors          int
	ProcessingTime  string // Duration in human-readable format
}

// ValidateStreamEvent validates a stream event and returns any issues.
func ValidateStreamEvent(event StreamEvent) []StreamIssue {
	var issues []StreamIssue

	// Validate UUID
	if event.UUID == "" {
		issues = append(issues, StreamIssue{
			UUID:   event.UUID,
			Type:   "validation",
			Detail: "missing UUID",
		})
	}

	// Validate SessionID
	if event.SessionID == "" {
		issues = append(issues, StreamIssue{
			UUID:   event.UUID,
			Type:   "validation",
			Detail: "missing session_id",
		})
	}

	// Validate Event
	if event.Event == nil {
		issues = append(issues, StreamIssue{
			UUID:   event.UUID,
			Type:   "validation",
			Detail: "missing event",
		})
		return issues
	}

	// Validate event type
	eventType, ok := event.Event["type"].(string)
	if !ok {
		issues = append(issues, StreamIssue{
			UUID:   event.UUID,
			Type:   "validation",
			Detail: "missing or invalid event type",
		})
		return issues
	}

	// Validate based on event type
	switch eventType {
	case StreamEventTypeContentBlockStart:
		validateContentBlockStart(event.Event, &issues)
	case StreamEventTypeContentBlockDelta:
		validateContentBlockDelta(event.Event, &issues)
	case StreamEventTypeContentBlockStop:
		validateContentBlockStop(event.Event, &issues)
	case StreamEventTypeMessageStart:
		validateMessageStart(event.Event, &issues)
	case StreamEventTypeMessageDelta:
		validateMessageDelta(event.Event, &issues)
	case StreamEventTypeMessageStop:
		validateMessageStop(event.Event, &issues)
	default:
		issues = append(issues, StreamIssue{
			UUID:   event.UUID,
			Type:   "validation",
			Detail: fmt.Sprintf("unknown event type: %s", eventType),
		})
	}

	return issues
}

// validateContentBlockStart validates content_block_start events.
func validateContentBlockStart(event map[string]any, issues *[]StreamIssue) {
	// Validate index
	if _, ok := event["index"]; !ok {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "missing index in content_block_start",
		})
	} else if index, ok := event["index"].(float64); !ok || index < 0 {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "invalid index in content_block_start",
		})
	}

	// Validate content_block
	if _, ok := event["content_block"]; !ok {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "missing content_block in content_block_start",
		})
	}
}

// validateContentBlockDelta validates content_block_delta events.
func validateContentBlockDelta(event map[string]any, issues *[]StreamIssue) {
	// Validate index
	if _, ok := event["index"]; !ok {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "missing index in content_block_delta",
		})
	} else if index, ok := event["index"].(float64); !ok || index < 0 {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "invalid index in content_block_delta",
		})
	}

	// Validate delta
	if _, ok := event["delta"]; !ok {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "missing delta in content_block_delta",
		})
	}
}

// validateContentBlockStop validates content_block_stop events.
func validateContentBlockStop(event map[string]any, issues *[]StreamIssue) {
	// Validate index
	if _, ok := event["index"]; !ok {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "missing index in content_block_stop",
		})
	} else if index, ok := event["index"].(float64); !ok || index < 0 {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "invalid index in content_block_stop",
		})
	}
}

// validateMessageStart validates message_start events.
func validateMessageStart(event map[string]any, issues *[]StreamIssue) {
	// Validate message
	if _, ok := event["message"]; !ok {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "missing message in message_start",
		})
	}
}

// validateMessageDelta validates message_delta events.
func validateMessageDelta(event map[string]any, issues *[]StreamIssue) {
	// Validate delta
	if _, ok := event["delta"]; !ok {
		*issues = append(*issues, StreamIssue{
			Type:   "validation",
			Detail: "missing delta in message_delta",
		})
	}

	// Validate optional usage
	if _, ok := event["usage"]; ok {
		if _, ok := event["usage"].(map[string]any); !ok {
			*issues = append(*issues, StreamIssue{
				Type:   "validation",
				Detail: "invalid usage format in message_delta",
			})
		}
	}
}

// validateMessageStop validates message_stop events.
func validateMessageStop(event map[string]any, issues *[]StreamIssue) {
	// No additional validation required for message_stop
}

// ParseStreamEvent parses a raw JSON message into a StreamEvent.
func ParseStreamEvent(raw string) (*StreamEvent, error) {
	var event StreamEvent

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(raw), &event); err != nil {
		return nil, NewParserError(0, 0, raw, fmt.Sprintf("failed to parse stream event: %v", err))
	}

	// Validate the parsed event
	if issues := ValidateStreamEvent(event); len(issues) > 0 {
		return &event, fmt.Errorf("stream event validation failed: %v", issues)
	}

	return &event, nil
}

// ExtractDelta extracts the text content from a content_block_delta event.
func ExtractDelta(event map[string]any) (string, error) {
	delta, ok := event["delta"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("invalid delta format")
	}

	// Handle both text and thinking delta
	if text, ok := delta["text"].(string); ok {
		return text, nil
	}
	if thinking, ok := delta["thinking"].(string); ok {
		return thinking, nil
	}

	return "", fmt.Errorf("no text or thinking content in delta")
}

// FormatStats formats StreamStats into a human-readable string.
func FormatStats(stats StreamStats) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Total messages: %d", stats.TotalMessages))

	if stats.PartialMessages > 0 {
		builder.WriteString(", ")
		builder.WriteString(fmt.Sprintf("Partial messages: %d", stats.PartialMessages))
	}

	if stats.Errors > 0 {
		builder.WriteString(", ")
		builder.WriteString(fmt.Sprintf("Errors: %d", stats.Errors))
	}

	if stats.ProcessingTime != "" {
		builder.WriteString(", ")
		builder.WriteString(fmt.Sprintf("Processing time: %s", stats.ProcessingTime))
	}

	return builder.String()
}

// IsCriticalStreamEvent returns true if the event type is critical for response processing.
func IsCriticalStreamEvent(eventType string) bool {
	switch eventType {
	case StreamEventTypeMessageStart,
		StreamEventTypeMessageStop,
		StreamEventTypeContentBlockStart,
		StreamEventTypeContentBlockStop:
		return true
	default:
		return false
	}
}

// IsDeltaStreamEvent returns true if the event type contains delta content.
func IsDeltaStreamEvent(eventType string) bool {
	switch eventType {
	case StreamEventTypeContentBlockDelta,
		StreamEventTypeMessageDelta:
		return true
	default:
		return false
	}
}

// StreamEventTypeToString converts a stream event type to a human-readable string.
func StreamEventTypeToString(eventType string) string {
	switch eventType {
	case StreamEventTypeContentBlockStart:
		return "Content Block Started"
	case StreamEventTypeContentBlockDelta:
		return "Content Block Delta"
	case StreamEventTypeContentBlockStop:
		return "Content Block Stopped"
	case StreamEventTypeMessageStart:
		return "Message Started"
	case StreamEventTypeMessageDelta:
		return "Message Delta"
	case StreamEventTypeMessageStop:
		return "Message Stopped"
	default:
		return "Unknown Event"
	}
}

// ParseIndex extracts the index from a stream event.
func ParseIndex(event map[string]any) (int, error) {
	if index, ok := event["index"].(float64); ok {
		return int(index), nil
	}
	return 0, fmt.Errorf("invalid or missing index")
}

// CloneEvent creates a deep copy of a stream event.
func CloneEvent(event map[string]any) map[string]any {
	// Simple clone for maps - this is not a full deep clone but works for our use case
	clone := make(map[string]any)
	for k, v := range event {
		clone[k] = v
	}
	return clone
}