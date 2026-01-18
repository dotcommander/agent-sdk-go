package claude

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewTool tests the tool creation function.
func TestNewTool(t *testing.T) {
	t.Parallel()

	t.Run("creates tool with all fields", func(t *testing.T) {
		handler := func(_ context.Context, args map[string]any) (*McpToolResult, error) {
			return &McpToolResult{
				Content: []McpContent{{Type: "text", Text: "result"}},
			}, nil
		}

		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{"type": "string"},
			},
		}

		tool := NewTool("greet", "Greet a person", schema, handler)

		assert.Equal(t, "greet", tool.Name())
		assert.Equal(t, "Greet a person", tool.Description())
		assert.Equal(t, schema, tool.InputSchema())
	})

	t.Run("tool handler is callable", func(t *testing.T) {
		called := false
		tool := NewTool("test", "Test tool", nil, func(ctx context.Context, args map[string]any) (*McpToolResult, error) {
			called = true
			name, _ := args["name"].(string)
			return &McpToolResult{
				Content: []McpContent{{Type: "text", Text: fmt.Sprintf("Hello, %s!", name)}},
			}, nil
		})

		result, err := tool.Call(context.Background(), map[string]any{"name": "World"})

		require.NoError(t, err)
		assert.True(t, called)
		require.Len(t, result.Content, 1)
		assert.Equal(t, "text", result.Content[0].Type)
		assert.Equal(t, "Hello, World!", result.Content[0].Text)
	})

	t.Run("nil handler returns error", func(t *testing.T) {
		tool := NewTool("test", "Test tool", nil, nil)

		_, err := tool.Call(context.Background(), nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no handler")
	})
}

// TestSdkMcpServer tests the in-process MCP server.
func TestSdkMcpServer(t *testing.T) {
	t.Parallel()

	addTool := NewTool("add", "Add numbers", map[string]any{
		"type": "object",
		"properties": map[string]any{
			"a": map[string]any{"type": "number"},
			"b": map[string]any{"type": "number"},
		},
	}, func(_ context.Context, args map[string]any) (*McpToolResult, error) {
		a, _ := args["a"].(float64)
		b, _ := args["b"].(float64)
		return &McpToolResult{
			Content: []McpContent{{Type: "text", Text: fmt.Sprintf("%.2f", a+b)}},
		}, nil
	})

	sqrtTool := NewTool("sqrt", "Square root", nil, func(_ context.Context, args map[string]any) (*McpToolResult, error) {
		return &McpToolResult{Content: []McpContent{{Type: "text", Text: "result"}}}, nil
	})

	t.Run("creates server with tools", func(t *testing.T) {
		config := CreateSDKMcpServer("calculator", "1.0.0", addTool, sqrtTool)

		assert.Equal(t, "sdk", config.Type)
		assert.Equal(t, "calculator", config.Name)
		assert.NotNil(t, config.Instance)
	})

	t.Run("server name and version", func(t *testing.T) {
		config := CreateSDKMcpServer("calc", "2.0.0", addTool)
		server := config.Instance.(*SdkMcpServer)

		assert.Equal(t, "calc", server.Name())
		assert.Equal(t, "2.0.0", server.Version())
	})

	t.Run("ListTools returns tool definitions", func(t *testing.T) {
		config := CreateSDKMcpServer("calc", "1.0.0", addTool, sqrtTool)
		server := config.Instance.(*SdkMcpServer)

		tools, err := server.ListTools(context.Background())

		require.NoError(t, err)
		assert.Len(t, tools, 2)

		// Find add tool
		var foundAdd bool
		for _, tool := range tools {
			if tool.Name == "add" {
				foundAdd = true
				assert.Equal(t, "Add numbers", tool.Description)
			}
		}
		assert.True(t, foundAdd)
	})

	t.Run("CallTool executes handler", func(t *testing.T) {
		config := CreateSDKMcpServer("calc", "1.0.0", addTool)
		server := config.Instance.(*SdkMcpServer)

		result, err := server.CallTool(context.Background(), "add", map[string]any{"a": 2.0, "b": 3.0})

		require.NoError(t, err)
		require.Len(t, result.Content, 1)
		assert.Equal(t, "5.00", result.Content[0].Text)
	})

	t.Run("CallTool returns error for unknown tool", func(t *testing.T) {
		config := CreateSDKMcpServer("calc", "1.0.0", addTool)
		server := config.Instance.(*SdkMcpServer)

		_, err := server.CallTool(context.Background(), "unknown", nil)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("AddTool and RemoveTool", func(t *testing.T) {
		config := CreateSDKMcpServer("calc", "1.0.0")
		server := config.Instance.(*SdkMcpServer)

		// Initially no tools
		tools, _ := server.ListTools(context.Background())
		assert.Len(t, tools, 0)

		// Add tool
		server.AddTool(addTool)
		tools, _ = server.ListTools(context.Background())
		assert.Len(t, tools, 1)

		// Remove tool
		removed := server.RemoveTool("add")
		assert.True(t, removed)

		tools, _ = server.ListTools(context.Background())
		assert.Len(t, tools, 0)

		// Remove non-existent tool
		removed = server.RemoveTool("nonexistent")
		assert.False(t, removed)
	})

	t.Run("nil tools are ignored", func(t *testing.T) {
		config := CreateSDKMcpServer("calc", "1.0.0", addTool, nil, sqrtTool, nil)
		server := config.Instance.(*SdkMcpServer)

		tools, _ := server.ListTools(context.Background())
		assert.Len(t, tools, 2)
	})
}

// TestMcpToolResult tests the tool result types.
func TestMcpToolResult(t *testing.T) {
	t.Parallel()

	t.Run("text content", func(t *testing.T) {
		result := McpToolResult{
			Content: []McpContent{{Type: "text", Text: "Hello"}},
		}
		assert.Len(t, result.Content, 1)
		assert.Equal(t, "text", result.Content[0].Type)
		assert.Equal(t, "Hello", result.Content[0].Text)
	})

	t.Run("image content", func(t *testing.T) {
		result := McpToolResult{
			Content: []McpContent{{
				Type:     "image",
				Data:     "base64data",
				MimeType: "image/png",
			}},
		}
		assert.Equal(t, "image", result.Content[0].Type)
		assert.Equal(t, "base64data", result.Content[0].Data)
		assert.Equal(t, "image/png", result.Content[0].MimeType)
	})

	t.Run("error result", func(t *testing.T) {
		result := McpToolResult{
			Content: []McpContent{{Type: "text", Text: "Error occurred"}},
			IsError: true,
		}
		assert.True(t, result.IsError)
		assert.Equal(t, "Error occurred", result.Content[0].Text)
	})
}

// TestWithSdkMcpServer tests the option function.
func TestWithSdkMcpServer(t *testing.T) {
	t.Parallel()

	t.Run("adds server to options", func(t *testing.T) {
		tool := NewTool("test", "Test", nil, nil)
		config := CreateSDKMcpServer("calc", "1.0.0", tool)

		opts := &ClientOptions{}
		WithSdkMcpServer("calc", config)(opts)

		assert.NotNil(t, opts.McpServers)
		assert.Contains(t, opts.McpServers, "calc")
	})

	t.Run("multiple servers accumulate", func(t *testing.T) {
		config1 := CreateSDKMcpServer("server1", "1.0.0")
		config2 := CreateSDKMcpServer("server2", "1.0.0")

		opts := &ClientOptions{}
		WithSdkMcpServer("s1", config1)(opts)
		WithSdkMcpServer("s2", config2)(opts)

		assert.Contains(t, opts.McpServers, "s1")
		assert.Contains(t, opts.McpServers, "s2")
	})
}

// TestWithAllowedTools tests the allowed tools option.
func TestWithAllowedTools(t *testing.T) {
	t.Parallel()

	t.Run("adds tools to custom args", func(t *testing.T) {
		opts := &ClientOptions{}
		WithAllowedTools("mcp__calc__add", "mcp__calc__sqrt")(opts)

		assert.Contains(t, opts.CustomArgs, "--allowed-tools")
		assert.Contains(t, opts.CustomArgs, "mcp__calc__add")
		assert.Contains(t, opts.CustomArgs, "mcp__calc__sqrt")
	})
}
