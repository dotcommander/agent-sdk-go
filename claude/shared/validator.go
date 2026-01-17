package shared

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
func ValidateModel(model string) error {
	if strings.TrimSpace(model) == "" {
		return errors.New("model name cannot be empty")
	}

	if !strings.HasPrefix(model, "claude-") {
		return errors.New("model name must start with 'claude-'")
	}

	// Additional model validation could be added here
	// For now, just check the prefix

	return nil
}

// ValidatePermissionMode validates a permission mode.
func ValidatePermissionMode(mode string) error {
	validModes := []string{"auto", "read", "write", "restricted"}
	for _, valid := range validModes {
		if mode == valid {
			return nil
		}
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
func SanitizeMessage(msg Message) interface{} {
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
func sanitizeUserMessage(msg *UserMessage) map[string]interface{} {
	sanitized := map[string]interface{}{
		"type": msg.Type(),
	}

	switch content := msg.Content.(type) {
	case string:
		sanitized["content"] = "<redacted>" // Redact user content for logging
	case []ContentBlock:
		sanitizedBlocks := make([]interface{}, len(content))
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

func sanitizeAssistantMessage(msg *AssistantMessage) map[string]interface{} {
	sanitized := map[string]interface{}{
		"type":  msg.Type(),
		"model": msg.Model,
	}

	if msg.Error != nil {
		sanitized["error"] = string(*msg.Error)
	}

	sanitizedBlocks := make([]interface{}, len(msg.Content))
	for i, block := range msg.Content {
		sanitizedBlocks[i] = sanitizeContentBlock(block)
	}
	sanitized["content"] = sanitizedBlocks

	return sanitized
}

func sanitizeSystemMessage(msg *SystemMessage) map[string]interface{} {
	return map[string]interface{}{
		"type":    msg.Type(),
		"subtype": msg.Subtype,
	}
}

func sanitizeResultMessage(msg *ResultMessage) map[string]interface{} {
	sanitized := map[string]interface{}{
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

func sanitizeStreamEvent(msg *StreamEvent) map[string]interface{} {
	sanitized := map[string]interface{}{
		"type":  msg.Type(),
		"uuid":  msg.UUID,
		"event": "<redacted>", // Redact event content for logging
	}

	if msg.ParentToolUseID != nil {
		sanitized["parent_tool_use_id"] = *msg.ParentToolUseID
	}

	return sanitized
}

func sanitizeContentBlock(block ContentBlock) interface{} {
	switch b := block.(type) {
	case *TextBlock:
		return map[string]interface{}{
			"type": "text",
			"text": "<redacted>", // Redact text content for logging
		}
	case *ThinkingBlock:
		return map[string]interface{}{
			"type":      "thinking",
			"thinking":  "<redacted>", // Redact thinking content for logging
			"signature": b.Signature,
		}
	case *ToolUseBlock:
		return map[string]interface{}{
			"type":       "tool_use",
			"tool_use_id": b.ToolUseID,
			"name":       b.Name,
			"input":      "<redacted>", // Redact input for logging
		}
	case *ToolResultBlock:
		return map[string]interface{}{
			"type":       "tool_result",
			"tool_use_id": b.ToolUseID,
			"is_error":   b.IsError,
		}
	default:
		return fmt.Sprintf("unknown content block type: %T", block)
	}
}