package shared

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"
)

// ValidateMessage validates a message for protocol compliance.
func ValidateMessage(msg Message) error {
	switch m := msg.(type) {
	case *UserMessage:
		return ValidateUserMessage(m)
	case *AssistantMessage:
		return ValidateAssistantMessage(m)
	case *SystemMessage:
		return ValidateSystemMessage(m)
	case *ResultMessage:
		return ValidateResultMessage(m)
	case *StreamEvent:
		if issues := ValidateStreamEvent(*m); len(issues) > 0 {
			return fmt.Errorf("stream event validation failed: %v", issues)
		}
		return nil
	default:
		return NewProtocolError("", fmt.Sprintf("unknown message type: %T", msg))
	}
}

// ValidateUserMessage validates a UserMessage.
func ValidateUserMessage(msg *UserMessage) error {
	if msg == nil {
		return errors.New("UserMessage is nil")
	}

	// Validate content
	switch content := msg.Content.(type) {
	case string:
		if strings.TrimSpace(content) == "" {
			return errors.New("UserMessage content cannot be empty")
		}
	case []ContentBlock:
		if len(content) == 0 {
			return errors.New("UserMessage content blocks cannot be empty")
		}
		for i, block := range content {
			if err := ValidateContentBlock(block); err != nil {
				return fmt.Errorf("content block %d: %w", i, err)
			}
		}
	case nil:
		return errors.New("UserMessage content cannot be nil")
	default:
		return fmt.Errorf("invalid UserMessage content type: %T", content)
	}

	return nil
}

// ValidateAssistantMessage validates an AssistantMessage.
func ValidateAssistantMessage(msg *AssistantMessage) error {
	if msg == nil {
		return errors.New("AssistantMessage is nil")
	}

	// Validate model
	if strings.TrimSpace(msg.Model) == "" {
		return errors.New("AssistantMessage model cannot be empty")
	}

	// Validate content
	if len(msg.Content) == 0 {
		return errors.New("AssistantMessage content cannot be empty")
	}

	for i, block := range msg.Content {
		if err := ValidateContentBlock(block); err != nil {
			return fmt.Errorf("content block %d: %w", i, err)
		}
	}

	// Validate error if present
	if msg.Error != nil {
		// Check if error is one of the valid types
		validErrors := map[AssistantMessageError]bool{
			AssistantMessageErrorAuthFailed:     true,
			AssistantMessageErrorBilling:        true,
			AssistantMessageErrorRateLimit:      true,
			AssistantMessageErrorInvalidRequest: true,
			AssistantMessageErrorServer:         true,
			AssistantMessageErrorUnknown:        true,
		}
		if !validErrors[*msg.Error] {
			return fmt.Errorf("invalid AssistantMessage error type: %s", *msg.Error)
		}
	}

	return nil
}

// ValidateSystemMessage validates a SystemMessage.
func ValidateSystemMessage(msg *SystemMessage) error {
	if msg == nil {
		return errors.New("SystemMessage is nil")
	}

	// Validate subtype
	if strings.TrimSpace(msg.Subtype) == "" {
		return errors.New("SystemMessage subtype cannot be empty")
	}

	// Validate data
	if msg.Data == nil {
		return errors.New("SystemMessage data cannot be nil")
	}

	return nil
}

// ValidateResultMessage validates a ResultMessage.
func ValidateResultMessage(msg *ResultMessage) error {
	if msg == nil {
		return errors.New("ResultMessage is nil")
	}

	// Validate subtype
	if strings.TrimSpace(msg.Subtype) == "" {
		return errors.New("ResultMessage subtype cannot be empty")
	}

	// Validate duration
	if msg.DurationMs < 0 {
		return fmt.Errorf("ResultMessage duration_ms cannot be negative: %d", msg.DurationMs)
	}

	// Validate API duration
	if msg.DurationAPIMs < 0 {
		return fmt.Errorf("ResultMessage duration_api_ms cannot be negative: %d", msg.DurationAPIMs)
	}

	// Validate number of turns
	if msg.NumTurns < 0 {
		return fmt.Errorf("ResultMessage num_turns cannot be negative: %d", msg.NumTurns)
	}

	// Validate session ID
	if strings.TrimSpace(msg.SessionID) == "" {
		return errors.New("ResultMessage session_id cannot be empty")
	}

	// Validate cost if present
	if msg.TotalCostUSD != nil {
		if *msg.TotalCostUSD < 0 {
			return fmt.Errorf("ResultMessage total_cost_usd cannot be negative: %f", *msg.TotalCostUSD)
		}
	}

	return nil
}

// ValidateContentBlock validates a ContentBlock.
func ValidateContentBlock(block ContentBlock) error {
	if block == nil {
		return errors.New("ContentBlock is nil")
	}

	switch b := block.(type) {
	case *TextBlock:
		return ValidateTextBlock(b)
	case *ThinkingBlock:
		return ValidateThinkingBlock(b)
	case *ToolUseBlock:
		return ValidateToolUseBlock(b)
	case *ToolResultBlock:
		return ValidateToolResultBlock(b)
	default:
		return fmt.Errorf("unknown ContentBlock type: %T", block)
	}
}

// ValidateTextBlock validates a TextBlock.
func ValidateTextBlock(block *TextBlock) error {
	if block == nil {
		return errors.New("TextBlock is nil")
	}

	if strings.TrimSpace(block.Text) == "" {
		return errors.New("TextBlock text cannot be empty")
	}

	return nil
}

// ValidateThinkingBlock validates a ThinkingBlock.
func ValidateThinkingBlock(block *ThinkingBlock) error {
	if block == nil {
		return errors.New("ThinkingBlock is nil")
	}

	if strings.TrimSpace(block.Thinking) == "" {
		return errors.New("ThinkingBlock thinking cannot be empty")
	}

	if strings.TrimSpace(block.Signature) == "" {
		return errors.New("ThinkingBlock signature cannot be empty")
	}

	return nil
}

// ValidateToolUseBlock validates a ToolUseBlock.
func ValidateToolUseBlock(block *ToolUseBlock) error {
	if block == nil {
		return errors.New("ToolUseBlock is nil")
	}

	if strings.TrimSpace(block.ToolUseID) == "" {
		return errors.New("ToolUseBlock tool_use_id cannot be empty")
	}

	if strings.TrimSpace(block.Name) == "" {
		return errors.New("ToolUseBlock name cannot be empty")
	}

	if block.Input == nil {
		return errors.New("ToolUseBlock input cannot be nil")
	}

	return nil
}

// ValidateToolResultBlock validates a ToolResultBlock.
func ValidateToolResultBlock(block *ToolResultBlock) error {
	if block == nil {
		return errors.New("ToolResultBlock is nil")
	}

	if strings.TrimSpace(block.ToolUseID) == "" {
		return errors.New("ToolResultBlock tool_use_id cannot be empty")
	}

	switch block.Content.(type) {
	case nil, string, map[string]any:
		// Valid types
	default:
		return fmt.Errorf("invalid ToolResultBlock content type: %T", block.Content)
	}

	if block.IsError != nil && *block.IsError && block.Content == nil {
		return errors.New("ToolResultBlock content cannot be nil when is_error is true")
	}

	return nil
}

// ValidateContextFile validates a context file path.
func ValidateContextFile(file string) error {
	if strings.TrimSpace(file) == "" {
		return errors.New("context file path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("context file does not exist: %s", file)
	}

	// Check if file is a regular file (not directory)
	if info, err := os.Stat(file); err == nil && info.IsDir() {
		return fmt.Errorf("context file is a directory: %s", file)
	}

	return nil
}

// ValidateContextFiles validates a list of context files.
func ValidateContextFiles(files []string) error {
	if len(files) == 0 {
		return errors.New("no context files provided")
	}

	for i, file := range files {
		if err := ValidateContextFile(file); err != nil {
			return fmt.Errorf("context file %d: %w", i, err)
		}
	}

	return nil
}

// ValidateModel validates a model name.
// Accepts any non-empty model name to support alternate providers (zai, synthetic).
func ValidateModel(model string) error {
	if strings.TrimSpace(model) == "" {
		return errors.New("model name cannot be empty")
	}

	// Note: We intentionally don't validate the claude- prefix here
	// to support alternate providers that use different model names
	// (e.g., GLM-4.7 for zai, DeepSeek-V3.2 for synthetic)

	return nil
}

// ValidatePermissionMode validates a permission mode.
func ValidatePermissionMode(mode string) error {
	validModes := []string{"auto", "read", "write", "restricted"}
	if slices.Contains(validModes, mode) {
		return nil
	}
	return fmt.Errorf("invalid permission mode: %s (must be one of: %v)", mode, validModes)
}

// ValidateTimeout validates a timeout string.
func ValidateTimeout(timeout string) error {
	if strings.TrimSpace(timeout) == "" {
		return errors.New("timeout cannot be empty")
	}

	// Simple validation - just check that it ends with time unit
	if !strings.HasSuffix(timeout, "s") && !strings.HasSuffix(timeout, "m") && !strings.HasSuffix(timeout, "h") {
		return errors.New("timeout must end with 's', 'm', or 'h'")
	}

	return nil
}

// ValidateEnvironmentVariables validates environment variables.
func ValidateEnvironmentVariables(env map[string]string) error {
	for key, value := range env {
		if strings.TrimSpace(key) == "" {
			return errors.New("environment variable key cannot be empty")
		}

		// Validate key format (only letters, numbers, underscores)
		for _, r := range key {
			if !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_') {
				return fmt.Errorf("invalid environment variable key '%s': contains invalid character '%c'", key, r)
			}
		}

		if value == "" {
			return fmt.Errorf("environment variable '%s' cannot have empty value", key)
		}
	}

	return nil
}

// ValidateFileExtension validates a file extension against allowed extensions.
func ValidateFileExtension(file string, allowedExtensions []string) error {
	ext := strings.ToLower(filepath.Ext(file))
	if ext == "" {
		return fmt.Errorf("file '%s' has no extension", file)
	}

	for _, allowed := range allowedExtensions {
		if strings.ToLower(allowed) == ext {
			return nil
		}
	}

	return fmt.Errorf("file '%s' has disallowed extension '%s' (allowed: %v)", file, ext, allowedExtensions)
}

// ValidateFileSize validates a file size against a maximum size.
func ValidateFileSize(file string, maxSize int64) error {
	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("failed to stat file '%s': %w", file, err)
	}

	if info.Size() > maxSize {
		return fmt.Errorf("file '%s' exceeds maximum size of %d bytes (actual: %d bytes)",
			file, maxSize, info.Size())
	}

	return nil
}

// SanitizeMessage returns a sanitized version of a message for logging.
func SanitizeMessage(msg Message) any {
	switch m := msg.(type) {
	case *UserMessage:
		return sanitizeUserMessage(m)
	case *AssistantMessage:
		return sanitizeAssistantMessage(m)
	case *SystemMessage:
		return sanitizeSystemMessage(m)
	case *ResultMessage:
		return sanitizeResultMessage(m)
	case *StreamEvent:
		return sanitizeStreamEvent(m)
	default:
		return fmt.Sprintf("unknown message type: %T", msg)
	}
}

// Helper functions for sanitization
func sanitizeUserMessage(msg *UserMessage) map[string]any {
	sanitized := map[string]any{
		"type": msg.Type(),
	}

	switch content := msg.Content.(type) {
	case string:
		sanitized["content"] = "<redacted>" // Redact user content for logging
	case []ContentBlock:
		sanitizedBlocks := make([]any, len(content))
		for i, block := range content {
			sanitizedBlocks[i] = sanitizeContentBlock(block)
		}
		sanitized["content"] = sanitizedBlocks
	}

	if msg.UUID != nil {
		sanitized["uuid"] = *msg.UUID
	}
	if msg.ParentToolUseID != nil {
		sanitized["parent_tool_use_id"] = *msg.ParentToolUseID
	}

	return sanitized
}

func sanitizeAssistantMessage(msg *AssistantMessage) map[string]any {
	sanitized := map[string]any{
		"type":  msg.Type(),
		"model": msg.Model,
	}

	if msg.Error != nil {
		sanitized["error"] = string(*msg.Error)
	}

	sanitizedBlocks := make([]any, len(msg.Content))
	for i, block := range msg.Content {
		sanitizedBlocks[i] = sanitizeContentBlock(block)
	}
	sanitized["content"] = sanitizedBlocks

	return sanitized
}

func sanitizeSystemMessage(msg *SystemMessage) map[string]any {
	return map[string]any{
		"type":    msg.Type(),
		"subtype": msg.Subtype,
	}
}

func sanitizeResultMessage(msg *ResultMessage) map[string]any {
	sanitized := map[string]any{
		"type":        msg.Type(),
		"subtype":     msg.Subtype,
		"duration_ms": msg.DurationMs,
		"num_turns":   msg.NumTurns,
		"session_id":  msg.SessionID,
		"is_error":    msg.IsError,
	}

	if msg.TotalCostUSD != nil {
		sanitized["total_cost_usd"] = *msg.TotalCostUSD
	}

	return sanitized
}

func sanitizeStreamEvent(msg *StreamEvent) map[string]any {
	sanitized := map[string]any{
		"type":  msg.Type(),
		"uuid":  msg.UUID,
		"event": "<redacted>", // Redact event content for logging
	}

	if msg.ParentToolUseID != nil {
		sanitized["parent_tool_use_id"] = *msg.ParentToolUseID
	}

	return sanitized
}

func sanitizeContentBlock(block ContentBlock) any {
	switch b := block.(type) {
	case *TextBlock:
		return map[string]any{
			"type": "text",
			"text": "<redacted>", // Redact text content for logging
		}
	case *ThinkingBlock:
		return map[string]any{
			"type":      "thinking",
			"thinking":  "<redacted>", // Redact thinking content for logging
			"signature": b.Signature,
		}
	case *ToolUseBlock:
		return map[string]any{
			"type":        "tool_use",
			"tool_use_id": b.ToolUseID,
			"name":        b.Name,
			"input":       "<redacted>", // Redact input for logging
		}
	case *ToolResultBlock:
		return map[string]any{
			"type":        "tool_result",
			"tool_use_id": b.ToolUseID,
			"is_error":    b.IsError,
		}
	default:
		return fmt.Sprintf("unknown content block type: %T", block)
	}
}

// StreamValidator tracks tool requests and results to detect incomplete streams.
// It provides validation for stream integrity and collects statistics about message processing.
type StreamValidator struct {
	mu              sync.RWMutex
	toolsRequested  map[string]bool // tool_use_id -> seen
	toolsReceived   map[string]bool // tool_use_id -> seen
	totalMessages   int
	partialMessages int
	errors          int
	startTime       time.Time
	streamEnded     bool
}

// NewStreamValidator creates a new stream validator.
func NewStreamValidator() *StreamValidator {
	return &StreamValidator{
		toolsRequested: make(map[string]bool),
		toolsReceived:  make(map[string]bool),
		startTime:      time.Now(),
	}
}

// TrackMessage tracks a message and updates validation state.
// Call this for each message received from the stream.
func (v *StreamValidator) TrackMessage(msg Message) {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.totalMessages++

	if msg == nil {
		return
	}

	switch m := msg.(type) {
	case *AssistantMessage:
		// Track tool use requests from assistant message content blocks
		for _, block := range m.Content {
			switch b := block.(type) {
			case *ToolUseBlock:
				if b.ToolUseID != "" {
					v.toolsRequested[b.ToolUseID] = true
				}
			case *ToolResultBlock:
				if b.ToolUseID != "" {
					v.toolsReceived[b.ToolUseID] = true
				}
			}
		}

	case *UserMessage:
		// Track tool results from user message content blocks
		if blocks, ok := m.Content.([]ContentBlock); ok {
			for _, block := range blocks {
				if toolResult, ok := block.(*ToolResultBlock); ok {
					if toolResult.ToolUseID != "" {
						v.toolsReceived[toolResult.ToolUseID] = true
					}
				}
			}
		}

	case *StreamEvent:
		// Stream events are partial messages
		v.partialMessages++

	case *ResultMessage:
		// Track result messages with errors
		if m.IsError {
			v.errors++
		}
	}
}

// TrackError records an error encountered during stream processing.
func (v *StreamValidator) TrackError() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.errors++
}

// TrackToolRequest records a tool use request by ID.
func (v *StreamValidator) TrackToolRequest(toolUseID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.toolsRequested[toolUseID] = true
}

// TrackToolResult records a tool result by ID.
func (v *StreamValidator) TrackToolResult(toolUseID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.toolsReceived[toolUseID] = true
}

// MarkStreamEnd marks the stream as ended.
// Call this when the stream closes normally.
func (v *StreamValidator) MarkStreamEnd() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.streamEnded = true
}

// GetIssues returns validation issues detected in the stream.
// Call after MarkStreamEnd for complete results.
func (v *StreamValidator) GetIssues() []StreamIssue {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var issues []StreamIssue

	// Check for tool requests without responses
	for toolID := range v.toolsRequested {
		if !v.toolsReceived[toolID] {
			issues = append(issues, StreamIssue{
				UUID:   toolID,
				Type:   "incomplete_tool",
				Detail: "tool request has no matching result",
			})
		}
	}

	// Check for orphaned tool results
	for toolID := range v.toolsReceived {
		if !v.toolsRequested[toolID] {
			issues = append(issues, StreamIssue{
				UUID:   toolID,
				Type:   "orphan_result",
				Detail: "tool result has no matching request",
			})
		}
	}

	// Check if stream ended without messages or tracked tools
	if v.streamEnded && v.totalMessages == 0 && len(v.toolsRequested) == 0 && len(v.toolsReceived) == 0 {
		issues = append(issues, StreamIssue{
			Type:   "empty_stream",
			Detail: "stream ended without any messages",
		})
	}

	return issues
}

// GetStats returns statistics about the stream.
func (v *StreamValidator) GetStats() StreamStats {
	v.mu.RLock()
	defer v.mu.RUnlock()

	return StreamStats{
		TotalMessages:   v.totalMessages,
		PartialMessages: v.partialMessages,
		Errors:          v.errors,
		ProcessingTime:  time.Since(v.startTime).String(),
	}
}

// Reset resets the validator state for reuse.
func (v *StreamValidator) Reset() {
	v.mu.Lock()
	defer v.mu.Unlock()

	v.toolsRequested = make(map[string]bool)
	v.toolsReceived = make(map[string]bool)
	v.totalMessages = 0
	v.partialMessages = 0
	v.errors = 0
	v.startTime = time.Now()
	v.streamEnded = false
}

// IsComplete returns true if the stream is complete and has no issues.
func (v *StreamValidator) IsComplete() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if !v.streamEnded {
		return false
	}

	// Check all tool requests have responses
	for toolID := range v.toolsRequested {
		if !v.toolsReceived[toolID] {
			return false
		}
	}

	return true
}

// PendingToolCount returns the number of tool requests awaiting results.
func (v *StreamValidator) PendingToolCount() int {
	v.mu.RLock()
	defer v.mu.RUnlock()

	count := 0
	for toolID := range v.toolsRequested {
		if !v.toolsReceived[toolID] {
			count++
		}
	}
	return count
}