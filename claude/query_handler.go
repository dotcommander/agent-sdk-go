// Package claude provides query handler utilities for MCP server management.
package claude

import (
	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

// extractSdkMcpServers extracts SDK MCP server instances from server configurations.
// This is used to separate SDK server instances (which cannot be JSON serialized)
// from the configuration that gets passed to the CLI.
func extractSdkMcpServers(servers map[string]shared.McpServerConfig) map[string]any {
	if servers == nil {
		return nil
	}

	sdkMcpServers := make(map[string]any)
	for name, config := range servers {
		if sdkConfig, ok := config.(shared.McpSdkServerConfig); ok {
			// Store the instance for runtime handling
			sdkMcpServers[name] = sdkConfig.Instance
		}
	}

	if len(sdkMcpServers) == 0 {
		return nil
	}
	return sdkMcpServers
}