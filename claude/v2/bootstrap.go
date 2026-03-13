package v2

import (
	"fmt"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude"
	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

// buildClient performs the common bootstrap sequence shared by session and prompt creation:
// CLI availability check, factory resolution, and client construction.
func buildClient(model string, timeout time.Duration, cliChecker shared.CLIChecker, factory ClientFactory) (claude.Client, error) {
	if cliChecker == nil {
		cliChecker = shared.CLICheckerFunc(cli.IsCLIAvailable)
	}
	if !cliChecker.IsCLIAvailable() {
		return nil, fmt.Errorf("claude CLI not found. Please install it first")
	}

	if factory == nil {
		factory = DefaultClientFactory()
	}

	client, err := factory.NewClient(
		claude.WithModel(model),
		claude.WithTimeout(timeout.String()),
	)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	return client, nil
}
