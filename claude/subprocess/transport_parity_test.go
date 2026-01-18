package subprocess

import (
	"testing"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransportConfig_CwdField tests the cwd configuration.
func TestTransportConfig_CwdField(t *testing.T) {
	config := &TransportConfig{
		Model: "claude-sonnet-4-5-20250929",
		Cwd:   "/custom/path",
	}

	transport, err := NewTransport(config)
	require.NoError(t, err)
	assert.Equal(t, "/custom/path", transport.cwd)
}

// TestTransportConfig_ToolsPreset tests the tools preset configuration.
func TestTransportConfig_ToolsPreset(t *testing.T) {
	config := &TransportConfig{
		Model: "claude-sonnet-4-5-20250929",
		Tools: shared.ToolsPreset("claude_code"),
	}

	transport, err := NewTransport(config)
	require.NoError(t, err)
	assert.NotNil(t, transport.tools)
	assert.Equal(t, "preset", transport.tools.Type)
	assert.Equal(t, "claude_code", transport.tools.Preset)
}

// TestTransportConfig_ToolsExplicit tests the explicit tools configuration.
func TestTransportConfig_ToolsExplicit(t *testing.T) {
	config := &TransportConfig{
		Model: "claude-sonnet-4-5-20250929",
		Tools: shared.ToolsExplicit("Read", "Write", "Bash"),
	}

	transport, err := NewTransport(config)
	require.NoError(t, err)
	assert.NotNil(t, transport.tools)
	assert.Equal(t, "explicit", transport.tools.Type)
	assert.Equal(t, []string{"Read", "Write", "Bash"}, transport.tools.Tools)
}

// TestTransportConfig_StderrCallback tests the stderr callback configuration.
func TestTransportConfig_StderrCallback(t *testing.T) {
	called := false
	callback := func(line string) {
		called = true
	}

	config := &TransportConfig{
		Model:          "claude-sonnet-4-5-20250929",
		StderrCallback: callback,
	}

	transport, err := NewTransport(config)
	require.NoError(t, err)
	assert.NotNil(t, transport.stderrCallback)

	// Test the callback
	transport.stderrCallback("test")
	assert.True(t, called)
}

// TestTransport_buildArgs_ToolsPreset tests CLI args generation for tools preset.
func TestTransport_buildArgs_ToolsPreset(t *testing.T) {
	transport := &Transport{
		promptArg: nil,
		model:     "claude-sonnet-4-5-20250929",
		tools:     shared.ToolsPreset("claude_code"),
	}

	args := transport.buildArgs()

	// Should contain --tools=preset:claude_code
	found := false
	for _, arg := range args {
		if arg == "--tools=preset:claude_code" {
			found = true
			break
		}
	}
	assert.True(t, found, "should contain --tools=preset:claude_code, got: %v", args)
}

// TestTransport_buildArgs_ToolsExplicit tests CLI args generation for explicit tools.
func TestTransport_buildArgs_ToolsExplicit(t *testing.T) {
	transport := &Transport{
		promptArg: nil,
		model:     "claude-sonnet-4-5-20250929",
		tools:     shared.ToolsExplicit("Read", "Write"),
	}

	args := transport.buildArgs()

	// Should contain --allowed-tools=Read,Write
	found := false
	for _, arg := range args {
		if arg == "--allowed-tools=Read,Write" {
			found = true
			break
		}
	}
	assert.True(t, found, "should contain --allowed-tools=Read,Write, got: %v", args)
}

// TestTransport_buildArgs_NoTools tests CLI args generation with no tools config.
func TestTransport_buildArgs_NoTools(t *testing.T) {
	transport := &Transport{
		promptArg: nil,
		model:     "claude-sonnet-4-5-20250929",
		tools:     nil,
	}

	args := transport.buildArgs()

	// Should not contain any tools args
	for _, arg := range args {
		assert.NotContains(t, arg, "tools")
		assert.NotContains(t, arg, "allowed-tools")
	}
}

// TestTransportConfig_Defaults tests default values are set correctly.
func TestTransportConfig_Defaults(t *testing.T) {
	config := &TransportConfig{}

	transport, err := NewTransport(config)
	require.NoError(t, err)

	// Check defaults
	assert.Equal(t, "claude-sonnet-4-5-20250929", transport.model)
	assert.Equal(t, defaultTimeout, transport.timeout)
	assert.Equal(t, "", transport.cwd)
	assert.Nil(t, transport.tools)
	assert.Nil(t, transport.stderrCallback)
}

// TestTransportConfig_WithAllOptions tests transport with all new options set.
func TestTransportConfig_WithAllOptions(t *testing.T) {
	stderrCallback := func(line string) {}

	config := &TransportConfig{
		CLIPath:        "/custom/cli",
		CLICommand:     "claude-custom",
		Model:          "claude-3-opus",
		Timeout:        120 * time.Second,
		SystemPrompt:   "You are helpful",
		CustomArgs:     []string{"--debug"},
		Env:            map[string]string{"CUSTOM_VAR": "value"},
		Cwd:            "/project",
		Tools:          shared.ToolsPreset("claude_code"),
		StderrCallback: stderrCallback,
	}

	transport, err := NewTransport(config)
	require.NoError(t, err)

	assert.Equal(t, "/custom/cli", transport.cliPath)
	assert.Equal(t, "claude-custom", transport.cliCommand)
	assert.Equal(t, "claude-3-opus", transport.model)
	assert.Equal(t, 120*time.Second, transport.timeout)
	assert.Equal(t, "You are helpful", transport.systemPrompt)
	assert.Equal(t, []string{"--debug"}, transport.customArgs)
	assert.Equal(t, map[string]string{"CUSTOM_VAR": "value"}, transport.env)
	assert.Equal(t, "/project", transport.cwd)
	assert.NotNil(t, transport.tools)
	assert.NotNil(t, transport.stderrCallback)
}

// TestNewTransportWithPrompt_ToolsPreset tests one-shot mode with tools preset.
func TestNewTransportWithPrompt_ToolsPreset(t *testing.T) {
	config := &TransportConfig{
		Model: "claude-sonnet-4-5-20250929",
		Tools: shared.ToolsPreset("claude_code"),
	}

	transport, err := NewTransportWithPrompt(config, "Hello")
	require.NoError(t, err)
	assert.NotNil(t, transport.promptArg)
	assert.NotNil(t, transport.tools)

	// Check args include tools preset
	args := transport.buildArgs()
	found := false
	for _, arg := range args {
		if arg == "--tools=preset:claude_code" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

// TestTransport_buildArgs_Order tests that args are built in correct order.
func TestTransport_buildArgs_Order(t *testing.T) {
	transport := &Transport{
		promptArg:    strPtr("test prompt"),
		model:        "claude-sonnet-4-5-20250929",
		systemPrompt: "Be helpful",
		customArgs:   []string{"--debug"},
		tools:        shared.ToolsPreset("claude_code"),
	}

	args := transport.buildArgs()

	// Prompt should be last in one-shot mode
	assert.Equal(t, "test prompt", args[len(args)-1])

	// Model should appear
	modelIdx := -1
	for i, arg := range args {
		if arg == "--model" {
			modelIdx = i
			break
		}
	}
	assert.NotEqual(t, -1, modelIdx)
	assert.Equal(t, "claude-sonnet-4-5-20250929", args[modelIdx+1])
}
