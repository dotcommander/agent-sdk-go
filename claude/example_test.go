package claude_test

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude"
	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

// ExampleNewClient demonstrates creating a basic Claude client.
func ExampleNewClient() {
	client, err := claude.NewClient(
		claude.WithModel("claude-sonnet-4-20250514"),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	_ = client
	fmt.Println("Client created successfully")
	// Output: Client created successfully
}

// ExampleWithModel demonstrates setting the Claude model.
func ExampleWithModel() {
	client, err := claude.NewClient(
		claude.WithModel("claude-sonnet-4-20250514"),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.Model)
	// Output: claude-sonnet-4-20250514
}

// ExampleWithPermissionMode demonstrates setting permission modes.
// Valid modes are: auto, read, write, restricted
func ExampleWithPermissionMode() {
	client, err := claude.NewClient(
		claude.WithPermissionMode("restricted"),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.PermissionMode)
	// Output: restricted
}

// ExampleWithTimeout demonstrates setting operation timeout.
func ExampleWithTimeout() {
	client, err := claude.NewClient(
		claude.WithTimeout("2m30s"),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.Timeout)
	// Output: 2m30s
}

// ExampleWithEnv demonstrates setting environment variables.
func ExampleWithEnv() {
	client, err := claude.NewClient(
		claude.WithEnv(map[string]string{
			"MY_API_KEY": "secret",
			"DEBUG":      "true",
		}),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.Env["DEBUG"])
	// Output: true
}

// ExampleWithCanUseTool demonstrates permission callbacks.
func ExampleWithCanUseTool() {
	client, err := claude.NewClient(
		claude.WithCanUseTool(func(ctx context.Context, toolName string, toolInput map[string]any, opts shared.CanUseToolOptions) (shared.PermissionResult, error) {
			// Block dangerous bash commands
			if toolName == "Bash" {
				if cmd, ok := toolInput["command"].(string); ok {
					if len(cmd) >= 2 && cmd[0:2] == "rm" {
						return shared.NewPermissionResultDeny("rm commands are not allowed"), nil
					}
				}
			}
			return shared.NewPermissionResultAllow(), nil
		}),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.CanUseTool != nil)
	// Output: true
}

// ExampleWithJSONSchema demonstrates structured output with JSON schema.
func ExampleWithJSONSchema() {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
			"age":  map[string]any{"type": "number"},
		},
		"required": []string{"name", "age"},
	}

	client, err := claude.NewClient(
		claude.WithJSONSchema(schema),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.OutputFormat.Type)
	// Output: json_schema
}

// ExampleWithHooks demonstrates registering lifecycle hooks.
func ExampleWithHooks() {
	hooks := map[shared.HookEvent][]shared.HookConfig{
		shared.HookEventPreToolUse: {
			{
				Matcher: "Bash",
				Handler: func(ctx context.Context, input any) (*shared.SyncHookOutput, error) {
					return &shared.SyncHookOutput{}, nil
				},
			},
		},
	}

	client, err := claude.NewClient(
		claude.WithHooks(hooks),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(len(opts.Hooks))
	// Output: 1
}

// ExampleWithPreToolUseHook demonstrates the convenience hook helper.
func ExampleWithPreToolUseHook() {
	client, err := claude.NewClient(
		claude.WithPreToolUseHook(func(ctx context.Context, input *shared.PreToolUseHookInput) (*shared.SyncHookOutput, error) {
			return &shared.SyncHookOutput{}, nil
		}),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(len(opts.Hooks[shared.HookEventPreToolUse]))
	// Output: 1
}

// ExampleWithAgents demonstrates configuring custom subagents.
// Agent models must be: sonnet, opus, haiku, or inherit
func ExampleWithAgents() {
	agents := map[string]shared.AgentDefinition{
		"coder": {
			Model:       "sonnet",
			Tools:       []string{"Read", "Write", "Bash"},
			Description: "A coding specialist agent",
		},
		"reviewer": {
			Model:       "opus",
			Description: "A code review specialist",
		},
	}

	client, err := claude.NewClient(
		claude.WithAgents(agents),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(len(opts.Agents))
	// Output: 2
}

// ExampleWithMcpServers demonstrates configuring MCP servers.
func ExampleWithMcpServers() {
	servers := map[string]shared.McpServerConfig{
		"filesystem": shared.McpStdioServerConfig{
			Type:    "stdio",
			Command: "npx",
			Args:    []string{"-y", "@anthropic/mcp-server-filesystem"},
		},
	}

	client, err := claude.NewClient(
		claude.WithMcpServers(servers),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(len(opts.McpServers))
	// Output: 1
}

// ExampleWithSandboxSettings demonstrates sandbox configuration.
func ExampleWithSandboxSettings() {
	client, err := claude.NewClient(
		claude.WithSandboxSettings(&shared.SandboxSettings{
			Enabled:    true,
			Type:       "docker",
			WorkingDir: "/workspace",
		}),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.Sandbox.Enabled)
	// Output: true
}

// ExampleWithPluginConfig demonstrates full plugin configuration.
func ExampleWithPluginConfig() {
	client, err := claude.NewClient(
		claude.WithPluginConfig(&shared.SdkPluginConfig{
			Enabled:       true,
			PluginPath:    "/usr/local/lib/claude-plugins/analyzer.so",
			Config:        map[string]any{"log_level": "debug"},
			Timeout:       30 * time.Second,
			MaxConcurrent: 4,
		}),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.SdkPluginConfig.Enabled)
	// Output: true
}

// ExampleWithDebugWriter demonstrates capturing debug output.
func ExampleWithDebugWriter() {
	client, err := claude.NewClient(
		claude.WithDebugWriter(os.Stderr),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.DebugWriter != nil)
	// Output: true
}

// ExampleWithStderrCallback demonstrates stderr line callbacks.
func ExampleWithStderrCallback() {
	client, err := claude.NewClient(
		claude.WithStderrCallback(func(line string) {
			// Handle stderr output
		}),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.StderrCallback != nil)
	// Output: true
}

// ExampleWithSettingSources demonstrates controlling setting sources.
func ExampleWithSettingSources() {
	client, err := claude.NewClient(
		claude.WithSettingSources(shared.SettingSourceUser, shared.SettingSourceProject),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(len(opts.SettingSources))
	// Output: 2
}

// ExampleWithClient demonstrates the recommended client lifecycle pattern.
func ExampleWithClient() {
	ctx := context.Background()

	// WithClient handles connection and cleanup automatically
	// Note: This example won't actually connect without Claude CLI
	_ = claude.WithClient(ctx, func(c claude.Client) error {
		// Use the client here
		return nil
	}, claude.WithModel("claude-sonnet-4-20250514"))

	fmt.Println("Client lifecycle managed")
	// Output: Client lifecycle managed
}

// ExampleWithContextFiles demonstrates adding context files.
func ExampleWithContextFiles() {
	client, err := claude.NewClient(
		claude.WithContextFiles("main.go", "README.md"),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(len(opts.ContextFiles))
	// Output: 2
}

// ExampleWithBufferSize demonstrates setting the message buffer size.
func ExampleWithBufferSize() {
	client, err := claude.NewClient(
		claude.WithBufferSize(200),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.BufferSize)
	// Output: 200
}

// ExampleWithMaxMessages demonstrates setting the max messages limit.
func ExampleWithMaxMessages() {
	client, err := claude.NewClient(
		claude.WithMaxMessages(500),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.MaxMessages)
	// Output: 500
}

// ExampleWithTrace demonstrates enabling protocol tracing.
func ExampleWithTrace() {
	client, err := claude.NewClient(
		claude.WithTrace(true),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.Trace)
	// Output: true
}

// ExampleWithIncludePartialMessages demonstrates enabling partial message streaming.
func ExampleWithIncludePartialMessages() {
	client, err := claude.NewClient(
		claude.WithIncludePartialMessages(true),
	)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	opts := client.GetOptions()
	fmt.Println(opts.IncludePartialMessages)
	// Output: true
}
