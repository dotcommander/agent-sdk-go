package cli

import (
	"strings"
	"testing"
)

func TestGetDefaultCommand(t *testing.T) {
	command := GetDefaultCommand()
	if command == "" {
		t.Error("Default command should not be empty")
	}
}

func TestGetDefaultPath(t *testing.T) {
	path := GetDefaultPath()
	if path == "" {
		t.Error("Default path should not be empty")
	}
}

func TestGetCommonPaths(t *testing.T) {
	paths := GetCommonPaths()
	if len(paths) == 0 {
		t.Error("Common paths should not be empty")
	}
}

func TestGetSuggestedCommands(t *testing.T) {
	commands := GetSuggestedCommands()
	if len(commands) == 0 {
		t.Error("Suggested commands should not be empty")
	}

	// Verify that suggested commands contain installation instructions
	hasInstallCommand := false
	for _, cmd := range commands {
		if contains(cmd, "install") || contains(cmd, "brew") || contains(cmd, "winget") {
			hasInstallCommand = true
			break
		}
	}

	if !hasInstallCommand {
		t.Error("Suggested commands should contain installation instructions")
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty", "", ""},
		{"Simple", "/path/to/file", "/path/to/file"},
		{"Home", "~/.config/file", ""},
		{"Env", "$HOME/file", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExpandPath(tt.input)
			if tt.expected != "" {
				if result != tt.expected {
					t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, result, tt.expected)
				}
			}
		})
	}
}

// TestIsCLIAvailable tests if Claude CLI is available (skip if not installed)
func TestIsCLIAvailable(t *testing.T) {
	available := IsCLIAvailable()
	// This test just logs the result - don't fail if CLI is not available
	t.Logf("Claude CLI available: %v", available)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}