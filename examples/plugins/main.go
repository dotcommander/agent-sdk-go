// Package main demonstrates plugin configuration with the Claude Agent SDK.
//
// Plugins allow you to:
// - Extend Claude's capabilities with local plugins
// - Configure plugin paths and types
// - Manage plugin lifecycle
//
// Based on TypeScript example: https://github.com/anthropics/claude-agent-sdk-demos
package main

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

func main() {
	fmt.Println("=== Plugins Example ===")
	fmt.Println("This demonstrates configuring and using plugins.")
	fmt.Println()

	// Show basic plugin configuration
	demonstrateBasicPlugins()

	// Show plugin options
	demonstratePluginOptions()

	// Show multiple plugins
	demonstrateMultiplePlugins()

	// Show plugin patterns
	demonstratePluginPatterns()

	fmt.Println()
	fmt.Println("=== Plugins Example Complete ===")
}

// demonstrateBasicPlugins shows basic plugin configuration.
func demonstrateBasicPlugins() {
	fmt.Println("--- Basic Plugin Configuration ---")
	fmt.Println()

	// The actual PluginConfig struct uses Type and Path
	plugin := shared.PluginConfig{
		Type: "local",
		Path: "/path/to/code-analysis-plugin",
	}

	printJSON("Plugin Config", plugin)

	fmt.Println(`
  // Register plugins with client
  client, err := claude.NewClient(
      claude.WithPlugins([]shared.PluginConfig{plugin}),
  )
`)
}

// demonstratePluginOptions shows various plugin configuration options.
func demonstratePluginOptions() {
	fmt.Println("--- Plugin Configuration Options ---")
	fmt.Println()

	fmt.Println("1. Local plugin:")
	localPlugin := shared.PluginConfig{
		Type: "local",
		Path: "/home/user/plugins/my-plugin",
	}
	printJSON("Local Plugin", localPlugin)

	fmt.Println("2. Multiple local plugins:")
	plugins := []shared.PluginConfig{
		{Type: "local", Path: "/plugins/security-scanner"},
		{Type: "local", Path: "/plugins/code-formatter"},
		{Type: "local", Path: "/plugins/documentation-generator"},
	}
	printJSON("Multiple Plugins", plugins)

	fmt.Println("Plugin configuration is passed via CLI flags or config files.")
	fmt.Println("The SDK passes these to the Claude CLI subprocess.")
}

// demonstrateMultiplePlugins shows configuring multiple plugins.
func demonstrateMultiplePlugins() {
	fmt.Println("--- Multiple Plugins ---")
	fmt.Println()

	plugins := []shared.PluginConfig{
		{Type: "local", Path: "/plugins/security-scanner"},
		{Type: "local", Path: "/plugins/code-formatter"},
		{Type: "local", Path: "/plugins/documentation-generator"},
	}

	fmt.Println("Configured Plugins:")
	for _, p := range plugins {
		fmt.Printf("  - Type: %s, Path: %s\n", p.Type, p.Path)
	}
	fmt.Println()

	fmt.Println(`
  // Register all plugins
  client, err := claude.NewClient(
      claude.WithModel("claude-sonnet-4-20250514"),
      claude.WithPlugins(plugins),
  )
`)
}

// demonstratePluginPatterns shows common plugin patterns.
func demonstratePluginPatterns() {
	fmt.Println("--- Plugin Patterns ---")
	fmt.Println()

	fmt.Println("1. Environment-aware plugins:")
	fmt.Println(`
  func getPlugins(env string) []shared.PluginConfig {
      base := []shared.PluginConfig{
          {Type: "local", Path: "/plugins/logging"},
          {Type: "local", Path: "/plugins/metrics"},
      }

      if env == "development" {
          base = append(base, shared.PluginConfig{
              Type: "local",
              Path: "/plugins/debug-tools",
          })
      }

      if env == "production" {
          base = append(base, shared.PluginConfig{
              Type: "local",
              Path: "/plugins/performance-monitor",
          })
      }

      return base
  }
`)

	fmt.Println("2. Plugin discovery from directory:")
	fmt.Println(`
  func discoverPlugins(dir string) ([]shared.PluginConfig, error) {
      entries, err := os.ReadDir(dir)
      if err != nil {
          return nil, err
      }

      var plugins []shared.PluginConfig
      for _, entry := range entries {
          if entry.IsDir() {
              plugins = append(plugins, shared.PluginConfig{
                  Type: "local",
                  Path: filepath.Join(dir, entry.Name()),
              })
          }
      }
      return plugins, nil
  }

  // Usage
  plugins, _ := discoverPlugins("/home/user/claude-plugins")
  client, _ := claude.NewClient(claude.WithPlugins(plugins))
`)

	fmt.Println("3. Dynamic plugin configuration from file:")
	fmt.Println(`
  // plugins.json:
  // {
  //   "plugins": [
  //     {"type": "local", "path": "/plugins/scanner"},
  //     {"type": "local", "path": "/plugins/formatter"}
  //   ]
  // }

  func loadPluginConfig(path string) ([]shared.PluginConfig, error) {
      data, err := os.ReadFile(path)
      if err != nil {
          return nil, err
      }

      var config struct {
          Plugins []shared.PluginConfig ` + "`json:\"plugins\"`" + `
      }
      if err := json.Unmarshal(data, &config); err != nil {
          return nil, err
      }

      return config.Plugins, nil
  }

  // Usage
  plugins, _ := loadPluginConfig("plugins.json")
  client, _ := claude.NewClient(claude.WithPlugins(plugins))
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
