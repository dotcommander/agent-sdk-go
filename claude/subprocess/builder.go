// Package subprocess provides subprocess communication with the Claude CLI.
package subprocess

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

// BuildArgs constructs CLI arguments from transport configuration.
// This is a pure function for easy testing without Transport instantiation.
//
// Parameters:
//   - model: the Claude model to use (required)
//   - systemPrompt: optional system prompt (empty string means no system prompt)
//   - customArgs: additional CLI arguments to include
//   - tools: optional tool configuration (preset or explicit list)
//   - mcpServers: optional MCP server configurations
//   - promptArg: if non-nil, enables one-shot mode with this prompt as positional argument
//
// Returns the slice of CLI arguments ready for exec.Command.
func BuildArgs(model, systemPrompt string, customArgs []string, tools *shared.ToolsConfig, mcpServers map[string]shared.McpServerConfig, promptArg *string) []string {
	var args []string

	if promptArg != nil {
		// One-shot mode: -p flag enables print mode, prompt is positional arg at end
		// --verbose is required for stream-json output in print mode
		args = append(args, "-p", "--output-format", "stream-json", "--verbose")
	} else {
		// Interactive mode: use streaming JSON for both input and output
		args = append(args, "--output-format", "stream-json", "--input-format", "stream-json")
	}

	// Add model
	args = append(args, "--model", model)

	// Add system prompt if set
	if systemPrompt != "" {
		args = append(args, "--system-prompt", systemPrompt)
	}

	// Add custom args
	args = append(args, customArgs...)

	// Add tools configuration
	if tools != nil {
		switch tools.Type {
		case "preset":
			// Preset: --tools=preset:claude_code
			if tools.Preset != "" {
				args = append(args, fmt.Sprintf("--tools=preset:%s", tools.Preset))
			}
		case "explicit":
			// Explicit list: --allowed-tools=Read,Write,Bash
			if len(tools.Tools) > 0 {
				args = append(args, fmt.Sprintf("--allowed-tools=%s", strings.Join(tools.Tools, ",")))
			}
		}
	}

	// MCP servers
	if len(mcpServers) > 0 {
		serversForCLI := make(map[string]any)
		for name, config := range mcpServers {
			if sdkConfig, ok := config.(shared.McpSdkServerConfig); ok {
				// For SDK servers, pass everything except instance
				serversForCLI[name] = map[string]any{
					"type": sdkConfig.Type,
					"name": sdkConfig.Name,
				}
			} else {
				serversForCLI[name] = config
			}
		}
		if len(serversForCLI) > 0 {
			mcpJSON, _ := json.Marshal(map[string]any{"mcpServers": serversForCLI})
			args = append(args, "--mcp-config", string(mcpJSON))
		}
	}

	// In one-shot mode, prompt goes last as positional argument
	if promptArg != nil {
		args = append(args, *promptArg)
	}

	return args
}
