package shared

import (
	"errors"
	"strings"
	"testing"
)

func TestBaseErrorFormatting(t *testing.T) {
	tests := []struct {
		name       string
		base       BaseError
		wantReason string
		wantInner  string
	}{
		{
			name:       "empty base",
			base:       BaseError{},
			wantReason: "",
			wantInner:  "",
		},
		{
			name:       "with reason",
			base:       BaseError{Reason: "something failed"},
			wantReason: ": something failed",
			wantInner:  "",
		},
		{
			name:       "with inner error",
			base:       BaseError{Inner: errors.New("inner error")},
			wantReason: "",
			wantInner:  ": inner error",
		},
		{
			name:       "with both",
			base:       BaseError{Reason: "operation failed", Inner: errors.New("inner")},
			wantReason: ": operation failed",
			wantInner:  ": inner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bReason, bInner strings.Builder
			tt.base.FormatReason(&bReason)
			tt.base.FormatInner(&bInner)

			if bReason.String() != tt.wantReason {
				t.Errorf("FormatReason() = %q, want %q", bReason.String(), tt.wantReason)
			}
			if bInner.String() != tt.wantInner {
				t.Errorf("FormatInner() = %q, want %q", bInner.String(), tt.wantInner)
			}
		})
	}
}

func TestBaseErrorUnwrap(t *testing.T) {
	inner := errors.New("inner error")
	base := BaseError{Inner: inner}

	if base.Unwrap() != inner {
		t.Error("Unwrap() should return inner error")
	}

	baseEmpty := BaseError{}
	if baseEmpty.Unwrap() != nil {
		t.Error("Unwrap() should return nil for empty inner")
	}
}

func TestConnectionError(t *testing.T) {
	inner := errors.New("connection refused")
	err := NewConnectionError("timeout connecting", inner)

	if !strings.Contains(err.Error(), "failed to connect to Claude CLI") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(err.Error(), "timeout connecting") {
		t.Error("Error message should contain reason")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Error("Error message should contain inner error")
	}

	// Test error unwrapping
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find inner error")
	}
}

func TestTimeoutError(t *testing.T) {
	err := NewTimeoutError("query", "30s")

	if !strings.Contains(err.Error(), "query timed out after 30s") {
		t.Errorf("Error message incorrect: %s", err.Error())
	}
}

func TestParserError(t *testing.T) {
	err := NewParserError(10, 5, `{"invalid": json}`, "unexpected token")

	msg := err.Error()
	if !strings.Contains(msg, "failed to parse JSON") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, "unexpected token") {
		t.Error("Error message should contain reason")
	}
	if !strings.Contains(msg, "line 10") {
		t.Error("Error message should contain line number")
	}
	if !strings.Contains(msg, "offset 5") {
		t.Error("Error message should contain offset")
	}
}

func TestParserErrorTruncatesLongData(t *testing.T) {
	longData := strings.Repeat("x", 200)
	err := NewParserError(1, 0, longData, "test")

	msg := err.Error()
	if len(msg) > 300 {
		t.Error("Error message should truncate long data")
	}
	if !strings.Contains(msg, "...") {
		t.Error("Error message should indicate truncation")
	}
}

func TestProtocolError(t *testing.T) {
	err := NewProtocolError("assistant", "invalid content block")

	msg := err.Error()
	if !strings.Contains(msg, "protocol error") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, `type="assistant"`) {
		t.Error("Error message should contain message type")
	}
	if !strings.Contains(msg, "invalid content block") {
		t.Error("Error message should contain reason")
	}
}

func TestConfigurationError(t *testing.T) {
	err := NewConfigurationError("Model", "invalid-model", "model must start with claude-")

	msg := err.Error()
	if !strings.Contains(msg, "invalid configuration") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, `field="Model"`) {
		t.Error("Error message should contain field")
	}
	if !strings.Contains(msg, `value="invalid-model"`) {
		t.Error("Error message should contain value")
	}
	if !strings.Contains(msg, "model must start with claude-") {
		t.Error("Error message should contain reason")
	}
}

func TestProcessError(t *testing.T) {
	err := NewProcessError(12345, "claude", "exit code 1", "SIGTERM")

	msg := err.Error()
	if !strings.Contains(msg, "process 12345 failed") {
		t.Error("Error message should contain PID")
	}
	if !strings.Contains(msg, "exit code 1") {
		t.Error("Error message should contain reason")
	}
	if !strings.Contains(msg, `command="claude"`) {
		t.Error("Error message should contain command")
	}
	if !strings.Contains(msg, "signal=SIGTERM") {
		t.Error("Error message should contain signal")
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		checker func(error) bool
		want    bool
	}{
		{
			name:    "ConnectionError positive",
			err:     NewConnectionError("test", nil),
			checker: IsConnectionError,
			want:    true,
		},
		{
			name:    "ConnectionError negative",
			err:     errors.New("random error"),
			checker: IsConnectionError,
			want:    false,
		},
		{
			name:    "TimeoutError positive",
			err:     NewTimeoutError("op", "5s"),
			checker: IsTimeoutError,
			want:    true,
		},
		{
			name:    "ParserError positive",
			err:     NewParserError(1, 0, "", "test"),
			checker: IsParserError,
			want:    true,
		},
		{
			name:    "ProtocolError positive",
			err:     NewProtocolError("", "test"),
			checker: IsProtocolError,
			want:    true,
		},
		{
			name:    "CLINotFoundError positive",
			err:     NewCLINotFoundError("/path", "claude"),
			checker: IsCLINotFound,
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checker(tt.err); got != tt.want {
				t.Errorf("checker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCLINotFoundError(t *testing.T) {
	err := NewCLINotFoundError("/usr/bin/claude", "claude")

	msg := err.Error()
	if !strings.Contains(msg, "Claude CLI not found") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, "/usr/bin/claude") {
		t.Error("Error message should contain path")
	}

	// Verify suggestions are included
	if len(err.Suggestions) == 0 {
		t.Error("CLINotFoundError should have suggestions")
	}
}

// Test errors.Is/errors.As compatibility
func TestErrorsIsAs(t *testing.T) {
	inner := errors.New("original error")
	connErr := NewConnectionError("test", inner)

	// errors.Is should work with inner error
	if !errors.Is(connErr, inner) {
		t.Error("errors.Is should find inner error")
	}

	// errors.As should work
	var target *ConnectionError
	if !errors.As(connErr, &target) {
		t.Error("errors.As should match ConnectionError")
	}
}

// TestSDKErrorInterface verifies all error types implement SDKError.
func TestSDKErrorInterface(t *testing.T) {
	tests := []struct {
		name     string
		err      SDKError
		wantType string
	}{
		{
			name:     "CLINotFoundError",
			err:      NewCLINotFoundError("/path", "cmd"),
			wantType: "cli_not_found",
		},
		{
			name:     "ConnectionError",
			err:      NewConnectionError("reason", nil),
			wantType: "connection",
		},
		{
			name:     "TimeoutError",
			err:      NewTimeoutError("op", "5s"),
			wantType: "timeout",
		},
		{
			name:     "ParserError",
			err:      NewParserError(1, 0, "", "reason"),
			wantType: "parser",
		},
		{
			name:     "ProtocolError",
			err:      NewProtocolError("type", "reason"),
			wantType: "protocol",
		},
		{
			name:     "ConfigurationError",
			err:      NewConfigurationError("field", "value", "reason"),
			wantType: "configuration",
		},
		{
			name:     "ProcessError",
			err:      NewProcessError(123, "cmd", "reason", ""),
			wantType: "process",
		},
		{
			name:     "JSONDecodeError",
			err:      NewJSONDecodeError(10, 5, "reason", nil),
			wantType: "json_decode",
		},
		{
			name:     "MessageParseError",
			err:      NewMessageParseError(nil, "type", "reason"),
			wantType: "message_parse",
		},
		{
			name:     "PermissionError",
			err:      NewPermissionError("Bash", "/path", "execute", "reason"),
			wantType: "permission",
		},
		{
			name:     "ModelError",
			err:      NewModelError("claude-invalid", "reason", nil),
			wantType: "model",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify Type() returns expected value
			if got := tt.err.Type(); got != tt.wantType {
				t.Errorf("Type() = %q, want %q", got, tt.wantType)
			}

			// Verify it implements error interface
			if tt.err.Error() == "" {
				t.Error("Error() should return non-empty string")
			}
		})
	}
}

func TestJSONDecodeError(t *testing.T) {
	inner := errors.New("unexpected EOF")
	err := NewJSONDecodeError(10, 25, "invalid JSON syntax", inner)

	msg := err.Error()
	if !strings.Contains(msg, "json decode error") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, "line=10") {
		t.Error("Error message should contain line number")
	}
	if !strings.Contains(msg, "position=25") {
		t.Error("Error message should contain position")
	}
	if !strings.Contains(msg, "invalid JSON syntax") {
		t.Error("Error message should contain reason")
	}

	// Test error unwrapping
	if !errors.Is(err, inner) {
		t.Error("errors.Is should find original error")
	}

	// Test Type() method
	if err.Type() != "json_decode" {
		t.Errorf("Type() = %q, want %q", err.Type(), "json_decode")
	}
}

func TestJSONDecodeErrorUnwrap(t *testing.T) {
	// With OriginalError
	origErr := errors.New("original")
	err := NewJSONDecodeError(1, 0, "test", origErr)
	if err.Unwrap() != origErr {
		t.Error("Unwrap() should return OriginalError")
	}

	// Without OriginalError but with Inner
	innerErr := errors.New("inner")
	err2 := &JSONDecodeError{
		BaseError: BaseError{Inner: innerErr},
	}
	if err2.Unwrap() != innerErr {
		t.Error("Unwrap() should fall back to BaseError.Inner")
	}

	// Without either
	err3 := NewJSONDecodeError(1, 0, "test", nil)
	if err3.Unwrap() != nil {
		t.Error("Unwrap() should return nil when no error")
	}
}

func TestMessageParseError(t *testing.T) {
	data := map[string]any{"invalid": "structure"}
	err := NewMessageParseError(data, "assistant", "missing content field")

	msg := err.Error()
	if !strings.Contains(msg, "message parse error") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, `type="assistant"`) {
		t.Error("Error message should contain message type")
	}
	if !strings.Contains(msg, "missing content field") {
		t.Error("Error message should contain reason")
	}

	// Verify stored data
	if err.Data == nil {
		t.Error("Data should be stored")
	}

	// Test Type() method
	if err.Type() != "message_parse" {
		t.Errorf("Type() = %q, want %q", err.Type(), "message_parse")
	}
}

func TestAsErrorHelpers(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		helper func(error) (any, bool)
		want   bool
	}{
		{
			name:   "AsConnectionError positive",
			err:    NewConnectionError("test", nil),
			helper: func(e error) (any, bool) { return AsConnectionError(e) },
			want:   true,
		},
		{
			name:   "AsConnectionError negative",
			err:    errors.New("random"),
			helper: func(e error) (any, bool) { return AsConnectionError(e) },
			want:   false,
		},
		{
			name:   "AsTimeoutError positive",
			err:    NewTimeoutError("op", "5s"),
			helper: func(e error) (any, bool) { return AsTimeoutError(e) },
			want:   true,
		},
		{
			name:   "AsParserError positive",
			err:    NewParserError(1, 0, "", "test"),
			helper: func(e error) (any, bool) { return AsParserError(e) },
			want:   true,
		},
		{
			name:   "AsProtocolError positive",
			err:    NewProtocolError("type", "test"),
			helper: func(e error) (any, bool) { return AsProtocolError(e) },
			want:   true,
		},
		{
			name:   "AsConfigurationError positive",
			err:    NewConfigurationError("f", "v", "r"),
			helper: func(e error) (any, bool) { return AsConfigurationError(e) },
			want:   true,
		},
		{
			name:   "AsProcessError positive",
			err:    NewProcessError(1, "cmd", "r", ""),
			helper: func(e error) (any, bool) { return AsProcessError(e) },
			want:   true,
		},
		{
			name:   "AsCLINotFoundError positive",
			err:    NewCLINotFoundError("path", "cmd"),
			helper: func(e error) (any, bool) { return AsCLINotFoundError(e) },
			want:   true,
		},
		{
			name:   "AsJSONDecodeError positive",
			err:    NewJSONDecodeError(1, 0, "r", nil),
			helper: func(e error) (any, bool) { return AsJSONDecodeError(e) },
			want:   true,
		},
		{
			name:   "AsJSONDecodeError negative",
			err:    errors.New("random"),
			helper: func(e error) (any, bool) { return AsJSONDecodeError(e) },
			want:   false,
		},
		{
			name:   "AsMessageParseError positive",
			err:    NewMessageParseError(nil, "t", "r"),
			helper: func(e error) (any, bool) { return AsMessageParseError(e) },
			want:   true,
		},
		{
			name:   "AsMessageParseError negative",
			err:    errors.New("random"),
			helper: func(e error) (any, bool) { return AsMessageParseError(e) },
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := tt.helper(tt.err)
			if got != tt.want {
				t.Errorf("helper returned %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAsErrorHelpersWithWrappedErrors(t *testing.T) {
	// Test that As* helpers work with wrapped errors
	inner := NewJSONDecodeError(1, 0, "test", nil)
	wrapped := errors.Join(errors.New("wrapper"), inner)

	extracted, ok := AsJSONDecodeError(wrapped)
	if !ok {
		t.Error("AsJSONDecodeError should find error in wrapped chain")
	}
	if extracted.Line != 1 {
		t.Error("Extracted error should have correct Line")
	}

	// Test with fmt.Errorf wrapping
	inner2 := NewMessageParseError(nil, "type", "reason")
	wrapped2 := errors.Join(errors.New("context"), inner2)

	extracted2, ok2 := AsMessageParseError(wrapped2)
	if !ok2 {
		t.Error("AsMessageParseError should find error in wrapped chain")
	}
	if extracted2.MessageType != "type" {
		t.Error("Extracted error should have correct MessageType")
	}
}

func TestIsNewErrorTypeHelpers(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		checker func(error) bool
		want    bool
	}{
		{
			name:    "IsJSONDecodeError positive",
			err:     NewJSONDecodeError(1, 0, "", nil),
			checker: IsJSONDecodeError,
			want:    true,
		},
		{
			name:    "IsJSONDecodeError negative",
			err:     errors.New("random"),
			checker: IsJSONDecodeError,
			want:    false,
		},
		{
			name:    "IsMessageParseError positive",
			err:     NewMessageParseError(nil, "", ""),
			checker: IsMessageParseError,
			want:    true,
		},
		{
			name:    "IsMessageParseError negative",
			err:     errors.New("random"),
			checker: IsMessageParseError,
			want:    false,
		},
		{
			name:    "IsPermissionError positive",
			err:     NewPermissionError("Bash", "", "execute", "command blocked"),
			checker: IsPermissionError,
			want:    true,
		},
		{
			name:    "IsPermissionError negative",
			err:     errors.New("random"),
			checker: IsPermissionError,
			want:    false,
		},
		{
			name:    "IsModelError positive",
			err:     NewModelError("claude-invalid", "model not found", nil),
			checker: IsModelError,
			want:    true,
		},
		{
			name:    "IsModelError negative",
			err:     errors.New("random"),
			checker: IsModelError,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.checker(tt.err); got != tt.want {
				t.Errorf("checker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionError(t *testing.T) {
	err := NewPermissionError("Bash", "/etc/passwd", "read", "file access denied")

	msg := err.Error()
	if !strings.Contains(msg, "permission denied") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, "Bash") {
		t.Error("Error message should contain tool name")
	}
	if !strings.Contains(msg, "/etc/passwd") {
		t.Error("Error message should contain path")
	}
	if !strings.Contains(msg, "read") {
		t.Error("Error message should contain operation")
	}
	if !strings.Contains(msg, "file access denied") {
		t.Error("Error message should contain reason")
	}

	// Test Type() method
	if err.Type() != "permission" {
		t.Errorf("Type() = %q, want %q", err.Type(), "permission")
	}
}

func TestModelError(t *testing.T) {
	supported := []string{"claude-sonnet-4", "claude-opus-4", "claude-haiku"}
	err := NewModelError("claude-invalid-99", "model not available", supported)

	msg := err.Error()
	if !strings.Contains(msg, "model error") {
		t.Error("Error message should contain base message")
	}
	if !strings.Contains(msg, "claude-invalid-99") {
		t.Error("Error message should contain model name")
	}
	if !strings.Contains(msg, "model not available") {
		t.Error("Error message should contain reason")
	}
	if !strings.Contains(msg, "claude-sonnet-4") {
		t.Error("Error message should contain supported models")
	}

	// Test Type() method
	if err.Type() != "model" {
		t.Errorf("Type() = %q, want %q", err.Type(), "model")
	}
}

func TestPermissionErrorAsExtraction(t *testing.T) {
	err := NewPermissionError("Write", "/root/file", "write", "denied")

	permErr, ok := AsPermissionError(err)
	if !ok {
		t.Error("AsPermissionError should return true for PermissionError")
	}
	if permErr.Tool != "Write" {
		t.Errorf("Tool = %q, want %q", permErr.Tool, "Write")
	}

	// Test with non-PermissionError
	_, ok = AsPermissionError(errors.New("random"))
	if ok {
		t.Error("AsPermissionError should return false for non-PermissionError")
	}
}

func TestModelErrorAsExtraction(t *testing.T) {
	err := NewModelError("claude-test", "unavailable", []string{"claude-sonnet-4"})

	modelErr, ok := AsModelError(err)
	if !ok {
		t.Error("AsModelError should return true for ModelError")
	}
	if modelErr.Model != "claude-test" {
		t.Errorf("Model = %q, want %q", modelErr.Model, "claude-test")
	}
	if len(modelErr.SupportedModels) != 1 {
		t.Errorf("SupportedModels length = %d, want 1", len(modelErr.SupportedModels))
	}

	// Test with non-ModelError
	_, ok = AsModelError(errors.New("random"))
	if ok {
		t.Error("AsModelError should return false for non-ModelError")
	}
}

// TestIsErrorTypeGeneric tests the generic IsErrorType helper directly.
func TestIsErrorTypeGeneric(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ConnectionError direct",
			err:  NewConnectionError("test", nil),
			want: true,
		},
		{
			name: "ConnectionError wrapped",
			err:  errors.Join(errors.New("wrapper"), NewConnectionError("test", nil)),
			want: true,
		},
		{
			name: "wrong type",
			err:  NewTimeoutError("op", "5s"),
			want: false,
		},
		{
			name: "plain error",
			err:  errors.New("random"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsErrorType[*ConnectionError](tt.err)
			if got != tt.want {
				t.Errorf("IsErrorType[*ConnectionError]() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAsErrorTypeGeneric tests the generic AsErrorType helper directly.
func TestAsErrorTypeGeneric(t *testing.T) {
	t.Run("direct match", func(t *testing.T) {
		orig := NewTimeoutError("query", "30s")
		extracted, ok := AsErrorType[*TimeoutError](orig)
		if !ok {
			t.Error("AsErrorType should return true for matching type")
		}
		if extracted.Operation != "query" {
			t.Errorf("Operation = %q, want %q", extracted.Operation, "query")
		}
		if extracted.Timeout != "30s" {
			t.Errorf("Timeout = %q, want %q", extracted.Timeout, "30s")
		}
	})

	t.Run("wrapped match", func(t *testing.T) {
		orig := NewParserError(42, 10, "bad json", "unexpected token")
		wrapped := errors.Join(errors.New("outer"), orig)
		extracted, ok := AsErrorType[*ParserError](wrapped)
		if !ok {
			t.Error("AsErrorType should find error in wrapped chain")
		}
		if extracted.Line != 42 {
			t.Errorf("Line = %d, want %d", extracted.Line, 42)
		}
	})

	t.Run("no match", func(t *testing.T) {
		err := errors.New("plain error")
		extracted, ok := AsErrorType[*ProcessError](err)
		if ok {
			t.Error("AsErrorType should return false for non-matching type")
		}
		if extracted != nil {
			t.Error("extracted should be nil when not found")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		extracted, ok := AsErrorType[*ConfigurationError](nil)
		if ok {
			t.Error("AsErrorType should return false for nil error")
		}
		if extracted != nil {
			t.Error("extracted should be nil for nil input")
		}
	})
}

// TestGenericHelpersMatchSpecificHelpers verifies generic helpers produce same results.
func TestGenericHelpersMatchSpecificHelpers(t *testing.T) {
	testCases := []error{
		NewCLINotFoundError("/path", "cmd"),
		NewConnectionError("reason", nil),
		NewTimeoutError("op", "5s"),
		NewParserError(1, 0, "", "reason"),
		NewProtocolError("type", "reason"),
		NewConfigurationError("field", "value", "reason"),
		NewProcessError(123, "cmd", "reason", ""),
		NewJSONDecodeError(1, 0, "reason", nil),
		NewMessageParseError(nil, "type", "reason"),
		NewPermissionError("tool", "path", "op", "reason"),
		NewModelError("model", "reason", nil),
		errors.New("plain error"),
		nil,
	}

	for _, err := range testCases {
		// Test Is* functions match
		if IsConnectionError(err) != IsErrorType[*ConnectionError](err) {
			t.Errorf("IsConnectionError mismatch for %T", err)
		}
		if IsTimeoutError(err) != IsErrorType[*TimeoutError](err) {
			t.Errorf("IsTimeoutError mismatch for %T", err)
		}
		if IsParserError(err) != IsErrorType[*ParserError](err) {
			t.Errorf("IsParserError mismatch for %T", err)
		}
		if IsProtocolError(err) != IsErrorType[*ProtocolError](err) {
			t.Errorf("IsProtocolError mismatch for %T", err)
		}
		if IsConfigurationError(err) != IsErrorType[*ConfigurationError](err) {
			t.Errorf("IsConfigurationError mismatch for %T", err)
		}
		if IsProcessError(err) != IsErrorType[*ProcessError](err) {
			t.Errorf("IsProcessError mismatch for %T", err)
		}
		if IsPermissionError(err) != IsErrorType[*PermissionError](err) {
			t.Errorf("IsPermissionError mismatch for %T", err)
		}
		if IsModelError(err) != IsErrorType[*ModelError](err) {
			t.Errorf("IsModelError mismatch for %T", err)
		}
		if IsJSONDecodeError(err) != IsErrorType[*JSONDecodeError](err) {
			t.Errorf("IsJSONDecodeError mismatch for %T", err)
		}
		if IsMessageParseError(err) != IsErrorType[*MessageParseError](err) {
			t.Errorf("IsMessageParseError mismatch for %T", err)
		}

		// Test As* functions match
		a1, ok1 := AsConnectionError(err)
		a2, ok2 := AsErrorType[*ConnectionError](err)
		if ok1 != ok2 || a1 != a2 {
			t.Errorf("AsConnectionError mismatch for %T", err)
		}

		b1, ok1 := AsTimeoutError(err)
		b2, ok2 := AsErrorType[*TimeoutError](err)
		if ok1 != ok2 || b1 != b2 {
			t.Errorf("AsTimeoutError mismatch for %T", err)
		}

		c1, ok1 := AsCLINotFoundError(err)
		c2, ok2 := AsErrorType[*CLINotFoundError](err)
		if ok1 != ok2 || c1 != c2 {
			t.Errorf("AsCLINotFoundError mismatch for %T", err)
		}
	}
}
