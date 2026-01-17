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
