# MCP Tools Example

This example demonstrates SDK MCP (Model Context Protocol) tool functionality for `agent-sdk-go`.

## What This Demonstrates

- Creating custom in-process MCP tools with Go handlers
- Building an SDK MCP server with multiple tools
- Configuring MCP servers in the Claude client
- Using MCP server management APIs (`SetMcpServers`, `McpServerStatus`)

## Key Concepts

### SDK MCP Servers
Unlike external MCP servers (stdio, SSE, HTTP), SDK MCP servers run in-process within your Go application. When Claude CLI needs to execute a tool, the request is routed back to your in-process handler.

### Tool Creation
Tools are created with:
- Name and description
- Input schema (map of parameter names to types)
- Handler function that processes arguments and returns content

### Content Helpers
The MCP package provides helper functions for standardized responses:
- `TextContent()` - Text responses
- `ErrorContent()` - Error responses
- `ImageContent()` - Image responses
- `MixedContent()` - Mixed content types

## Running the Example

```bash
# From the project root
go run examples/mcp_tools/main.go
```

## Architecture

```
┌─────────────────────────────────────────────┐
│  Your Go Application                        │
│  ┌─────────────────────────────────────┐    │
│  │  SDK MCP Server                     │    │
│  │  ┌─────────────┐ ┌─────────────┐   │    │
│  │  │ Tool: add   │ │ Tool: multiply│ │    │
│  │  └─────────────┘ └─────────────┘   │    │
│  └─────────────────────────────────────┘    │
│                 │                           │
│  ┌──────────────▼──────────────┐           │
│  │ Claude Client               │           │
│  └──────────────┬──────────────┘           │
└─────────────────│──────────────────────────┘
                  │
┌─────────────────▼──────────────────────────┐
│ Claude CLI                                 │
│  ┌─────────────────────────────────────┐  │
│  │ MCP Config:                          │  │
│  │ {                                    │  │
│  │   "mcpServers": {                    │  │
│  │     "calc": {                        │  │
│  │       "type": "sdk",                 │  │
│  │       "name": "calculator"           │  │
│  │     }                                │  │
│  │   }                                  │  │
│  │ }                                    │  │
│  └─────────────────────────────────────┘  │
└───────────────────────────────────────────┘
```

## Implementation Details

The MCP SDK parity implementation includes:

### 1. Core Types (`claude/mcp/`)
- `SdkMcpTool` - Tool definition with handler
- `SdkMcpServer` - Server with tool collection
- JSON-RPC handlers for MCP protocol

### 2. Transport Integration (`claude/subprocess/`)
- `McpServers` field in `TransportConfig`
- `--mcp-config` flag generation in `buildArgs()`
- SDK server instance extraction (non-serialized)

### 3. Client Interface (`claude/types.go`)
- `McpServerStatus()` method in `Controller` interface
- `SetMcpServers()` method for dynamic configuration

### 4. Shared Types (`claude/shared/mcp.go`)
- `McpSdkServerConfig` with `Instance` field (marked `json:"-"`)
- Custom `MarshalJSON()` to exclude instance from serialization

## Notes

- This SDK uses subprocess transport to communicate with Claude CLI
- MCP tools run in-process within your Go application
- Full tool execution requires Claude CLI to be installed
- The example shows API usage; actual tool execution requires CLI integration

## See Also

- `SPEC-MCP-SDK-PARITY.md` - Full implementation specification
- `claude/mcp/sdk_server_test.go` - Test suite