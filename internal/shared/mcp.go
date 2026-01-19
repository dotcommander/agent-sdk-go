package shared

import (
	"encoding/json"
)

// McpServerConfig is the interface for MCP server configurations.
type McpServerConfig interface {
	mcpServerConfig()
}

// McpStdioServerConfig configures an MCP server using stdio.
type McpStdioServerConfig struct {
	Type    string            `json:"type,omitempty"` // "stdio" (default)
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

func (McpStdioServerConfig) mcpServerConfig() {}

// McpSSEServerConfig configures an MCP server using Server-Sent Events.
type McpSSEServerConfig struct {
	Type    string            `json:"type"` // "sse"
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (McpSSEServerConfig) mcpServerConfig() {}

// McpHttpServerConfig configures an MCP server using HTTP.
type McpHttpServerConfig struct {
	Type    string            `json:"type"` // "http"
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
}

func (McpHttpServerConfig) mcpServerConfig() {}

// McpSdkServerConfig configures an in-process MCP server.
type McpSdkServerConfig struct {
	Type     string      `json:"type"` // "sdk"
	Name     string      `json:"name"`
	Instance any `json:"-"` // MCP Server instance (not serialized)
}

func (McpSdkServerConfig) mcpServerConfig() {}

// MarshalJSON allows JSON serialization (excluding Instance field).
func (c McpSdkServerConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"type": c.Type,
		"name": c.Name,
	})
}

// McpServerStatus represents the connection status of an MCP server.
type McpServerStatus struct {
	Name       string         `json:"name"`
	Status     string         `json:"status"` // "connected" | "failed" | "needs-auth" | "pending"
	ServerInfo *McpServerInfo `json:"serverInfo,omitempty"`
	Error      string         `json:"error,omitempty"`
}

// McpServerInfo contains information about a connected MCP server.
type McpServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// McpSetServersResult is the result of a setMcpServers operation.
type McpSetServersResult struct {
	Added   []string          `json:"added"`
	Removed []string          `json:"removed"`
	Errors  map[string]string `json:"errors"`
}
