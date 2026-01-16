// This file is deprecated. Use cmd/agent/main.go instead.
// The V2 SDK API is available via internal/claude/v2 package.

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"agent-sdk-go/internal/app"
	"agent-sdk-go/internal/claude/v2"
)

func main() {
	// Load configuration from environment variables and config files
	cfg, err := app.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	// Use V2 SDK for one-shot prompt
	ctx := context.Background()

	opts := []v2.PromptOption{
		v2.WithPromptModel(cfg.GetModel("claude-3-5-sonnet-20241022")),
	}

	if cfg.APIKey != "" {
		// API key is passed via environment variable
		// Use WithPromptEnv to set it for the subprocess
		opts = append(opts, v2.WithPromptEnv(map[string]string{"ANTHROPIC_API_KEY": cfg.APIKey}))
	}

	if cfg.Timeout > 0 {
		opts = append(opts, v2.WithPromptTimeout(time.Duration(cfg.Timeout)*time.Second))
	}

	result, err := v2.Prompt(ctx, "Hello, Claude!", opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "prompt: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(result.Result)
}
