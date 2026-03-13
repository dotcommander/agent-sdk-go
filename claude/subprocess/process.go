package subprocess

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ProcessManager holds the subprocess handle and its I/O streams.
type ProcessManager struct {
	cmd        *exec.Cmd
	cliPath    string
	cliCommand string
	stdin      io.WriteCloser
	stdout     io.ReadCloser
	stderr     io.ReadCloser
	cwd        string
	env        map[string]string
}

// buildEnv builds the environment variables for the subprocess.
func (p *ProcessManager) buildEnv() []string {
	env := os.Environ()

	// Add custom environment variables with validation
	for k, v := range p.env {
		if isValidEnvVar(k, v) {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		} else if debugEnabled {
			log.Printf("claude-sdk: skipping invalid env var %q", k)
		}
	}

	return env
}

var envKeyPattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// isValidEnvVar checks if an environment variable key-value pair is safe to use.
func isValidEnvVar(k, v string) bool {
	keyValid := envKeyPattern.MatchString(k)
	// Value must not contain dangerous characters
	valueValid := !strings.ContainsAny(v, "\n\r\x00")
	return keyValid && valueValid
}

// isValidPrompt checks if a prompt string is safe to use as a CLI argument.
// Note: exec.Command properly escapes arguments, so most characters are safe.
// We only block null bytes which could cause issues.
func isValidPrompt(prompt string) bool {
	// Only block null bytes - exec.Command handles escaping properly
	return !strings.ContainsAny(prompt, "\x00")
}
