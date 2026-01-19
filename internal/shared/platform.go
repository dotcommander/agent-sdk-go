package shared

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// GetDefaultCommand returns the default command name for the current platform.
func GetDefaultCommand() string {
	switch runtime.GOOS {
	case "windows":
		return "claude.exe"
	default:
		return "claude"
	}
}

// GetDefaultPath returns the default installation path for the current platform.
func GetDefaultPath() string {
	switch runtime.GOOS {
	case "darwin":
		// Check Homebrew prefix
		if brewPrefix, err := exec.LookPath("brew"); err == nil {
			output, err := exec.Command(brewPrefix, "--prefix").Output()
			if err == nil {
				prefix := strings.TrimSpace(string(output))
				return filepath.Join(prefix, "bin", GetDefaultCommand())
			}
		}
		// Default to /usr/local/bin
		return filepath.Join("/usr/local/bin", GetDefaultCommand())

	case "linux":
		// Common locations for Linux
		locations := []string{
			filepath.Join("/usr/local/bin", GetDefaultCommand()),
			filepath.Join("/usr/bin", GetDefaultCommand()),
			filepath.Join("/opt/claude/bin", GetDefaultCommand()),
		}

		for _, loc := range locations {
			if _, err := os.Stat(loc); err == nil {
				return loc
			}
		}

	case "windows":
		// Default to Program Files
		return filepath.Join(os.Getenv("ProgramFiles"), "Claude", GetDefaultCommand())
	}

	// Default fallback - just the command name
	return GetDefaultCommand()
}
