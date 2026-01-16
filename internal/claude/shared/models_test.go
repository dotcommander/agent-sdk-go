package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveModelName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "opus resolves to full model ID",
			input: "opus",
			want:  "claude-opus-4-5-20251101",
		},
		{
			name:  "sonnet resolves to full model ID",
			input: "sonnet",
			want:  "claude-sonnet-4-5-20250929",
		},
		{
			name:  "haiku resolves to full model ID",
			input: "haiku",
			want:  "claude-3-5-haiku-20241022",
		},
		{
			name:  "full model ID passes through unchanged",
			input: "claude-sonnet-4-5-20250929",
			want:  "claude-sonnet-4-5-20250929",
		},
		{
			name:  "older model ID passes through unchanged",
			input: "claude-3-5-sonnet-20241022",
			want:  "claude-3-5-sonnet-20241022",
		},
		{
			name:  "unknown name passes through unchanged",
			input: "unknown-model",
			want:  "unknown-model",
		},
		{
			name:  "empty string passes through unchanged",
			input: "",
			want:  "",
		},
		{
			name:  "case sensitive - OPUS not resolved",
			input: "OPUS",
			want:  "OPUS",
		},
		{
			name:  "case sensitive - Sonnet not resolved",
			input: "Sonnet",
			want:  "Sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveModelName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidModelShortName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "opus is valid",
			input: "opus",
			want:  true,
		},
		{
			name:  "sonnet is valid",
			input: "sonnet",
			want:  true,
		},
		{
			name:  "haiku is valid",
			input: "haiku",
			want:  true,
		},
		{
			name:  "full model ID is not a short name",
			input: "claude-sonnet-4-5-20250929",
			want:  false,
		},
		{
			name:  "unknown is not valid",
			input: "unknown",
			want:  false,
		},
		{
			name:  "empty string is not valid",
			input: "",
			want:  false,
		},
		{
			name:  "case sensitive - OPUS is not valid",
			input: "OPUS",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidModelShortName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetModelShortNames(t *testing.T) {
	names := GetModelShortNames()

	assert.Len(t, names, 3)
	assert.Contains(t, names, "opus")
	assert.Contains(t, names, "sonnet")
	assert.Contains(t, names, "haiku")
}

func TestModelShortNameConstants(t *testing.T) {
	assert.Equal(t, "opus", ModelShortNameOpus)
	assert.Equal(t, "sonnet", ModelShortNameSonnet)
	assert.Equal(t, "haiku", ModelShortNameHaiku)
}
