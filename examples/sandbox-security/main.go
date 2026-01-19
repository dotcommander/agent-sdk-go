// Package main demonstrates sandbox security settings with the Claude Agent SDK.
//
// Sandbox settings allow you to:
// - Restrict file system access
// - Control network access
// - Configure command execution limits
// - Implement defense in depth
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos
package main

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude"
)

func main() {
	fmt.Println("=== Sandbox Security Example ===")
	fmt.Println("This demonstrates configuring sandbox security settings.")
	fmt.Println()

	// Show basic sandbox configuration
	demonstrateBasicSandbox()

	// Show sandbox types
	demonstrateSandboxTypes()

	// Show integration with hooks
	demonstrateHookBasedSecurity()

	// Show security patterns
	demonstrateSecurityPatterns()

	fmt.Println()
	fmt.Println("=== Sandbox Security Example Complete ===")
}

// demonstrateBasicSandbox shows basic sandbox configuration.
func demonstrateBasicSandbox() {
	fmt.Println("--- Basic Sandbox Configuration ---")
	fmt.Println()

	// The actual SandboxSettings struct
	sandbox := claude.SandboxSettings{
		Enabled:    true,
		Type:       "docker",
		Image:      "claude-sandbox:latest",
		WorkingDir: "/workspace",
		Options: map[string]string{
			"network":   "none",
			"read-only": "true",
		},
	}

	printJSON("Sandbox Settings", sandbox)

	fmt.Println(`
  // Configure client with sandbox
  client, err := claude.NewClient(
      claude.WithModel("claude-sonnet-4-20250514"),
      claude.WithSandbox(sandbox),
  )
`)
}

// demonstrateSandboxTypes shows different sandbox types.
func demonstrateSandboxTypes() {
	fmt.Println("--- Sandbox Types ---")
	fmt.Println()

	fmt.Println("1. Docker sandbox:")
	dockerSandbox := claude.SandboxSettings{
		Enabled: true,
		Type:    "docker",
		Image:   "claude-sandbox:latest",
		Options: map[string]string{
			"memory":    "512m",
			"cpus":      "0.5",
			"network":   "none",
			"read-only": "true",
		},
		WorkingDir: "/workspace",
	}
	printJSON("Docker Sandbox", dockerSandbox)

	fmt.Println("2. nsjail sandbox:")
	nsjailSandbox := claude.SandboxSettings{
		Enabled: true,
		Type:    "nsjail",
		Options: map[string]string{
			"time_limit": "60",
			"max_cpus":   "1",
			"rlimit_as":  "512",
		},
		WorkingDir: "/tmp/sandbox",
	}
	printJSON("nsjail Sandbox", nsjailSandbox)

	fmt.Println("3. Custom sandbox:")
	customSandbox := claude.SandboxSettings{
		Enabled: true,
		Type:    "custom",
		Options: map[string]string{
			"command":     "/usr/local/bin/my-sandbox",
			"config_file": "/etc/sandbox.conf",
		},
	}
	printJSON("Custom Sandbox", customSandbox)
}

// demonstrateHookBasedSecurity shows using hooks for security.
func demonstrateHookBasedSecurity() {
	fmt.Println("--- Hook-Based Security ---")
	fmt.Println()

	fmt.Println("Use PreToolUse hooks for fine-grained security control:")
	fmt.Println(`
  // Block dangerous operations via hooks
  hooks := map[claude.HookEvent][]claude.HookConfig{
      claude.HookEventPreToolUse: {
          {
              Matcher: "Bash",
              Handler: func(ctx context.Context, input *claude.PreToolUseHookInput) (*claude.SyncHookOutput, error) {
                  cmd := input.ToolInput["command"].(string)

                  // Block dangerous commands
                  dangerous := []string{"rm -rf", "sudo", "chmod 777", "> /dev"}
                  for _, d := range dangerous {
                      if strings.Contains(cmd, d) {
                          return &claude.SyncHookOutput{
                              Decision:   "block",
                              StopReason: "Dangerous command blocked",
                          }, nil
                      }
                  }

                  return &claude.SyncHookOutput{Continue: true}, nil
              },
          },
      },
  }
`)

	fmt.Println("Restrict file access via hooks:")
	fmt.Println(`
  hooks := map[claude.HookEvent][]claude.HookConfig{
      claude.HookEventPreToolUse: {
          {
              Matcher: "Read|Write|Edit",
              Handler: func(ctx context.Context, input *claude.PreToolUseHookInput) (*claude.SyncHookOutput, error) {
                  path := input.ToolInput["file_path"].(string)
                  allowedPaths := []string{"/home/user/project", "/tmp"}

                  allowed := false
                  for _, ap := range allowedPaths {
                      if strings.HasPrefix(path, ap) {
                          allowed = true
                          break
                      }
                  }

                  if !allowed {
                      return &claude.SyncHookOutput{
                          Decision:   "block",
                          StopReason: fmt.Sprintf("Access denied: %s", path),
                      }, nil
                  }

                  return &claude.SyncHookOutput{Continue: true}, nil
              },
          },
      },
  }
`)
}

// demonstrateSecurityPatterns shows common security patterns.
func demonstrateSecurityPatterns() {
	fmt.Println("--- Security Patterns ---")
	fmt.Println()

	fmt.Println("1. Defense in depth:")
	fmt.Println(`
  // Layer 1: Sandbox container
  sandbox := claude.SandboxSettings{
      Enabled: true,
      Type:    "docker",
      Options: map[string]string{
          "network": "none",
      },
  }

  // Layer 2: Hook-based validation
  hooks := createSecurityHooks()

  // Layer 3: Permission mode
  permissionMode := claude.PermissionModeDefault

  // Layer 4: Tool restrictions
  disallowedTools := []string{"Bash"}

  client, err := claude.NewClient(
      claude.WithSandbox(sandbox),
      claude.WithHooks(hooks),
      claude.WithPermissionMode(permissionMode),
      claude.WithDisallowedTools(disallowedTools),
  )
`)

	fmt.Println("2. Environment-based security:")
	fmt.Println(`
  func getSandbox(env string) claude.SandboxSettings {
      base := claude.SandboxSettings{
          Enabled: true,
          Type:    "docker",
      }

      switch env {
      case "development":
          base.Options = map[string]string{
              "network": "bridge",  // Allow network
          }

      case "staging":
          base.Options = map[string]string{
              "network":   "none",
              "read-only": "false",
          }

      case "production":
          base.Options = map[string]string{
              "network":   "none",
              "read-only": "true",
              "no-new-privileges": "true",
          }
      }

      return base
  }
`)

	fmt.Println("3. Audit logging:")
	fmt.Println(`
  // Log all tool usage for security audit
  hooks := map[claude.HookEvent][]claude.HookConfig{
      claude.HookEventPostToolUse: {
          {
              Matcher: "",  // Match all tools
              Handler: func(ctx context.Context, input *claude.PostToolUseHookInput) (*claude.SyncHookOutput, error) {
                  // Log to audit trail
                  auditLog.Info("tool_executed",
                      "tool", input.ToolName,
                      "session", input.SessionID,
                      "input", input.ToolInput,
                  )
                  return &claude.SyncHookOutput{Continue: true}, nil
              },
          },
      },
      claude.HookEventPostToolUseFailure: {
          {
              Matcher: "",
              Handler: func(ctx context.Context, input *claude.PostToolUseFailureHookInput) (*claude.SyncHookOutput, error) {
                  // Log failures
                  auditLog.Warn("tool_failed",
                      "tool", input.ToolName,
                      "error", input.Error,
                  )
                  return &claude.SyncHookOutput{Continue: true}, nil
              },
          },
      },
  }
`)
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
