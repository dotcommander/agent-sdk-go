// Package main demonstrates SDK MCP tool functionality for agent-sdk-go.
// This example shows how to create custom in-process MCP tools and use them with Claude.
//
// Note: This is a CLI wrapper, not an HTTP client SDK. The MCP tools run in-process
// in the Go application and are passed to Claude CLI via --mcp-config flag.
package main

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/dotcommander/agent-sdk-go/claude"
	"github.com/dotcommander/agent-sdk-go/claude/mcp"
	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

func main() {
	fmt.Println("=== SDK MCP Tools Example ===")
	fmt.Println("This demonstrates in-process MCP tool integration with agent-sdk-go")

	// Create custom tools
	addTool := mcp.Tool(
		"add",
		"Add two numbers together",
		map[string]string{"a": "number", "b": "number"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			sum := a + b
			return mcp.TextContent(fmt.Sprintf("The sum of %.2f and %.2f is %.2f", a, b, sum)), nil
		},
	)

	multiplyTool := mcp.Tool(
		"multiply",
		"Multiply two numbers",
		map[string]string{"a": "number", "b": "number"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			product := a * b
			return mcp.TextContent(fmt.Sprintf("The product of %.2f and %.2f is %.2f", a, b, product)), nil
		},
	)

	subtractTool := mcp.Tool(
		"subtract",
		"Subtract one number from another",
		map[string]string{"a": "number", "b": "number"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			a := args["a"].(float64)
			b := args["b"].(float64)
			difference := a - b
			return mcp.TextContent(fmt.Sprintf("%.2f minus %.2f is %.2f", a, b, difference)), nil
		},
	)

	powerTool := mcp.Tool(
		"power",
		"Raise a number to a power",
		map[string]string{"base": "number", "exponent": "number"},
		func(ctx context.Context, args map[string]any) (map[string]any, error) {
			base := args["base"].(float64)
			exponent := args["exponent"].(float64)
			result := math.Pow(base, exponent)
			return mcp.TextContent(fmt.Sprintf("%.2f raised to the power of %.2f is %.2f", base, exponent, result)), nil
		},
	)

	// Create SDK MCP server
	calculatorServer := mcp.CreateSdkMcpServer(
		"calculator",
		"1.0.0",
		[]*mcp.SdkMcpTool{addTool, subtractTool, multiplyTool, powerTool},
	)

	// Convert server to config
	serverConfig := calculatorServer.ToConfig()

	// Configure client options with MCP servers
	servers := map[string]shared.McpServerConfig{
		"calc": serverConfig,
	}

	options := []claude.ClientOption{
		claude.WithModel("claude-sonnet-4-5-20250929"),
		claude.WithMcpServers(servers),
	}

	// Create client
	ctx := context.Background()
	client, err := claude.NewClient(options...)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Demonstrate MCP server configuration
	fmt.Println("\n--- MCP Server Configuration ---")
	fmt.Printf("Created MCP server: %s v%s\n", calculatorServer.Name, calculatorServer.Version)
	fmt.Printf("Tools: %d math tools\n", len(calculatorServer.Tools))

	// Server config already created earlier
	fmt.Printf("Server config type: %s\n", serverConfig.Type)
	fmt.Printf("Server config name: %s\n", serverConfig.Name)

	// Demonstrate setting MCP servers
	fmt.Println("\n--- Setting MCP Servers ---")
	// Using already configured MCP servers

	result, err := client.SetMcpServers(ctx, servers)
	if err != nil {
		fmt.Printf("Note: SetMcpServers returned error: %v\n", err)
		fmt.Println("This is expected as the full implementation requires CLI integration")
	} else {
		fmt.Printf("SetMcpServers result: %d servers added\n", len(result.Added))
	}

	// Demonstrate MCP server status
	fmt.Println("\n--- MCP Server Status ---")
	status, err := client.McpServerStatus(ctx)
	if err != nil {
		fmt.Printf("Note: McpServerStatus returned error: %v\n", err)
		fmt.Println("This is expected as the full implementation requires CLI integration")
	} else {
		fmt.Printf("Retrieved %d server statuses\n", len(status))
	}

	// Note: In the actual CLI wrapper, MCP tools are passed to Claude CLI via --mcp-config
	// The SDK MCP server instance runs in-process and handles tool execution locally
	// When Claude CLI requests tool execution, it's routed back to the in-process server

	fmt.Println("\n=== Example Completed ===")
	fmt.Println("This demonstrates the MCP SDK parity implementation.")
	fmt.Println("Full tool execution requires Claude CLI to be installed and configured.")
}

// This example shows the API surface and demonstrates how to:
// 1. Create custom MCP tools with handlers
// 2. Create an SDK MCP server
// 3. Configure MCP servers in the client
// 4. Use MCP server management APIs
//
// For full tool execution, you would need to:
// 1. Connect to Claude CLI
// 2. Pass MCP config via --mcp-config flag
// 3. Handle tool execution requests from the CLI
//
// The MCP SDK parity implementation provides:
// - SdkMcpTool and SdkMcpServer types
// - Content helper functions (TextContent, ErrorContent, etc.)
// - Schema conversion utilities
// - Transport integration for passing MCP config to CLI
// - Client interface methods for MCP server management