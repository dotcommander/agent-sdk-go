// Package mcp provides in-process MCP (Model Context Protocol) server support.
// This allows creating tools that run directly in the Go process without external servers.
package mcp

import (
	"context"
	"fmt"
	"reflect"

	"agent-sdk-go/claude/shared"
)

// SdkMcpTool represents a tool that can be provided to Claude.
type SdkMcpTool struct {
	Name        string
	Description string
	InputSchema any // Can be struct type, map, or JSON schema
	Handler     func(context.Context, map[string]any) (map[string]any, error)
}

// Tool creates a new SDK MCP tool.
//
// Example:
//
//	greet := Tool("greet", "Greet a user", map[string]string{"name": "string"},
//	    func(ctx context.Context, args map[string]any) (map[string]any, error) {
//	        name := args["name"].(string)
//	        return map[string]any{
//	            "content": []map[string]any{
//	                {"type": "text", "text": fmt.Sprintf("Hello, %s!", name)},
//	            },
//	        }, nil
//	    })
func Tool(
	name string,
	description string,
	inputSchema any,
	handler func(context.Context, map[string]any) (map[string]any, error),
) *SdkMcpTool {
	return &SdkMcpTool{
		Name:        name,
		Description: description,
		InputSchema: inputSchema,
		Handler:     handler,
	}
}

// SdkMcpServer represents an in-process MCP server.
type SdkMcpServer struct {
	Name    string
	Version string
	Tools   []*SdkMcpTool
	toolMap map[string]*SdkMcpTool
}

// CreateSdkMcpServer creates an in-process MCP server.
//
// Unlike external MCP servers that run as separate processes, SDK MCP servers
// run directly in your application's process. This provides:
//   - Better performance (no IPC overhead)
//   - Simpler deployment (single process)
//   - Easier debugging (same process)
//   - Direct access to your application's state
//
// Example:
//
//	server := CreateSdkMcpServer("my-tools", "1.0.0", []*SdkMcpTool{
//	    Tool("add", "Add numbers", map[string]string{"a": "number", "b": "number"},
//	        func(ctx context.Context, args map[string]any) (map[string]any, error) {
//	            a := args["a"].(float64)
//	            b := args["b"].(float64)
//	            return map[string]any{
//	                "content": []map[string]any{
//	                    {"type": "text", "text": fmt.Sprintf("Sum: %f", a+b)},
//	                },
//	            }, nil
//	        }),
//	})
//
//	options := &shared.ClaudeAgentOptions{
//	    McpServers: map[string]shared.McpServerConfig{
//	        "calc": server.ToConfig(),
//	    },
//	}
func CreateSdkMcpServer(name string, version string, tools []*SdkMcpTool) *SdkMcpServer {
	if version == "" {
		version = "1.0.0"
	}

	toolMap := make(map[string]*SdkMcpTool)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	return &SdkMcpServer{
		Name:    name,
		Version: version,
		Tools:   tools,
		toolMap: toolMap,
	}
}

// ToConfig converts the server to a shared.McpSdkServerConfig.
func (s *SdkMcpServer) ToConfig() shared.McpSdkServerConfig {
	return shared.McpSdkServerConfig{
		Type:     "sdk",
		Name:     s.Name,
		Instance: s,
	}
}

// HandleRequest handles MCP JSON-RPC requests.
func (s *SdkMcpServer) HandleRequest(ctx context.Context, message map[string]any) map[string]any {
	method, _ := message["method"].(string)
	params, _ := message["params"].(map[string]any)
	msgID := message["id"]

	switch method {
	case "initialize":
		return s.handleInitialize(msgID)
	case "tools/list":
		return s.handleListTools(msgID)
	case "tools/call":
		return s.handleCallTool(ctx, msgID, params)
	case "notifications/initialized":
		// Just acknowledge
		return map[string]any{
			"jsonrpc": "2.0",
			"result":  map[string]any{},
		}
	default:
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"error": map[string]any{
				"code":    -32601,
				"message": fmt.Sprintf("Method '%s' not found", method),
			},
		}
	}
}

func (s *SdkMcpServer) handleInitialize(msgID any) map[string]any {
	return map[string]any{
		"jsonrpc": "2.0",
		"id":      msgID,
		"result": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]any{
				"tools": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    s.Name,
				"version": s.Version,
			},
		},
	}
}

func (s *SdkMcpServer) handleListTools(msgID any) map[string]any {
	tools := make([]map[string]any, len(s.Tools))
	for i, tool := range s.Tools {
		tools[i] = map[string]any{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": s.convertSchema(tool.InputSchema),
		}
	}

	return map[string]any{
		"jsonrpc": "2.0",
		"id":      msgID,
		"result": map[string]any{
			"tools": tools,
		},
	}
}

func (s *SdkMcpServer) handleCallTool(ctx context.Context, msgID any, params map[string]any) map[string]any {
	toolName, _ := params["name"].(string)
	arguments, _ := params["arguments"].(map[string]any)

	tool, exists := s.toolMap[toolName]
	if !exists {
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"error": map[string]any{
				"code":    -32602,
				"message": fmt.Sprintf("Tool '%s' not found", toolName),
			},
		}
	}

	// Call handler
	result, err := tool.Handler(ctx, arguments)
	if err != nil {
		return map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"result": map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": fmt.Sprintf("Error: %v", err)},
				},
				"isError": true,
			},
		}
	}

	return map[string]any{
		"jsonrpc": "2.0",
		"id":      msgID,
		"result":  result,
	}
}

// convertSchema converts various schema formats to JSON Schema.
func (s *SdkMcpServer) convertSchema(schema any) map[string]any {
	// Handle map[string]string (simple type map)
	if schemaMap, ok := schema.(map[string]string); ok {
		properties := make(map[string]any)
		required := make([]string, 0, len(schemaMap))
		for name, typeStr := range schemaMap {
			required = append(required, name)
			properties[name] = s.convertTypeToSchema(typeStr)
		}

		return map[string]any{
			"type":       "object",
			"properties": properties,
			"required":   required,
		}
	}

	// If already a map, check if it's JSON schema
	if schemaMap, ok := schema.(map[string]any); ok {
		if _, hasType := schemaMap["type"]; hasType {
			if _, hasProps := schemaMap["properties"]; hasProps {
				// Already JSON schema
				return schemaMap
			}
		}

		// Simple map of name -> type
		properties := make(map[string]any)
		required := make([]string, 0, len(schemaMap))
		for name, typeVal := range schemaMap {
			required = append(required, name)
			properties[name] = s.convertTypeToSchema(typeVal)
		}

		return map[string]any{
			"type":       "object",
			"properties": properties,
			"required":   required,
		}
	}

	// If it's a struct type, use reflection
	if reflect.TypeOf(schema).Kind() == reflect.Struct {
		return s.structToSchema(reflect.TypeOf(schema))
	}

	// Default: empty object
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (s *SdkMcpServer) convertTypeToSchema(typeVal any) map[string]any {
	typeStr, ok := typeVal.(string)
	if !ok {
		return map[string]any{"type": "string"}
	}

	switch typeStr {
	case "string":
		return map[string]any{"type": "string"}
	case "number", "float", "float64":
		return map[string]any{"type": "number"}
	case "integer", "int":
		return map[string]any{"type": "integer"}
	case "boolean", "bool":
		return map[string]any{"type": "boolean"}
	default:
		return map[string]any{"type": "string"}
	}
}

func (s *SdkMcpServer) structToSchema(t reflect.Type) map[string]any {
	properties := make(map[string]any)
	required := make([]string, 0)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}

		// Parse json tag
		name := jsonTag
		for idx := 0; idx < len(jsonTag); idx++ {
			if jsonTag[idx] == ',' {
				name = jsonTag[:idx]
				break
			}
		}

		// Skip if "-"
		if name == "-" {
			continue
		}

		properties[name] = s.typeToSchema(field.Type)

		// Check if required (not a pointer)
		if field.Type.Kind() != reflect.Ptr {
			required = append(required, name)
		}
	}

	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

func (s *SdkMcpServer) typeToSchema(t reflect.Type) map[string]any {
	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Ptr:
		return s.typeToSchema(t.Elem())
	case reflect.Struct:
		return s.structToSchema(t)
	case reflect.Slice, reflect.Array:
		return map[string]any{
			"type":  "array",
			"items": s.typeToSchema(t.Elem()),
		}
	default:
		return map[string]any{"type": "string"}
	}
}

// TextContent creates a response with text content.
//
// Example:
//
//	return mcp.TextContent("The answer is 42")
func TextContent(text string) map[string]any {
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
	}
}

// ErrorContent creates an error response with text content.
//
// Example:
//
//	if input < 0 {
//	    return mcp.ErrorContent("Error: Input must be positive"), nil
//	}
func ErrorContent(text string) map[string]any {
	return map[string]any{
		"content": []map[string]any{
			{"type": "text", "text": text},
		},
		"isError": true,
	}
}

// ImageContent creates a response with image content.
// The data parameter should be a base64-encoded string.
// Common MIME types: "image/png", "image/jpeg", "image/gif", "image/webp"
//
// Example:
//
//	imageData := base64.StdEncoding.EncodeToString(pngBytes)
//	return mcp.ImageContent(imageData, "image/png"), nil
func ImageContent(data string, mimeType string) map[string]any {
	return map[string]any{
		"content": []map[string]any{
			{"type": "image", "data": data, "mimeType": mimeType},
		},
	}
}

// MixedContent creates a response with multiple content blocks (text, images, etc).
//
// Example:
//
//	return MixedContent(
//	    map[string]any{"type": "text", "text": "Here's the chart:"},
//	    map[string]any{"type": "image", "data": base64Data, "mimeType": "image/png"},
//	)
func MixedContent(blocks ...map[string]any) map[string]any {
	return map[string]any{
		"content": blocks,
	}
}