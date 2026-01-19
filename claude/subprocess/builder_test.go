package subprocess

import (
	"strings"
	"testing"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
	"github.com/stretchr/testify/assert"
)

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name               string
		model              string
		systemPrompt       string
		customArgs         []string
		tools              *shared.ToolsConfig
		mcpServers         map[string]shared.McpServerConfig
		promptArg          *string
		wantContains       []string
		wantNotContains    []string
		wantEndsWithPrompt bool
	}{
		{
			name:  "interactive mode basic",
			model: "claude-sonnet-4-5-20250929",
			wantContains: []string{
				"--output-format", "stream-json",
				"--input-format", "stream-json",
				"--model", "claude-sonnet-4-5-20250929",
			},
			wantNotContains: []string{"-p", "--verbose"},
		},
		{
			name:      "one-shot mode basic",
			model:     "claude-sonnet-4-5-20250929",
			promptArg: strPtr("What is 2+2?"),
			wantContains: []string{
				"-p", "--verbose",
				"--output-format", "stream-json",
				"--model", "claude-sonnet-4-5-20250929",
			},
			wantNotContains:    []string{"--input-format"},
			wantEndsWithPrompt: true,
		},
		{
			name:         "with system prompt",
			model:        "claude-sonnet-4-5-20250929",
			systemPrompt: "You are a helpful assistant",
			wantContains: []string{
				"--system-prompt", "You are a helpful assistant",
			},
		},
		{
			name:       "with custom args",
			model:      "claude-sonnet-4-5-20250929",
			customArgs: []string{"--debug", "--max-tokens", "1000"},
			wantContains: []string{
				"--debug", "--max-tokens", "1000",
			},
		},
		{
			name:  "with tools preset",
			model: "claude-sonnet-4-5-20250929",
			tools: &shared.ToolsConfig{
				Type:   "preset",
				Preset: "claude_code",
			},
			wantContains: []string{
				"--tools=preset:claude_code",
			},
		},
		{
			name:  "with tools explicit list",
			model: "claude-sonnet-4-5-20250929",
			tools: &shared.ToolsConfig{
				Type:  "explicit",
				Tools: []string{"Read", "Write", "Bash"},
			},
			wantContains: []string{
				"--allowed-tools=Read,Write,Bash",
			},
		},
		{
			name:  "with empty tools preset (no arg added)",
			model: "claude-sonnet-4-5-20250929",
			tools: &shared.ToolsConfig{
				Type:   "preset",
				Preset: "",
			},
			wantNotContains: []string{"--tools=preset:"},
		},
		{
			name:  "with empty tools list (no arg added)",
			model: "claude-sonnet-4-5-20250929",
			tools: &shared.ToolsConfig{
				Type:  "explicit",
				Tools: []string{},
			},
			wantNotContains: []string{"--allowed-tools="},
		},
		{
			name:  "with MCP servers",
			model: "claude-sonnet-4-5-20250929",
			mcpServers: map[string]shared.McpServerConfig{
				"my-server": shared.McpStdioServerConfig{
					Command: "node",
					Args:    []string{"server.js"},
				},
			},
			wantContains: []string{
				"--mcp-config",
			},
		},
		{
			name:         "full configuration interactive",
			model:        "claude-opus-4-20250514",
			systemPrompt: "Be concise",
			customArgs:   []string{"--no-cache"},
			tools: &shared.ToolsConfig{
				Type:  "explicit",
				Tools: []string{"Read", "Grep"},
			},
			wantContains: []string{
				"--output-format", "stream-json",
				"--input-format", "stream-json",
				"--model", "claude-opus-4-20250514",
				"--system-prompt", "Be concise",
				"--no-cache",
				"--allowed-tools=Read,Grep",
			},
		},
		{
			name:         "full configuration one-shot",
			model:        "claude-opus-4-20250514",
			systemPrompt: "Be brief",
			customArgs:   []string{"--timeout", "30"},
			tools: &shared.ToolsConfig{
				Type:   "preset",
				Preset: "claude_code",
			},
			promptArg: strPtr("Explain quantum computing"),
			wantContains: []string{
				"-p", "--verbose",
				"--output-format", "stream-json",
				"--model", "claude-opus-4-20250514",
				"--system-prompt", "Be brief",
				"--timeout", "30",
				"--tools=preset:claude_code",
			},
			wantEndsWithPrompt: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := BuildArgs(tt.model, tt.systemPrompt, tt.customArgs, tt.tools, tt.mcpServers, tt.promptArg)

			// Check expected args are present
			for _, want := range tt.wantContains {
				assert.Contains(t, args, want, "expected arg %q not found in %v", want, args)
			}

			// Check unwanted args are absent
			for _, notWant := range tt.wantNotContains {
				found := false
				for _, arg := range args {
					if strings.Contains(arg, notWant) {
						found = true
						break
					}
				}
				assert.False(t, found, "unexpected arg containing %q found in %v", notWant, args)
			}

			// Check prompt is last in one-shot mode
			if tt.wantEndsWithPrompt && tt.promptArg != nil {
				assert.Equal(t, *tt.promptArg, args[len(args)-1], "prompt should be last argument")
			}
		})
	}
}

func TestBuildArgs_ModelPosition(t *testing.T) {
	args := BuildArgs("test-model", "", nil, nil, nil, nil)

	// Find --model flag and verify next arg is the model name
	for i, arg := range args {
		if arg == "--model" {
			assert.Less(t, i+1, len(args), "--model flag should have a value")
			assert.Equal(t, "test-model", args[i+1])
			return
		}
	}
	t.Fatal("--model flag not found")
}

func TestBuildArgs_InteractiveModeFlags(t *testing.T) {
	args := BuildArgs("model", "", nil, nil, nil, nil)

	// Interactive mode should have both input and output format flags
	hasOutputFormat := false
	hasInputFormat := false

	for i, arg := range args {
		if arg == "--output-format" && i+1 < len(args) && args[i+1] == "stream-json" {
			hasOutputFormat = true
		}
		if arg == "--input-format" && i+1 < len(args) && args[i+1] == "stream-json" {
			hasInputFormat = true
		}
	}

	assert.True(t, hasOutputFormat, "interactive mode should have --output-format stream-json")
	assert.True(t, hasInputFormat, "interactive mode should have --input-format stream-json")
}

func TestBuildArgs_OneShotModeFlags(t *testing.T) {
	prompt := "test prompt"
	args := BuildArgs("model", "", nil, nil, nil, &prompt)

	// One-shot mode should have -p, --verbose, and output format
	assert.Contains(t, args, "-p", "one-shot mode should have -p flag")
	assert.Contains(t, args, "--verbose", "one-shot mode should have --verbose flag")

	// Should NOT have input-format
	assert.NotContains(t, args, "--input-format", "one-shot mode should not have --input-format")
}

func TestBuildArgs_CustomArgsOrder(t *testing.T) {
	customArgs := []string{"--custom1", "val1", "--custom2"}
	args := BuildArgs("model", "", customArgs, nil, nil, nil)

	// Custom args should appear in order
	custom1Idx := -1
	custom2Idx := -1
	for i, arg := range args {
		if arg == "--custom1" {
			custom1Idx = i
		}
		if arg == "--custom2" {
			custom2Idx = i
		}
	}

	assert.NotEqual(t, -1, custom1Idx, "--custom1 should be present")
	assert.NotEqual(t, -1, custom2Idx, "--custom2 should be present")
	assert.Less(t, custom1Idx, custom2Idx, "custom args should maintain order")
}

func TestBuildArgs_MCPServerJSON(t *testing.T) {
	mcpServers := map[string]shared.McpServerConfig{
		"test-server": shared.McpStdioServerConfig{
			Command: "node",
			Args:    []string{"index.js"},
		},
	}

	args := BuildArgs("model", "", nil, nil, mcpServers, nil)

	// Find --mcp-config and verify it's followed by JSON
	for i, arg := range args {
		if arg == "--mcp-config" {
			assert.Less(t, i+1, len(args), "--mcp-config should have a value")
			jsonArg := args[i+1]
			assert.Contains(t, jsonArg, "mcpServers", "MCP config should contain mcpServers key")
			return
		}
	}
	t.Fatal("--mcp-config flag not found")
}

func TestBuildArgs_EmptyMCPServers(t *testing.T) {
	// Empty map should not add --mcp-config
	args := BuildArgs("model", "", nil, nil, map[string]shared.McpServerConfig{}, nil)
	assert.NotContains(t, args, "--mcp-config", "empty MCP servers should not add --mcp-config")
}

func TestBuildArgs_NilMCPServers(t *testing.T) {
	// Nil map should not add --mcp-config
	args := BuildArgs("model", "", nil, nil, nil, nil)
	assert.NotContains(t, args, "--mcp-config", "nil MCP servers should not add --mcp-config")
}
