package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMcpStdioServerConfig(t *testing.T) {
	config := McpStdioServerConfig{
		Type:    "stdio",
		Command: "node",
		Args:    []string{"server.js"},
		Env:     map[string]string{"DEBUG": "true"},
	}
	assert.Equal(t, "stdio", config.Type)
	assert.Equal(t, "node", config.Command)
	assert.Equal(t, []string{"server.js"}, config.Args)
	assert.Equal(t, "true", config.Env["DEBUG"])
}

func TestMcpSSEServerConfig(t *testing.T) {
	config := McpSSEServerConfig{
		Type:    "sse",
		URL:     "http://localhost:3000",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}
	assert.Equal(t, "sse", config.Type)
	assert.Equal(t, "http://localhost:3000", config.URL)
	assert.Equal(t, "Bearer token", config.Headers["Authorization"])
}

func TestMcpServerStatus(t *testing.T) {
	status := McpServerStatus{
		Name:   "test-server",
		Status: "connected",
		ServerInfo: &McpServerInfo{
			Name:    "Test Server",
			Version: "1.0.0",
		},
	}
	assert.Equal(t, "test-server", status.Name)
	assert.Equal(t, "connected", status.Status)
	assert.Equal(t, "1.0.0", status.ServerInfo.Version)
}
