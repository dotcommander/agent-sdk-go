// Package subprocess provides subprocess communication with the Claude CLI.
// This file handles MCP message routing for SDK MCP servers.
package subprocess

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/mcp"
)

// handleMcpMessageRequest routes MCP JSONRPC messages to SDK servers.
// Follows handleCanUseToolRequest pattern with panic recovery.
func (p *Protocol) handleMcpMessageRequest(ctx context.Context, requestID string, request map[string]any) error {
	serverName := getString(request, "server_name")
	if serverName == "" {
		return p.sendErrorResponse(ctx, requestID, "missing server_name")
	}

	message, _ := request["message"].(map[string]any)
	if message == nil {
		return p.sendErrorResponse(ctx, requestID, "missing message")
	}

	// Thread-safe server lookup
	p.mu.Lock()
	server, exists := p.sdkMcpServers[serverName]
	p.mu.Unlock()

	if !exists {
		return p.sendMcpErrorResponse(ctx, requestID, message, -32601,
			fmt.Sprintf("server '%s' not found", serverName))
	}

	// Route JSONRPC method with panic recovery
	var mcpResponse map[string]any
	var routeErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				routeErr = fmt.Errorf("MCP handler panicked: %v", r)
			}
		}()
		mcpResponse = server.HandleRequest(ctx, message)
	}()

	if routeErr != nil {
		return p.sendMcpErrorResponse(ctx, requestID, message, -32603, routeErr.Error())
	}

	return p.sendMcpResponse(ctx, requestID, mcpResponse)
}

// sendMcpResponse sends an MCP success response.
func (p *Protocol) sendMcpResponse(ctx context.Context, requestID string, mcpResp map[string]any) error {
	response := SDKControlResponse{
		Type: MessageTypeControlResponse,
		Response: ControlResponse{
			Subtype:   ResponseSubtypeSuccess,
			RequestID: requestID,
			Response:  map[string]any{"mcp_response": mcpResp},
		},
	}
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("marshal MCP response: %w", err)
	}
	return p.transport.Write(ctx, append(data, '\n'))
}

// sendMcpErrorResponse sends an MCP JSONRPC error response.
func (p *Protocol) sendMcpErrorResponse(ctx context.Context, requestID string, msg map[string]any, code int, message string) error {
	errorResp := map[string]any{
		"jsonrpc": "2.0",
		"id":      msg["id"],
		"error": map[string]any{
			"code":    code,
			"message": message,
		},
	}
	return p.sendMcpResponse(ctx, requestID, errorResp)
}

// McpServerInstance is an interface that abstracts MCP server handling.
// This allows both mcp.SdkMcpServer and custom implementations to be used.
type McpServerInstance interface {
	HandleRequest(ctx context.Context, message map[string]any) map[string]any
}

// Ensure SdkMcpServer implements McpServerInstance
var _ McpServerInstance = (*mcp.SdkMcpServer)(nil)
