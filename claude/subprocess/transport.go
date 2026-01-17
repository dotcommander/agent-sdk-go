// Package subprocess provides subprocess communication with the Claude CLI.
// It implements both interactive mode (for multi-turn sessions) and one-shot mode
// (for single prompts with the prompt passed as a CLI argument).
package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"agent-sdk-go/claude/cli"
	"agent-sdk-go/claude/parser"
	"agent-sdk-go/claude/shared"
)

const (
	// defaultTimeout is the default timeout for subprocess operations.
	defaultTimeout = 60 * time.Second
	// channelBufferSize is the buffer size for message and error channels.
	channelBufferSize = 100
	// maxRetries is the maximum number of retry attempts for transient failures.
	maxRetries = 3
	// baseDelay is the base delay for exponential backoff (100ms).
	baseDelay = 100 * time.Millisecond
)

// withRetry executes a function with exponential backoff retry logic.
// Retries transient failures (connection errors, timeouts, process errors).
func withRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil // Success
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			return fmt.Errorf("%s: non-retryable error: %w", operation, err)
		}

		lastErr = fmt.Errorf("%s (attempt %d/%d): %w", operation, attempt+1, maxRetries, err)

		// Calculate delay with exponential backoff and jitter
		if attempt < maxRetries-1 {
			delay := calculateDelay(attempt)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operation, maxRetries, lastErr)
}

// isRetryableError determines if an error should be retried.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check specific error types that are retryable
	if shared.IsConnectionError(err) {
		return true
	}

	if shared.IsTimeoutError(err) {
		return true
	}

	// Check for process-related errors
	if _, ok := err.(*shared.ProcessError); ok {
		return true
	}

	// Check for transient I/O errors
	if strings.Contains(err.Error(), "resource temporarily unavailable") ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "EOF") {
		return true
	}

	return false
}

// calculateDelay calculates the delay with exponential backoff and jitter.
func calculateDelay(attempt int) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	delay := float64(baseDelay) * math.Pow(2, float64(attempt))

	// Add jitter (random factor between 0.5x and 1.5x)
	jitter := 0.5 + rand.Float64()*1.0
	delay = delay * jitter

	// Cap at 5 seconds to avoid excessive delays
	cappedDelay := time.Duration(delay)
	if cappedDelay > 5*time.Second {
		cappedDelay = 5 * time.Second
	}

	return cappedDelay
}

// isValidEnvVar checks if an environment variable key-value pair is safe to use.
func isValidEnvVar(k, v string) bool {
	// Key must be valid shell identifier
	keyValid := regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`).MatchString(k)
	// Value must not contain dangerous characters
	valueValid := !strings.ContainsAny(v, "\n\r\x00")
	return keyValid && valueValid
}

// isValidPrompt checks if a prompt string is safe to use as a CLI argument.
func isValidPrompt(prompt string) bool {
	// Prompt must not contain shell escape characters
	return !strings.ContainsAny(prompt, "`$!;&|<>")
}

// Transport represents a subprocess transport for communicating with Claude CLI.
type Transport struct {
	// Process management
	cmd       *exec.Cmd
	cliPath   string
	cliCommand string
	promptArg *string // nil = interactive mode, set = one-shot mode

	// Connection state
	connected bool
	mu        sync.RWMutex

	// I/O streams
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	// Configuration
	timeout      time.Duration
	model        string
	systemPrompt string
	customArgs   []string
	env          map[string]string
	mcpServers   map[string]shared.McpServerConfig

	// Parser registry for message type handling (OCP compliance - inject instead of switch)
	parserRegistry *parser.MessageParserRegistry

	// Channels for communication
	msgChan chan shared.Message
	errChan chan error

	// Control and cleanup
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// TransportConfig holds configuration for the transport.
type TransportConfig struct {
	// CLIPath is the path to the Claude CLI executable. If empty, will be discovered.
	CLIPath string
	// CLICommand is the command name to use (e.g., "claude").
	CLICommand string
	// Model is the Claude model to use.
	Model string
	// Timeout is the timeout for operations.
	Timeout time.Duration
	// SystemPrompt is the system prompt to use.
	SystemPrompt string
	// CustomArgs are additional CLI arguments.
	CustomArgs []string
	// Env are environment variables to set for the subprocess.
	Env map[string]string
	// PromptArg is the prompt for one-shot mode. If set, transport operates in one-shot mode.
	PromptArg *string
	// ParserRegistry is the registry for message type parsers (OCP compliance).
	// If nil, the default registry is used.
	ParserRegistry *parser.MessageParserRegistry
	// McpServers are MCP server configurations.
	McpServers map[string]shared.McpServerConfig
}

// createTransport creates a new transport with common initialization logic.
func createTransport(config *TransportConfig, promptArg *string) (*Transport, error) {
	if config == nil {
		config = &TransportConfig{}
	}

	// Set defaults
	if config.CLICommand == "" {
		config.CLICommand = cli.GetDefaultCommand()
	}
	if config.Model == "" {
		config.Model = "claude-sonnet-4-5-20250929"
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}

	// Use provided registry or default
	registry := config.ParserRegistry
	if registry == nil {
		registry = parser.DefaultRegistry()
	}

	return &Transport{
		cliPath:        config.CLIPath,
		cliCommand:     config.CLICommand,
		model:          config.Model,
		timeout:        config.Timeout,
		systemPrompt:   config.SystemPrompt,
		customArgs:     config.CustomArgs,
		env:            config.Env,
		mcpServers:     config.McpServers,
		promptArg:      promptArg,
		parserRegistry: registry,
	}, nil
}

// NewTransport creates a new subprocess transport in interactive mode.
// The transport will communicate with the CLI via stdin/stdout for multi-turn conversations.
func NewTransport(config *TransportConfig) (*Transport, error) {
	return createTransport(config, nil) // Interactive mode
}

// NewTransportWithPrompt creates a new subprocess transport in one-shot mode.
// The prompt is passed as a CLI argument, and the response is read from stdout.
func NewTransportWithPrompt(config *TransportConfig, prompt string) (*Transport, error) {
	// Validate the prompt input for security
	if !isValidPrompt(prompt) {
		return nil, fmt.Errorf("invalid prompt: contains potentially dangerous characters")
	}

	return createTransport(config, &prompt) // One-shot mode
}

// Connect starts the Claude CLI subprocess and establishes communication.
func (t *Transport) Connect(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.connected {
		return fmt.Errorf("transport already connected")
	}

	return withRetry(ctx, "connect", func() error {
		// Discover CLI path if not provided
		cliPath := t.cliPath
		if cliPath == "" {
			result, err := cli.DiscoverCLI("", t.cliCommand)
			if err != nil {
				return fmt.Errorf("discover CLI: %w", err)
			}
			cliPath = result.Path
		}

		// Build command arguments
		args := t.buildArgs()
		t.cmd = exec.CommandContext(ctx, cliPath, args...)

		// Set environment
		t.cmd.Env = t.buildEnv()

		// Set up I/O pipes
		// Only create stdin pipe for interactive mode - stdin pipe causes issues with one-shot mode
		if t.promptArg == nil {
			stdin, err := t.cmd.StdinPipe()
			if err != nil {
				return fmt.Errorf("create stdin pipe: %w", err)
			}
			t.stdin = stdin
		}

		stdout, err := t.cmd.StdoutPipe()
		if err != nil {
			if t.stdin != nil {
				t.stdin.Close()
			}
			return fmt.Errorf("create stdout pipe: %w", err)
		}
		t.stdout = stdout

		stderr, err := t.cmd.StderrPipe()
		if err != nil {
			if t.stdin != nil {
				t.stdin.Close()
			}
			stdout.Close()
			return fmt.Errorf("create stderr pipe: %w", err)
		}
		t.stderr = stderr

		// Initialize channels
		t.msgChan = make(chan shared.Message, channelBufferSize)
		t.errChan = make(chan error, channelBufferSize)

		// Start the process
		if err := t.cmd.Start(); err != nil {
			t.cleanup()
			return fmt.Errorf("start CLI process: %w", err)
		}

		// Set up context for goroutine management
		t.ctx, t.cancel = context.WithCancel(ctx)

		// Start stdout reader goroutine
		t.wg.Add(1)
		go t.handleStdout()

		// Start stderr reader goroutine for error reporting
		t.wg.Add(1)
		go t.handleStderr()

		t.connected = true
		return nil
	})
}

// buildArgs builds the CLI arguments based on the transport mode.
func (t *Transport) buildArgs() []string {
	var args []string

	if t.promptArg != nil {
		// One-shot mode: -p flag enables print mode, prompt is positional arg at end
		// --verbose is required for stream-json output in print mode
		args = append(args, "-p", "--output-format", "stream-json", "--verbose")
	} else {
		// Interactive mode: use streaming JSON for both input and output
		args = append(args, "--output-format", "stream-json", "--input-format", "stream-json")
	}

	// Add model
	args = append(args, "--model", t.model)

	// Add system prompt if set
	if t.systemPrompt != "" {
		args = append(args, "--system-prompt", t.systemPrompt)
	}

	// Add custom args
	args = append(args, t.customArgs...)

	// MCP servers
	if len(t.mcpServers) > 0 {
		serversForCLI := make(map[string]interface{})
		for name, config := range t.mcpServers {
			if sdkConfig, ok := config.(shared.McpSdkServerConfig); ok {
				// For SDK servers, pass everything except instance
				serversForCLI[name] = map[string]interface{}{
					"type": sdkConfig.Type,
					"name": sdkConfig.Name,
				}
			} else {
				serversForCLI[name] = config
			}
		}
		if len(serversForCLI) > 0 {
			mcpJSON, _ := json.Marshal(map[string]interface{}{"mcpServers": serversForCLI})
			args = append(args, "--mcp-config", string(mcpJSON))
		}
	}

	// In one-shot mode, prompt goes last as positional argument
	if t.promptArg != nil {
		args = append(args, *t.promptArg)
	}

	return args
}

// buildEnv builds the environment variables for the subprocess.
func (t *Transport) buildEnv() []string {
	env := os.Environ()

	// Add custom environment variables with validation
	// Invalid environment variables are silently skipped for security
	for k, v := range t.env {
		if isValidEnvVar(k, v) {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return env
}

// handleStdout reads and parses messages from stdout.
func (t *Transport) handleStdout() {
	defer t.wg.Done()
	defer func() {
		// Signal completion when stdout closes (subprocess exit)
		// This unblocks any goroutines waiting on msgChan
		t.mu.Lock()
		if t.connected {
			close(t.msgChan)
			close(t.errChan)
			t.connected = false
		}
		t.mu.Unlock()
	}()

	scanner := bufio.NewScanner(t.stdout)
	for scanner.Scan() {
		// Check for context cancellation
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse the line as JSON
		var rawMsg map[string]any
		if err := json.Unmarshal([]byte(line), &rawMsg); err != nil {
			select {
			case t.errChan <- fmt.Errorf("parse JSON: %w", err):
			case <-t.ctx.Done():
				return
			}
			continue
		}

		// Discriminate by message type
		msgType, ok := rawMsg["type"].(string)
		if !ok {
			select {
			case t.errChan <- fmt.Errorf("message missing type field"):
			case <-t.ctx.Done():
				return
			}
			continue
		}

		// Parse message using injected registry (OCP compliance)
		msg, err := t.parserRegistry.Parse(msgType, line, 0)
		if err != nil {
			// Check if it's an unknown type - pass as raw
			if !t.parserRegistry.HasParser(msgType) {
				msg = &shared.RawControlMessage{
					MessageType: msgType,
					Data:        rawMsg,
				}
			} else {
				select {
				case t.errChan <- err:
				case <-t.ctx.Done():
					return
				}
				continue
			}
		}

		// Send the message
		select {
		case t.msgChan <- msg:
		case <-t.ctx.Done():
			return
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		select {
		case t.errChan <- fmt.Errorf("stdout scanner error: %w", err):
		case <-t.ctx.Done():
		}
	}
}

// handleStderr reads from stderr and forwards to error channel.
func (t *Transport) handleStderr() {
	defer t.wg.Done()

	// Drain stderr to prevent blocking
	scanner := bufio.NewScanner(t.stderr)
	for {
		select {
		case <-t.ctx.Done():
			return
		default:
			if !scanner.Scan() {
				// Scanner reached EOF
				return
			}

			line := scanner.Text()
			if line != "" {
				select {
				case t.errChan <- fmt.Errorf("CLI stderr: %s", line):
				case <-t.ctx.Done():
					return
				}
			}
		}
	}
}

// SendMessage sends a message to the CLI via stdin.
// Only works in interactive mode (promptArg == nil).
func (t *Transport) SendMessage(ctx context.Context, message string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return fmt.Errorf("transport not connected")
	}

	if t.promptArg != nil {
		return fmt.Errorf("cannot send message in one-shot mode")
	}

	// Create a user message from the string
	userMsg := &shared.UserMessage{
		MessageType: shared.MessageTypeUser,
		Content:     message,
	}

	// Marshal the message to JSON
	data, err := json.Marshal(userMsg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	// Write to stdin with newline
	if _, err := fmt.Fprintln(t.stdin, string(data)); err != nil {
		return fmt.Errorf("write to stdin: %w", err)
	}

	return nil
}

// SendText sends a text message to the CLI via stdin.
// This is a convenience method for sending simple text messages.
func (t *Transport) SendText(ctx context.Context, text string) error {
	return t.SendMessage(ctx, text)
}

// ReceiveMessages returns channels for receiving messages and errors.
func (t *Transport) ReceiveMessages(ctx context.Context) (<-chan shared.Message, <-chan error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.connected {
		msgChan := make(chan shared.Message)
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("transport not connected")
		close(msgChan)
		close(errChan)
		return msgChan, errChan
	}

	return t.msgChan, t.errChan
}

// Close closes the transport and cleans up resources.
func (t *Transport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return nil
	}

	t.connected = false

	// Cancel context to stop goroutines
	if t.cancel != nil {
		t.cancel()
	}

	// Close stdin to signal EOF to the subprocess
	if t.stdin != nil {
		_ = t.stdin.Close()
	}

	// Wait for goroutines to finish
	done := make(chan struct{})
	go func() {
		t.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		// Goroutines finished
	case <-time.After(5 * time.Second):
		// Timeout - force cleanup by terminating goroutines
		// This prevents resource leaks when goroutines get stuck
		t.cancel()
		// Give goroutines a moment to clean up
		time.Sleep(1 * time.Second)
	}

	// Close stdout and stderr
	if t.stdout != nil {
		_ = t.stdout.Close()
	}
	if t.stderr != nil {
		_ = t.stderr.Close()
	}

	// Terminate the process
	if t.cmd != nil && t.cmd.Process != nil {
		_ = t.cmd.Process.Kill()
		_ = t.cmd.Wait()
	}

	// Note: channels are closed by handleStdout when it exits
	// to avoid double-close panic

	return nil
}

// IsConnected returns whether the transport is connected.
func (t *Transport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// GetPID returns the process ID of the CLI subprocess.
func (t *Transport) GetPID() int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.cmd != nil && t.cmd.Process != nil {
		return t.cmd.Process.Pid
	}
	return 0
}

// GetCommand returns the command used to start Claude CLI.
func (t *Transport) GetCommand() string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.cmd != nil {
		return t.cmd.String()
	}

	if t.cliPath != "" {
		return fmt.Sprintf("%s %s", t.cliPath, strings.Join(t.buildArgs(), " "))
	}

	return t.cliCommand
}

// Interrupt forcibly interrupts the transport.
func (t *Transport) Interrupt() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected || t.cmd == nil || t.cmd.Process == nil {
		return fmt.Errorf("process not running")
	}

	// Close stdin to signal EOF
	if t.stdin != nil {
		_ = t.stdin.Close()
	}

	// Try to terminate the process gracefully
	return t.cmd.Process.Kill()
}

// cleanup closes all resources without acquiring the lock.
// The caller must hold t.mu.
func (t *Transport) cleanup() {
	if t.stdin != nil {
		_ = t.stdin.Close()
	}
	if t.stdout != nil {
		_ = t.stdout.Close()
	}
	if t.stderr != nil {
		_ = t.stderr.Close()
	}
	if t.msgChan != nil {
		close(t.msgChan)
	}
	if t.errChan != nil {
		close(t.errChan)
	}
}