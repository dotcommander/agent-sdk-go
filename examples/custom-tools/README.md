# Example: Custom Tools

## What This Demonstrates

This example shows how to create and register custom tools for Claude using the Go SDK. It demonstrates:

- Creating tool definitions with JSON schemas
- Implementing tool executors with proper input validation
- Registering tools with MCP SDK servers
- Handling tool results and errors gracefully

## Prerequisites

- Claude Code CLI installed and authenticated
- Go 1.21+

## Quick Start

```bash
cd examples/custom-tools
go run main.go

# Or run with interactive demo (requires Claude CLI)
go run main.go demo
```

## Expected Output

```
=== Custom Tools Example ===
This demonstrates custom tool registration and execution patterns.

--- Registered Tools ---
Server: custom-tools v1.0.0
  - get_weather: Get current weather for a location
  - calculate: Perform mathematical calculations
  - analyze_text: Analyze text properties like word count, character count, etc.
  - fetch_url: Fetch metadata about a URL

--- Direct Tool Execution ---
Weather: {"content":[{"type":"text","text":"Weather for San Francisco:..."}]}
Calculator: {"content":[{"type":"text","text":"multiply(7, 8) = 56"}]}
Text analysis: {"content":[{"type":"text","text":"{\"character_count\":28,...}"}]}

--- Error Handling ---
Division by zero handled: {"content":[{"type":"text","text":"cannot divide by zero"}],"isError":true}

--- MCP Server Configuration ---
Server config type: sdk_mcp
Server config name: custom-tools

=== Example Completed ===
```

## Key Patterns

### Pattern 1: Tool Executor Interface

Tools implement a simple function signature that receives context and arguments:

```go
type WeatherTool struct{}

func (w *WeatherTool) Execute(ctx context.Context, args map[string]any) (map[string]any, error) {
    location, ok := args["location"].(string)
    if !ok {
        return mcp.ErrorContent("location must be a string"), nil
    }

    // Tool implementation
    return mcp.TextContent("Weather data..."), nil
}
```

### Pattern 2: MCP Tool Registration

Tools are registered using the `mcp.Tool()` helper:

```go
weatherTool := mcp.Tool(
    "get_weather",                           // Tool name
    "Get current weather for a location",   // Description
    map[string]string{"location": "string"}, // Parameter schema
    func(ctx context.Context, args map[string]any) (map[string]any, error) {
        return weatherExecutor.Execute(ctx, args)
    },
)
```

### Pattern 3: MCP Server Creation

Multiple tools are grouped into an MCP server:

```go
server := mcp.CreateSdkMcpServer(
    "custom-tools",  // Server name
    "1.0.0",         // Version
    []*mcp.SdkMcpTool{weatherTool, calculatorTool, ...},
)
```

### Pattern 4: Error Handling in Tools

Tools return errors using the `mcp.ErrorContent()` helper:

```go
if b == 0 {
    return mcp.ErrorContent("cannot divide by zero"), nil
}
```

## Tool Types Demonstrated

| Tool | Purpose | Parameters |
|------|---------|------------|
| `get_weather` | Fetch weather data | `location: string` |
| `calculate` | Math operations | `operation: string, a: number, b: number` |
| `analyze_text` | Text analysis | `text: string` |
| `fetch_url` | URL metadata | `url: string` |

## TypeScript Equivalent

This ports the TypeScript example from:
https://github.com/anthropics/claude-agent-sdk-demos/tree/main/custom-tools

The TypeScript version uses:
```typescript
const tool = {
    name: "get_weather",
    description: "Get weather for a location",
    inputSchema: { type: "object", properties: { location: { type: "string" } } }
};

class WeatherExecutor implements ToolExecutor {
    async execute(args: any): Promise<any> {
        return { result: "weather data" };
    }
}
```

## Related Documentation

- [MCP SDK Server Documentation](../../docs/usage.md#mcp-tools)
- [Tool Registration Guide](../../docs/usage.md#tool-registration)
- [Error Handling](../../docs/usage.md#error-handling)
