// Package claude provides MCP (Model Context Protocol) server support.
// This file implements SDK MCP servers for in-process tool hosting.
package claude

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

// McpToolHandler is the function signature for tool handlers.
// Context-first per Go idioms, explicit error return.
//
// Example:
//
//	handler := func(ctx context.Context, args map[string]any) (*McpToolResult, error) {
//	    a, _ := args["a"].(float64)
//	    b, _ := args["b"].(float64)
//	    return &McpToolResult{
//	        Content: []McpContent{{Type: "text", Text: fmt.Sprintf("%f", a+b)}},
//	    }, nil
//	}
type McpToolHandler func(ctx context.Context, args map[string]any) (*McpToolResult, error)

// McpToolResult represents the result of a tool call.
// Re-exported for convenience from the root claude package.
type McpToolResult struct {
	Content []McpContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

// McpContent represents content returned by a tool.
// Re-exported for convenience from the root claude package.
type McpContent struct {
	Type     string `json:"type"` // "text" or "image"
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`     // base64 for images
	MimeType string `json:"mimeType,omitempty"` // for images
}

// McpToolDefinition describes a tool exposed by an MCP server.
// Re-exported for convenience from the root claude package.
type McpToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// McpTool represents a tool for SDK MCP servers.
// This is the Go alternative to Python's @tool decorator.
//
// Create tools using NewTool() for proper initialization.
type McpTool struct {
	name        string
	description string
	inputSchema map[string]any
	handler     McpToolHandler
}

// NewTool creates a new MCP tool definition.
// This is the Go-idiomatic alternative to Python's @tool decorator.
//
// Example:
//
//	addTool := claude.NewTool(
//	    "add",
//	    "Add two numbers together",
//	    map[string]any{
//	        "type": "object",
//	        "properties": map[string]any{
//	            "a": map[string]any{"type": "number"},
//	            "b": map[string]any{"type": "number"},
//	        },
//	        "required": []string{"a", "b"},
//	    },
//	    func(ctx context.Context, args map[string]any) (*claude.McpToolResult, error) {
//	        a, _ := args["a"].(float64)
//	        b, _ := args["b"].(float64)
//	        return &claude.McpToolResult{
//	            Content: []claude.McpContent{
//	                {Type: "text", Text: fmt.Sprintf("%.2f + %.2f = %.2f", a, b, a+b)},
//	            },
//	        }, nil
//	    },
//	)
func NewTool(name, description string, inputSchema map[string]any, handler McpToolHandler) *McpTool {
	return &McpTool{
		name:        name,
		description: description,
		inputSchema: inputSchema,
		handler:     handler,
	}
}

// Name returns the tool's name.
func (t *McpTool) Name() string {
	return t.name
}

// Description returns the tool's description.
func (t *McpTool) Description() string {
	return t.description
}

// InputSchema returns the tool's input JSON schema.
func (t *McpTool) InputSchema() map[string]any {
	return t.inputSchema
}

// Call executes the tool handler with the given context and arguments.
// Returns an error if no handler is set.
func (t *McpTool) Call(ctx context.Context, args map[string]any) (*McpToolResult, error) {
	if t.handler == nil {
		return nil, fmt.Errorf("tool '%s' has no handler", t.name)
	}
	return t.handler(ctx, args)
}

// SdkMcpServer implements an in-process MCP server.
// It is thread-safe and can handle concurrent tool calls.
type SdkMcpServer struct {
	name    string
	version string
	mu      sync.RWMutex
	tools   map[string]*McpTool
}

// CreateSDKMcpServer creates an in-process MCP server with the given tools.
// This is the Go equivalent of Python's create_sdk_mcp_server().
//
// Example:
//
//	calculator := claude.CreateSDKMcpServer("calculator", "1.0.0", addTool, sqrtTool)
//
//	client, _ := claude.NewClient(
//	    claude.WithSdkMcpServer("calc", calculator),
//	    claude.WithAllowedTools("mcp__calc__add", "mcp__calc__sqrt"),
//	)
func CreateSDKMcpServer(name, version string, tools ...*McpTool) *shared.McpSdkServerConfig {
	server := &SdkMcpServer{
		name:    name,
		version: version,
		tools:   make(map[string]*McpTool),
	}
	for _, tool := range tools {
		if tool != nil {
			server.tools[tool.Name()] = tool
		}
	}
	return &shared.McpSdkServerConfig{
		Type:     "sdk",
		Name:     name,
		Instance: server,
	}
}

// Name returns the server name.
func (s *SdkMcpServer) Name() string {
	return s.name
}

// Version returns the server version.
func (s *SdkMcpServer) Version() string {
	return s.version
}

// ListTools returns all registered tools.
// This method is thread-safe.
func (s *SdkMcpServer) ListTools(_ context.Context) ([]McpToolDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	defs := make([]McpToolDefinition, 0, len(s.tools))
	for _, tool := range s.tools {
		defs = append(defs, McpToolDefinition{
			Name:        tool.Name(),
			Description: tool.Description(),
			InputSchema: tool.InputSchema(),
		})
	}
	return defs, nil
}

// CallTool executes a tool by name with the given arguments.
// Returns an error if the tool is not found.
// This method is thread-safe.
func (s *SdkMcpServer) CallTool(ctx context.Context, name string, args map[string]any) (*McpToolResult, error) {
	s.mu.RLock()
	tool, exists := s.tools[name]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tool '%s' not found", name)
	}

	return tool.Call(ctx, args)
}

// AddTool adds a tool to the server.
// This method is thread-safe.
func (s *SdkMcpServer) AddTool(tool *McpTool) {
	if tool == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name()] = tool
}

// RemoveTool removes a tool from the server.
// This method is thread-safe.
func (s *SdkMcpServer) RemoveTool(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tools[name]; exists {
		delete(s.tools, name)
		return true
	}
	return false
}

// WithSdkMcpServer adds an in-process SDK MCP server by name.
// This is a convenience method for adding SDK MCP servers created with CreateSDKMcpServer.
// Multiple calls accumulate servers.
//
// Example:
//
//	calculator := claude.CreateSDKMcpServer("calculator", "1.0.0", addTool, sqrtTool)
//	client, _ := claude.NewClient(
//	    claude.WithSdkMcpServer("calc", calculator),
//	    claude.WithAllowedTools("mcp__calc__add", "mcp__calc__sqrt"),
//	)
func WithSdkMcpServer(name string, server *shared.McpSdkServerConfig) ClientOption {
	return func(o *ClientOptions) {
		if o.McpServers == nil {
			o.McpServers = make(map[string]shared.McpServerConfig)
		}
		o.McpServers[name] = server
	}
}

// WithAllowedTools restricts Claude to only use the specified tools.
// When set, Claude can only use tools from this allowlist.
// Takes precedence over WithDisallowedTools if both are set.
//
// For SDK MCP tools, use the format: mcp__<server_name>__<tool_name>
//
// Example:
//
//	client, _ := claude.NewClient(
//	    claude.WithAllowedTools("Read", "Write", "Bash"),
//	)
//	// Claude can only use Read, Write, Bash - all other tools blocked
//
//	// With MCP tools:
//	client, _ := claude.NewClient(
//	    claude.WithSdkMcpServer("calc", calculator),
//	    claude.WithAllowedTools("mcp__calc__add", "mcp__calc__sqrt"),
//	)
func WithAllowedTools(tools ...string) ClientOption {
	return func(o *ClientOptions) {
		if len(tools) > 0 {
			o.CustomArgs = append(o.CustomArgs, "--allowed-tools", strings.Join(tools, ","))
		}
	}
}
