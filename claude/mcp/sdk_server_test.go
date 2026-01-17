package mcp

import (
	"context"
	"fmt"
	"testing"
)

func TestToolCreation(t *testing.T) {
	tool := Tool(
		"greet",
		"Greet a user",
		map[string]string{"name": "string"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			name := args["name"].(string)
			return map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": fmt.Sprintf("Hello, %s!", name)},
				},
			}, nil
		},
	)

	if tool.Name != "greet" {
		t.Errorf("expected name 'greet', got %s", tool.Name)
	}
	if tool.Description != "Greet a user" {
		t.Errorf("expected description 'Greet a user', got %s", tool.Description)
	}
	if tool.Handler == nil {
		t.Error("expected handler to be set")
	}
}

func TestSdkMcpServerCreation(t *testing.T) {
	tool1 := Tool("tool1", "First tool", map[string]string{}, nil)
	tool2 := Tool("tool2", "Second tool", map[string]string{}, nil)

	server := CreateSdkMcpServer("test-server", "1.0.0", []*SdkMcpTool{tool1, tool2})

	if server.Name != "test-server" {
		t.Errorf("expected name 'test-server', got %s", server.Name)
	}
	if server.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", server.Version)
	}
	if len(server.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(server.Tools))
	}
}

func TestMcpServerHandleInitialize(t *testing.T) {
	server := CreateSdkMcpServer("test", "1.0.0", nil)

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
	}

	response := server.HandleRequest(context.Background(), request)

	if response["jsonrpc"] != "2.0" {
		t.Error("response should have jsonrpc 2.0")
	}
	if response["id"] != 1 {
		t.Error("response should have matching id")
	}

	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatal("expected result to be map")
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("expected protocol version '2024-11-05', got %v", result["protocolVersion"])
	}

	serverInfo, ok := result["serverInfo"].(map[string]any)
	if !ok {
		t.Fatal("expected serverInfo to be map")
	}
	if serverInfo["name"] != "test" {
		t.Errorf("expected server name 'test', got %v", serverInfo["name"])
	}
}

func TestMcpServerHandleListTools(t *testing.T) {
	addTool := Tool(
		"add",
		"Add two numbers",
		map[string]string{"a": "number", "b": "number"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			return TextContent(fmt.Sprintf("Sum: %f", a+b)), nil
		},
	)

	server := CreateSdkMcpServer("calculator", "1.0.0", []*SdkMcpTool{addTool})

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
	}

	response := server.HandleRequest(context.Background(), request)

	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatal("expected result to be map")
	}

	tools, ok := result["tools"].([]map[string]any)
	if !ok {
		t.Fatal("expected tools to be array")
	}

	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	if tools[0]["name"] != "add" {
		t.Errorf("expected tool name 'add', got %v", tools[0]["name"])
	}
	if tools[0]["description"] != "Add two numbers" {
		t.Errorf("expected tool description 'Add two numbers', got %v", tools[0]["description"])
	}

	schema, ok := tools[0]["inputSchema"].(map[string]any)
	if !ok {
		t.Fatal("expected inputSchema to be map")
	}
	if schema["type"] != "object" {
		t.Errorf("expected schema type 'object', got %v", schema["type"])
	}
}

func TestMcpServerHandleCallTool(t *testing.T) {
	called := false
	greetTool := Tool(
		"greet",
		"Greet a user",
		map[string]string{"name": "string"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			called = true
			name := args["name"].(string)
			return map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": fmt.Sprintf("Hello, %s!", name)},
				},
			}, nil
		},
	)

	server := CreateSdkMcpServer("test", "1.0.0", []*SdkMcpTool{greetTool})

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      3,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "greet",
			"arguments": map[string]any{"name": "Alice"},
		},
	}

	response := server.HandleRequest(context.Background(), request)

	if !called {
		t.Error("tool handler should have been called")
	}

	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatal("expected result to be map")
	}

	content, ok := result["content"].([]map[string]any)
	if !ok {
		t.Fatal("expected content to be array")
	}

	if len(content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(content))
	}

	if content[0]["type"] != "text" {
		t.Errorf("expected content type 'text', got %v", content[0]["type"])
	}
	if content[0]["text"] != "Hello, Alice!" {
		t.Errorf("expected text 'Hello, Alice!', got %v", content[0]["text"])
	}
}

func TestMcpServerHandleCallToolError(t *testing.T) {
	errorTool := Tool(
		"fail",
		"A tool that fails",
		map[string]string{},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return nil, fmt.Errorf("intentional error")
		},
	)

	server := CreateSdkMcpServer("test", "1.0.0", []*SdkMcpTool{errorTool})

	request := map[string]any{
		"jsonrpc": "2.0",
		"id":      4,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "fail",
			"arguments": map[string]any{},
		},
	}

	response := server.HandleRequest(context.Background(), request)

	result, ok := response["result"].(map[string]any)
	if !ok {
		t.Fatal("expected result to be map")
	}

	isError, ok := result["isError"].(bool)
	if !ok || !isError {
		t.Error("expected isError to be true")
	}

	content, ok := result["content"].([]map[string]any)
	if !ok {
		t.Fatal("expected content to be array")
	}

	if len(content) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(content))
	}

	text, ok := content[0]["text"].(string)
	if !ok {
		t.Fatal("expected text field")
	}
	if text != "Error: intentional error" {
		t.Errorf("expected error message, got %s", text)
	}
}

func TestTextContentHelper(t *testing.T) {
	content := TextContent("test message")

	contentArray, ok := content["content"].([]map[string]any)
	if !ok {
		t.Fatal("expected content to be array")
	}

	if len(contentArray) != 1 {
		t.Fatalf("expected 1 content item, got %d", len(contentArray))
	}

	if contentArray[0]["type"] != "text" {
		t.Errorf("expected type 'text', got %v", contentArray[0]["type"])
	}
	if contentArray[0]["text"] != "test message" {
		t.Errorf("expected text 'test message', got %v", contentArray[0]["text"])
	}
}

func TestErrorContentHelper(t *testing.T) {
	content := ErrorContent("error message")

	isError, ok := content["isError"].(bool)
	if !ok || !isError {
		t.Error("expected isError to be true")
	}

	contentArray, ok := content["content"].([]map[string]any)
	if !ok {
		t.Fatal("expected content to be array")
	}

	if contentArray[0]["text"] != "error message" {
		t.Errorf("expected text 'error message', got %v", contentArray[0]["text"])
	}
}

func TestMcpServerToConfig(t *testing.T) {
	tool := Tool("test", "test tool", map[string]string{}, nil)
	server := CreateSdkMcpServer("test-server", "1.0.0", []*SdkMcpTool{tool})

	config := server.ToConfig()

	if config.Type != "sdk" {
		t.Errorf("expected type 'sdk', got %s", config.Type)
	}
	if config.Name != "test-server" {
		t.Errorf("expected name 'test-server', got %s", config.Name)
	}
	if config.Instance != server {
		t.Error("expected instance to be the server")
	}
}