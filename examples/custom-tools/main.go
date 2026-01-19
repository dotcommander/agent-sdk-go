// Package main demonstrates custom tool registration and execution with the Claude Agent SDK.
//
// This example shows:
// - Creating custom tool definitions with JSON schemas
// - Implementing the ToolExecutor interface
// - Registering tools with the MCP SDK server
// - Tool invocation and result handling
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos/tree/main/custom-tools
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude"
	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/claude/mcp"
)

// WeatherTool demonstrates a custom tool that fetches weather data.
// In a real application, this would call an actual weather API.
type WeatherTool struct{}

// Execute implements the tool execution logic.
func (w *WeatherTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	location, ok := args["location"].(string)
	if !ok {
		return mcp.ErrorContent("location must be a string"), nil
	}

	// Simulate API call with mock data
	weatherData := map[string]any{
		"location":    location,
		"temperature": 22.5,
		"unit":        "celsius",
		"conditions":  "partly cloudy",
		"humidity":    65,
		"wind_speed":  12.3,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(weatherData)
	if err != nil {
		return mcp.ErrorContent(fmt.Sprintf("failed to marshal weather data: %v", err)), nil
	}

	return mcp.TextContent(fmt.Sprintf("Weather for %s:\n%s", location, string(jsonData))), nil
}

// CalculatorTool demonstrates a more complex tool with multiple operations.
type CalculatorTool struct{}

// Execute implements the calculator tool logic.
func (c *CalculatorTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	operation, ok := args["operation"].(string)
	if !ok {
		return mcp.ErrorContent("operation must be a string"), nil
	}

	a, ok := args["a"].(float64)
	if !ok {
		return mcp.ErrorContent("parameter 'a' must be a number"), nil
	}

	b, ok := args["b"].(float64)
	if !ok {
		return mcp.ErrorContent("parameter 'b' must be a number"), nil
	}

	var result float64
	switch strings.ToLower(operation) {
	case "add":
		result = a + b
	case "subtract":
		result = a - b
	case "multiply":
		result = a * b
	case "divide":
		if b == 0 {
			return mcp.ErrorContent("cannot divide by zero"), nil
		}
		result = a / b
	case "power":
		result = math.Pow(a, b)
	case "modulo":
		if b == 0 {
			return mcp.ErrorContent("cannot modulo by zero"), nil
		}
		result = math.Mod(a, b)
	default:
		return mcp.ErrorContent(fmt.Sprintf("unknown operation: %s", operation)), nil
	}

	return mcp.TextContent(fmt.Sprintf("%s(%g, %g) = %g", operation, a, b, result)), nil
}

// TextAnalyzerTool demonstrates a tool that processes text.
type TextAnalyzerTool struct{}

// Execute implements text analysis logic.
func (t *TextAnalyzerTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	text, ok := args["text"].(string)
	if !ok {
		return mcp.ErrorContent("text must be a string"), nil
	}

	analysis := map[string]any{
		"character_count": len(text),
		"word_count":      len(strings.Fields(text)),
		"line_count":      len(strings.Split(text, "\n")),
		"has_numbers":     strings.ContainsAny(text, "0123456789"),
		"is_uppercase":    text == strings.ToUpper(text),
		"is_lowercase":    text == strings.ToLower(text),
	}

	jsonData, err := json.Marshal(analysis)
	if err != nil {
		return mcp.ErrorContent(fmt.Sprintf("failed to marshal analysis: %v", err)), nil
	}

	return mcp.TextContent(string(jsonData)), nil
}

// URLFetcherTool demonstrates a tool that fetches URL content.
type URLFetcherTool struct {
	client *http.Client
}

// NewURLFetcherTool creates a new URL fetcher with timeout.
func NewURLFetcherTool(timeout time.Duration) *URLFetcherTool {
	return &URLFetcherTool{
		client: &http.Client{Timeout: timeout},
	}
}

// Execute implements URL fetching logic.
func (u *URLFetcherTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
	url, ok := args["url"].(string)
	if !ok {
		return mcp.ErrorContent("url must be a string"), nil
	}

	// Validate URL format
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return mcp.ErrorContent("url must start with http:// or https://"), nil
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return mcp.ErrorContent(fmt.Sprintf("failed to create request: %v", err)), nil
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return mcp.ErrorContent(fmt.Sprintf("failed to fetch URL: %v", err)), nil
	}
	defer resp.Body.Close()

	result := map[string]any{
		"url":          url,
		"status_code":  resp.StatusCode,
		"content_type": resp.Header.Get("Content-Type"),
		"server":       resp.Header.Get("Server"),
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		return mcp.ErrorContent(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return mcp.TextContent(string(jsonData)), nil
}

func main() {
	// Check CLI availability first
	if !cli.IsCLIAvailable() {
		log.Fatal("Claude CLI not found. Please install it first: https://docs.anthropic.com/claude/docs/quickstart")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	fmt.Println("=== Custom Tools Example ===")
	fmt.Println("This demonstrates custom tool registration and execution patterns.")
	fmt.Println()

	// Create tool instances
	weatherTool := &WeatherTool{}
	calculatorTool := &CalculatorTool{}
	textAnalyzerTool := &TextAnalyzerTool{}
	urlFetcherTool := NewURLFetcherTool(10 * time.Second)

	// Create MCP tools with proper schemas
	weatherMcpTool := mcp.Tool(
		"get_weather",
		"Get current weather for a location",
		map[string]string{"location": "string"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return weatherTool.Execute(ctx, args)
		},
	)

	calculatorMcpTool := mcp.Tool(
		"calculate",
		"Perform mathematical calculations",
		map[string]string{
			"operation": "string",
			"a":         "number",
			"b":         "number",
		},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return calculatorTool.Execute(ctx, args)
		},
	)

	textAnalyzerMcpTool := mcp.Tool(
		"analyze_text",
		"Analyze text properties like word count, character count, etc.",
		map[string]string{"text": "string"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return textAnalyzerTool.Execute(ctx, args)
		},
	)

	urlFetcherMcpTool := mcp.Tool(
		"fetch_url",
		"Fetch metadata about a URL",
		map[string]string{"url": "string"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			return urlFetcherTool.Execute(ctx, args)
		},
	)

	// Create MCP server with all tools
	toolServer := mcp.CreateSdkMcpServer(
		"custom-tools",
		"1.0.0",
		[]*mcp.SdkMcpTool{
			weatherMcpTool,
			calculatorMcpTool,
			textAnalyzerMcpTool,
			urlFetcherMcpTool,
		},
	)

	// Convert to server config
	serverConfig := toolServer.ToConfig()

	fmt.Println("--- Registered Tools ---")
	fmt.Printf("Server: %s v%s\n", toolServer.Name, toolServer.Version)
	for _, tool := range toolServer.Tools {
		fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}
	fmt.Println()

	// Configure client with MCP servers
	servers := map[string]claude.McpServerConfig{
		"tools": serverConfig,
	}

	client, err := claude.NewClient(
		claude.WithModel("claude-sonnet-4-5-20250929"),
		claude.WithMcpServers(servers),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Demonstrate tool execution directly
	fmt.Println("--- Direct Tool Execution ---")

	// Test weather tool
	weatherResult, err := weatherTool.Execute(ctx, map[string]any{"location": "San Francisco"})
	if err != nil {
		log.Printf("Weather tool error: %v", err)
	} else {
		fmt.Printf("Weather: %v\n", weatherResult["content"])
	}
	fmt.Println()

	// Test calculator tool
	calcResult, err := calculatorTool.Execute(ctx, map[string]any{
		"operation": "multiply",
		"a":         float64(7),
		"b":         float64(8),
	})
	if err != nil {
		log.Printf("Calculator tool error: %v", err)
	} else {
		fmt.Printf("Calculator: %v\n", calcResult["content"])
	}
	fmt.Println()

	// Test text analyzer tool
	textResult, err := textAnalyzerTool.Execute(ctx, map[string]any{
		"text": "Hello World!\nThis is a test.",
	})
	if err != nil {
		log.Printf("Text analyzer error: %v", err)
	} else {
		fmt.Printf("Text analysis: %v\n", textResult["content"])
	}
	fmt.Println()

	// Demonstrate error handling
	fmt.Println("--- Error Handling ---")
	divResult, err := calculatorTool.Execute(ctx, map[string]any{
		"operation": "divide",
		"a":         float64(10),
		"b":         float64(0),
	})
	if err != nil {
		log.Printf("Expected error: %v", err)
	} else {
		fmt.Printf("Division by zero handled: %v\n", divResult["content"])
	}
	fmt.Println()

	// Show MCP configuration
	fmt.Println("--- MCP Server Configuration ---")
	fmt.Printf("Server config type: %s\n", serverConfig.Type)
	fmt.Printf("Server config name: %s\n", serverConfig.Name)

	// Note: Full tool execution through Claude requires CLI integration
	fmt.Println()
	fmt.Println("=== Example Completed ===")
	fmt.Println("Custom tools are registered and ready for Claude CLI integration.")
	fmt.Println("When Claude CLI invokes these tools, the handlers will execute.")

	// Demonstrate setting MCP servers on client
	fmt.Println()
	fmt.Println("--- Client MCP Server Status ---")
	result, err := client.SetMcpServers(ctx, servers)
	if err != nil {
		fmt.Printf("Note: SetMcpServers returned: %v\n", err)
		fmt.Println("This is expected without full CLI integration.")
	} else if result != nil {
		fmt.Printf("MCP servers configured: %d added\n", len(result.Added))
	}

	// Command line mode for interactive testing
	if len(os.Args) > 1 && os.Args[1] == "demo" {
		runInteractiveDemo(ctx, client)
	}
}

// runInteractiveDemo demonstrates tool usage with Claude.
func runInteractiveDemo(ctx context.Context, client claude.Client) {
	fmt.Println()
	fmt.Println("=== Interactive Demo ===")

	// Connect to Claude
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect()

	// Send a query that would use tools
	prompt := "Calculate 15 * 7 + 3, then analyze the text 'Hello World'"

	fmt.Printf("Prompt: %s\n\n", prompt)

	msgChan, errChan := client.QueryStream(ctx, prompt)

	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				fmt.Println("\nStream ended")
				return
			}
			if text := claude.GetContentText(msg); text != "" {
				fmt.Print(text)
			}
		case err := <-errChan:
			if err != nil {
				fmt.Printf("\nError: %v\n", err)
			}
			return
		case <-ctx.Done():
			fmt.Println("\nContext cancelled")
			return
		}
	}
}
