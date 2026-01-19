// Package main demonstrates creating SDK MCP servers with the Claude Agent SDK.
//
// SDK MCP servers allow you to:
// - Define tools that run in your Go process
// - Handle tool calls without external processes
// - Build custom integrations with full Go capabilities
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos
package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/mcp"
)

func main() {
	fmt.Println("=== SDK MCP Server Example ===")
	fmt.Println("This demonstrates creating in-process MCP tool servers.")
	fmt.Println()

	// Show basic SDK MCP server creation
	demonstrateBasicServer()

	// Show tool definition patterns
	demonstrateToolDefinitions()

	// Show tool handler implementation
	demonstrateToolHandlers()

	// Show integration with client
	demonstrateClientIntegration()

	fmt.Println()
	fmt.Println("=== SDK MCP Server Example Complete ===")
}

// demonstrateBasicServer shows basic SDK MCP server creation.
func demonstrateBasicServer() {
	fmt.Println("--- Basic SDK MCP Server ---")
	fmt.Println()

	// Create tools using the SdkMcpTool type
	weatherTool := &mcp.SdkMcpTool{
		Name:        "get_weather",
		Description: "Get current weather for a location",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"location": map[string]any{
					"type":        "string",
					"description": "City name or coordinates",
				},
				"units": map[string]any{
					"type":        "string",
					"enum":        []string{"celsius", "fahrenheit"},
					"description": "Temperature units",
				},
			},
			"required": []string{"location"},
		},
		Handler: func(ctx context.Context, args map[string]any) (map[string]any, error) {
			location := args["location"].(string)
			return map[string]any{
				"location":    location,
				"temperature": 22,
				"conditions":  "sunny",
			}, nil
		},
	}

	// Create the server with name, version, and tools
	server := mcp.CreateSdkMcpServer("weather-service", "1.0.0", []*mcp.SdkMcpTool{weatherTool})

	printJSON("Server Config", server.ToConfig())

	fmt.Println(`
  // Register with client
  client, err := claude.NewClient(
      claude.WithMcpServers(map[string]shared.McpServerConfig{
          "weather": server.ToConfig(),
      }),
  )`)
	fmt.Println()
}

// demonstrateToolDefinitions shows various tool definition patterns.
func demonstrateToolDefinitions() {
	fmt.Println("--- Tool Definition Patterns ---")
	fmt.Println()

	fmt.Println("1. Simple tool with basic types:")
	tool1 := &mcp.SdkMcpTool{
		Name:        "calculate",
		Description: "Perform arithmetic calculations",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"expression": map[string]any{
					"type":        "string",
					"description": "Mathematical expression to evaluate",
				},
			},
			"required": []string{"expression"},
		},
		Handler: func(ctx context.Context, args map[string]any) (map[string]any, error) {
			expr := args["expression"].(string)
			// In real implementation, evaluate the expression
			return map[string]any{"result": expr, "evaluated": true}, nil
		},
	}
	printJSON("Tool", map[string]any{
		"name":        tool1.Name,
		"description": tool1.Description,
		"inputSchema": tool1.InputSchema,
	})

	fmt.Println("2. Tool with complex input:")
	tool2 := &mcp.SdkMcpTool{
		Name:        "send_email",
		Description: "Send an email message",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"to": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "List of recipient email addresses",
				},
				"subject": map[string]any{
					"type":        "string",
					"description": "Email subject line",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "Email body content",
				},
			},
			"required": []string{"to", "subject", "body"},
		},
	}
	printJSON("Tool", map[string]any{
		"name":        tool2.Name,
		"description": tool2.Description,
	})

	fmt.Println("3. Using the Tool helper function:")
	fmt.Println(`
  // Helper function creates tool with handler
  tool := mcp.Tool(
      "query_database",
      "Execute a database query",
      map[string]any{
          "type": "object",
          "properties": map[string]any{
              "query":    map[string]any{"type": "string"},
              "database": map[string]any{"type": "string"},
          },
          "required": []string{"query", "database"},
      },
      func(ctx context.Context, args map[string]any) (map[string]any, error) {
          query := args["query"].(string)
          // Execute query...
          return map[string]any{"rows": 10, "query": query}, nil
      },
  )`)
	fmt.Println()
}

// demonstrateToolHandlers shows tool handler implementation.
func demonstrateToolHandlers() {
	fmt.Println("--- Tool Handler Implementation ---")
	fmt.Println()

	fmt.Println(`
  // Handler function signature
  type Handler func(ctx context.Context, args map[string]any) (map[string]any, error)

  // Example handler implementations
  func handleQuery(ctx context.Context, args map[string]any) (map[string]any, error) {
      query := args["query"].(string)
      database := args["database"].(string)

      // Execute query (with proper validation!)
      rows, err := db.QueryContext(ctx, query)
      if err != nil {
          return nil, fmt.Errorf("query failed: %%w", err)
      }

      // Return results as map
      return map[string]any{
          "rows":     formatResults(rows),
          "database": database,
      }, nil
  }`)
	fmt.Println()

	fmt.Println("Example: Returning different content types")
	fmt.Println(`
  // Text result
  return map[string]any{
      "text": "Operation completed successfully",
  }, nil

  // Structured data
  return map[string]any{
      "users": []map[string]any{
          {"id": 1, "name": "Alice"},
          {"id": 2, "name": "Bob"},
      },
      "total": 2,
  }, nil

  // Error (return error, not nil)
  if err != nil {
      return nil, fmt.Errorf("failed: %%w", err)
  }`)
	fmt.Println()
}

// demonstrateClientIntegration shows integrating SDK MCP servers with the client.
func demonstrateClientIntegration() {
	fmt.Println("--- Client Integration ---")
	fmt.Println()

	fmt.Println(`
  // Create tools with handlers
  tools := []*mcp.SdkMcpTool{
      mcp.Tool("get_user", "Get user by ID", userSchema, handleGetUser),
      mcp.Tool("update_user", "Update user data", updateSchema, handleUpdateUser),
      mcp.Tool("delete_user", "Delete a user", deleteSchema, handleDeleteUser),
  }

  // Create SDK MCP server
  server := mcp.CreateSdkMcpServer("user-service", "1.0.0", tools)

  // Create client with the server
  client, err := claude.NewClient(
      claude.WithModel("claude-sonnet-4-20250514"),
      claude.WithMcpServers(map[string]shared.McpServerConfig{
          "users": server.ToConfig(),
      }),
  )

  // Query - Claude will automatically use the tools
  response, err := client.Query(ctx, "Get the user with ID 12345")`)
	fmt.Println()

	fmt.Println("Example: Multiple SDK MCP servers")
	fmt.Println(`
  // Create multiple specialized servers
  userServer := mcp.CreateSdkMcpServer("user-service", "1.0.0", userTools)
  productServer := mcp.CreateSdkMcpServer("product-service", "1.0.0", productTools)
  orderServer := mcp.CreateSdkMcpServer("order-service", "1.0.0", orderTools)

  client, err := claude.NewClient(
      claude.WithMcpServers(map[string]shared.McpServerConfig{
          "users":    userServer.ToConfig(),
          "products": productServer.ToConfig(),
          "orders":   orderServer.ToConfig(),
      }),
  )

  // Claude can use tools from all servers
  response, err := client.Query(ctx,
      "Find user john@example.com, show their recent orders, " +
      "and check if product SKU-123 is in stock")`)
	fmt.Println()
}

// printJSON prints a labeled JSON object.
func printJSON(label string, v any) {
	data, err := json.MarshalIndent(v, "  ", "  ")
	if err != nil {
		fmt.Printf("  %s: (error: %v)\n", label, err)
		return
	}
	fmt.Printf("  %s:\n  %s\n\n", label, string(data))
}
