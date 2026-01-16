package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"agent-sdk-go/internal/claude/shared"
)

// DiscoveryResult contains information about the discovered Claude CLI.
type DiscoveryResult struct {
	Path    string
	Command string
	Version string
	Found   bool
}

// DiscoverCLI discovers the Claude CLI executable.
// Returns the path to the CLI, command name, and any error.
func DiscoverCLI(cliPath, cliCommand string) (*DiscoveryResult, error) {
	result := &DiscoveryResult{
		Command: cliCommand,
	}

	// If a specific path is provided, check if it exists
	if cliPath != "" {
		if _, err := os.Stat(cliPath); err != nil {
			return nil, shared.NewCLINotFoundError(cliPath, cliCommand)
		}
		result.Path = cliPath
		result.Found = true
		return result, nil
	}

	// Otherwise, try to discover in PATH
	path, err := exec.LookPath(cliCommand)
	if err != nil {
		return nil, shared.NewCLINotFoundError("", cliCommand)
	}

	result.Path = path
	result.Found = true

	// Try to get version info
	version, err := getCLIVersion(path)
	if err != nil {
		result.Version = "unknown"
	} else {
		result.Version = version
	}

	return result, nil
}

// getCLIVersion gets the version of the Claude CLI.
func getCLIVersion(cliPath string) (string, error) {
	cmd := exec.Command(cliPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	// Parse version from output (e.g., "claude 1.0.0" -> "1.0.0")
	versionStr := strings.TrimSpace(string(output))
	parts := strings.Split(versionStr, " ")
	if len(parts) > 1 {
		return parts[1], nil
	}

	return versionStr, nil
}

// ValidateCLI validates the Claude CLI installation.
// Returns an error if the CLI is not properly installed or configured.
func ValidateCLI(result *DiscoveryResult) error {
	if !result.Found {
		return fmt.Errorf("Claude CLI not found")
	}

	// Check if the CLI is executable
	if runtime.GOOS != "windows" {
		if info, err := os.Stat(result.Path); err != nil || info.Mode().Perm()&0111 == 0 {
			return fmt.Errorf("Claude CLI is not executable: %s", result.Path)
		}
	}

	// Try to run a simple command to verify it works
	cmd := exec.Command(result.Path, "--help")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Claude CLI --help failed: %w", err)
	}

	if len(output) == 0 {
		return fmt.Errorf("Claude CLI returned empty help output")
	}

	return nil
}

// GetDefaultCommand returns the default command name for the current platform.
// Delegates to shared.GetDefaultCommand() to avoid duplication.
func GetDefaultCommand() string {
	return shared.GetDefaultCommand()
}

// GetDefaultPath returns the default installation path for the current platform.
// Delegates to shared.GetDefaultPath() to avoid duplication.
func GetDefaultPath() string {
	return shared.GetDefaultPath()
}

// GetCommonPaths returns common paths where Claude CLI might be installed.
func GetCommonPaths() []string {
	var paths []string

	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			filepath.Join("/usr/local/bin", GetDefaultCommand()),
			filepath.Join("/opt/homebrew/bin", GetDefaultCommand()),
			filepath.Join("$HOME/.local/bin", GetDefaultCommand()),
			filepath.Join("$HOME/bin", GetDefaultCommand()),
		)

	case "linux":
		paths = append(paths,
			filepath.Join("/usr/local/bin", GetDefaultCommand()),
			filepath.Join("/usr/bin", GetDefaultCommand()),
			filepath.Join("/opt/claude/bin", GetDefaultCommand()),
			filepath.Join("$HOME/.local/bin", GetDefaultCommand()),
			filepath.Join("$HOME/bin", GetDefaultCommand()),
		)

	case "windows":
		paths = append(paths,
			filepath.Join(os.Getenv("ProgramFiles"), "Claude", GetDefaultCommand()),
			filepath.Join(os.Getenv("ProgramFiles(x86)"), "Claude", GetDefaultCommand()),
			filepath.Join(os.Getenv("APPDATA"), "Claude", GetDefaultCommand()),
		)
	}

	// Add PATH locations
	if path := os.Getenv("PATH"); path != "" {
		for dir := range strings.SplitSeq(path, string(os.PathListSeparator)) {
			paths = append(paths, filepath.Join(dir, GetDefaultCommand()))
		}
	}

	return paths
}

// FindInPATH searches for the CLI in PATH environment variable.
func FindInPATH(command string) (string, error) {
	return exec.LookPath(command)
}

// ExpandPath expands a path with environment variables and ~.
func ExpandPath(path string) string {
	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	// Expand environment variables
	if strings.Contains(path, "$") {
		path = os.ExpandEnv(path)
	}

	return path
}

// IsCLIAvailable checks if Claude CLI is available in the current environment.
func IsCLIAvailable() bool {
	_, err := DiscoverCLI("", GetDefaultCommand())
	return err == nil
}

// PrintDiscoveryInfo prints discovery information to stdout.
func PrintDiscoveryInfo(result *DiscoveryResult) {
	fmt.Printf("Claude CLI Discovery:\n")
	fmt.Printf("  Found: %t\n", result.Found)
	if result.Found {
		fmt.Printf("  Path: %s\n", result.Path)
		fmt.Printf("  Command: %s\n", result.Command)
		fmt.Printf("  Version: %s\n", result.Version)
	}
}

// GetSuggestedCommands returns suggested commands for installing Claude CLI.
func GetSuggestedCommands() []string {
	var commands []string

	switch runtime.GOOS {
	case "darwin":
		commands = append(commands,
			"brew install claude",
			"curl -fsSL https://anthropics.com/install-claude | sh",
		)

	case "linux":
		commands = append(commands,
			"curl -fsSL https://anthropics.com/install-claude | sh",
			"wget -qO - https://anthropics.com/install-claude | sh",
		)

	case "windows":
		commands = append(commands,
			"winget install anthropic.claude",
			"choco install claude",
		)
	}

	// Add manual installation options
	commands = append(commands,
		"Download from: https://github.com/anthropics/claude-cli/releases",
		"Add to PATH after installation",
	)

	return commands
}