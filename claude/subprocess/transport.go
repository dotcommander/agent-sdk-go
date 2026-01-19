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
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/cli"
	"github.com/dotcommander/agent-sdk-go/claude/mcp"
	"github.com/dotcommander/agent-sdk-go/claude/parser"
	"github.com/dotcommander/agent-sdk-go/internal/shared"
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
// Note: exec.Command properly escapes arguments, so most characters are safe.
// We only block null bytes which could cause issues.
func isValidPrompt(prompt string) bool {
	// Only block null bytes - exec.Command handles escaping properly
	return !strings.ContainsAny(prompt, "\x00")
}

// Transport represents a subprocess transport for communicating with Claude CLI.
type Transport struct {
	// Process management
	cmd        *exec.Cmd
	cliPath    string
	cliCommand string
	promptArg  *string // nil = interactive mode, set = one-shot mode

	// Connection state
	connected bool
	mu        sync.RWMutex

	// I/O streams
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser

	// Configuration
	timeout        time.Duration
	model          string
	systemPrompt   string
	customArgs     []string
	env            map[string]string
	cwd            string
	tools          *shared.ToolsConfig
	stderrCallback func(line string)
	canUseTool     shared.CanUseToolCallback
	mcpServers     map[string]shared.McpServerConfig

	// Parser registry for message type handling (OCP compliance - inject instead of switch)
	parserRegistry *parser.MessageParserRegistry

	// Control protocol for bidirectional communication
	protocol        *Protocol
	protocolAdapter *ProtocolAdapter

	// Control protocol configuration
	protocolHooks         map[shared.HookEvent][]ProtocolHookMatcher
	sdkMcpServers         map[string]*mcp.SdkMcpServer
	enableCheckpointing   bool
	enableControlProtocol bool // Explicitly enable control protocol

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
	// Cwd is the working directory for the subprocess.
	// If empty, inherits parent process working directory.
	Cwd string
	// Tools configures tool availability (preset or explicit list).
	Tools *shared.ToolsConfig
	// StderrCallback is invoked for each stderr line from subprocess.
	// If nil, stderr goes to parent process stderr.
	StderrCallback func(line string)
	// CanUseTool is the callback for runtime permission checks.
	CanUseTool shared.CanUseToolCallback
	// PromptArg is the prompt for one-shot mode. If set, transport operates in one-shot mode.
	PromptArg *string
	// ParserRegistry is the registry for message type parsers (OCP compliance).
	// If nil, the default registry is used.
	ParserRegistry *parser.MessageParserRegistry
	// McpServers are MCP server configurations.
	McpServers map[string]shared.McpServerConfig

	// Control protocol configuration
	// ProtocolHooks configures hook callbacks for the control protocol.
	ProtocolHooks map[shared.HookEvent][]ProtocolHookMatcher
	// SdkMcpServers configures in-process MCP servers for control protocol routing.
	SdkMcpServers map[string]*mcp.SdkMcpServer
	// EnableCheckpointing enables file checkpointing for RewindFiles.
	EnableCheckpointing bool
	// EnableControlProtocol explicitly enables the control protocol.
	// If false, control protocol is auto-enabled when hooks, permissions, or checkpointing are configured.
	EnableControlProtocol bool
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
		cliPath:               config.CLIPath,
		cliCommand:            config.CLICommand,
		model:                 config.Model,
		timeout:               config.Timeout,
		systemPrompt:          config.SystemPrompt,
		customArgs:            config.CustomArgs,
		env:                   config.Env,
		cwd:                   config.Cwd,
		tools:                 config.Tools,
		stderrCallback:        config.StderrCallback,
		canUseTool:            config.CanUseTool,
		mcpServers:            config.McpServers,
		promptArg:             promptArg,
		parserRegistry:        registry,
		protocolHooks:         config.ProtocolHooks,
		sdkMcpServers:         config.SdkMcpServers,
		enableCheckpointing:   config.EnableCheckpointing,
		enableControlProtocol: config.EnableControlProtocol,
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

		// Validate and set working directory if specified
		if t.cwd != "" {
			// Ensure path is absolute
			if !filepath.IsAbs(t.cwd) {
				return fmt.Errorf("cwd must be an absolute path: %s", t.cwd)
			}
			// Verify directory exists
			info, err := os.Stat(t.cwd)
			if err != nil {
				if os.IsNotExist(err) {
					return fmt.Errorf("cwd directory does not exist: %s", t.cwd)
				}
				return fmt.Errorf("cwd stat error: %w", err)
			}
			if !info.IsDir() {
				return fmt.Errorf("cwd is not a directory: %s", t.cwd)
			}
		}

		// Build command arguments
		args := t.buildArgs()
		t.cmd = exec.CommandContext(ctx, cliPath, args...)

		// Set working directory if specified
		if t.cwd != "" {
			t.cmd.Dir = t.cwd
		}

		// Set environment
		t.cmd.Env = t.buildEnv()

		// Debug: log command and relevant env vars to file
		if debugFile, err := os.OpenFile("/tmp/sdk-debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(debugFile, "\n=== %s ===\n", time.Now().Format(time.RFC3339))
			fmt.Fprintf(debugFile, "command: %s %v\n", cliPath, args)
			for _, e := range t.cmd.Env {
				if strings.HasPrefix(e, "ANTHROPIC_") || strings.HasPrefix(e, "SYNTHETIC_") || strings.HasPrefix(e, "ZAI_") {
					fmt.Fprintf(debugFile, "env: %s\n", e)
				}
			}
			debugFile.Close()
		}

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

		// Set up control protocol if needed
		if t.needsProtocolHandshake() {
			if err := t.setupControlProtocol(t.ctx); err != nil {
				t.cleanup()
				return fmt.Errorf("setup control protocol: %w", err)
			}
		}

		t.connected = true
		return nil
	})
}

// buildArgs builds the CLI arguments based on the transport mode.
// This is a thin wrapper around BuildArgs for backwards compatibility.
func (t *Transport) buildArgs() []string {
	return BuildArgs(t.model, t.systemPrompt, t.customArgs, t.tools, t.mcpServers, t.promptArg)
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
	// Increase buffer size to handle large JSON responses (default 64KB is often insufficient)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 1MB initial, 10MB max
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

		// Debug: log raw response line to file
		if debugFile, err := os.OpenFile("/tmp/sdk-debug.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(debugFile, "response: %s\n", line)
			debugFile.Close()
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

		// Route control messages to protocol if active
		if t.protocol != nil && (msgType == MessageTypeControlRequest || msgType == MessageTypeControlResponse) {
			if err := t.protocol.HandleIncomingMessage(t.ctx, rawMsg); err != nil {
				select {
				case t.errChan <- fmt.Errorf("control protocol: %w", err):
				case <-t.ctx.Done():
					return
				}
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

// handleStderr reads from stderr and forwards to callback or error channel.
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
				// If stderr callback is set, invoke it with panic recovery
				if t.stderrCallback != nil {
					func() {
						defer func() {
							if r := recover(); r != nil {
								// Log panic but don't crash the session
								select {
								case t.errChan <- fmt.Errorf("stderr callback panic: %v", r):
								case <-t.ctx.Done():
								}
							}
						}()
						t.stderrCallback(line)
					}()
				} else {
					// Default behavior: forward to error channel
					select {
					case t.errChan <- fmt.Errorf("CLI stderr: %s", line):
					case <-t.ctx.Done():
						return
					}
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

	// Close control protocol first
	if t.protocol != nil {
		_ = t.protocol.Close()
	}
	if t.protocolAdapter != nil {
		_ = t.protocolAdapter.Close()
	}

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
	if t.protocol != nil {
		_ = t.protocol.Close()
	}
	if t.protocolAdapter != nil {
		_ = t.protocolAdapter.Close()
	}
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

// needsProtocolHandshake returns true if the transport needs to set up the control protocol.
// Control protocol is needed when:
// - Hooks are configured
// - Permission callback is configured
// - File checkpointing is enabled
// - Control protocol is explicitly enabled
func (t *Transport) needsProtocolHandshake() bool {
	// Only interactive mode supports control protocol
	if t.promptArg != nil {
		return false
	}

	return t.enableControlProtocol ||
		t.canUseTool != nil ||
		len(t.protocolHooks) > 0 ||
		t.enableCheckpointing
}

// setupControlProtocol creates and initializes the control protocol.
func (t *Transport) setupControlProtocol(ctx context.Context) error {
	if t.stdin == nil {
		return fmt.Errorf("stdin not available for control protocol")
	}

	// Create protocol adapter wrapping stdin
	t.protocolAdapter = NewProtocolAdapter(t.stdin)

	// Build protocol options
	var opts []ProtocolOption

	if t.canUseTool != nil {
		opts = append(opts, WithCanUseToolCallback(t.canUseTool))
	}

	if len(t.protocolHooks) > 0 {
		opts = append(opts, WithHooks(t.protocolHooks))
	}

	if len(t.sdkMcpServers) > 0 {
		opts = append(opts, WithSdkMcpServers(t.sdkMcpServers))
	}

	// Create protocol handler
	t.protocol = NewProtocol(t.protocolAdapter, opts...)

	// Start protocol (this starts the message routing)
	if err := t.protocol.Start(ctx); err != nil {
		return fmt.Errorf("start protocol: %w", err)
	}

	// Perform initialization handshake
	if _, err := t.protocol.Initialize(ctx); err != nil {
		return fmt.Errorf("initialize protocol: %w", err)
	}

	return nil
}

// SetModel changes the AI model during a streaming session via the control protocol.
// Pass nil to reset to the default model.
// Returns error if control protocol is not active or the request fails.
func (t *Transport) SetModel(ctx context.Context, model *string) error {
	t.mu.RLock()
	protocol := t.protocol
	t.mu.RUnlock()

	if protocol == nil {
		return fmt.Errorf("control protocol not active")
	}

	return protocol.SetModel(ctx, model)
}

// SetPermissionMode changes the permission mode during a streaming session via the control protocol.
// Valid modes: "default", "acceptEdits", "plan", "bypassPermissions", "delegate", "dontAsk"
// Returns error if control protocol is not active or the request fails.
func (t *Transport) SetPermissionMode(ctx context.Context, mode string) error {
	t.mu.RLock()
	protocol := t.protocol
	t.mu.RUnlock()

	if protocol == nil {
		return fmt.Errorf("control protocol not active")
	}

	return protocol.SetPermissionMode(ctx, mode)
}

// InterruptProtocol sends an interrupt request via the control protocol.
// This is different from Interrupt() which forcibly kills the process.
// Returns error if control protocol is not active or the request fails.
func (t *Transport) InterruptProtocol(ctx context.Context) error {
	t.mu.RLock()
	protocol := t.protocol
	t.mu.RUnlock()

	if protocol == nil {
		return fmt.Errorf("control protocol not active")
	}

	return protocol.Interrupt(ctx)
}

// RewindFiles reverts tracked files to their state at a specific user message.
// The userMessageID should be the UUID from a UserMessage received during the session.
// Requires EnableCheckpointing to be set in TransportConfig.
// Returns error if control protocol is not active or the request fails.
func (t *Transport) RewindFiles(ctx context.Context, userMessageID string) error {
	t.mu.RLock()
	protocol := t.protocol
	t.mu.RUnlock()

	if protocol == nil {
		return fmt.Errorf("control protocol not active")
	}

	return protocol.RewindFiles(ctx, userMessageID)
}

// IsProtocolActive returns whether the control protocol is active.
func (t *Transport) IsProtocolActive() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.protocol != nil && !t.protocol.IsClosed()
}

// IsProtocolInitialized returns whether the control protocol has completed initialization.
func (t *Transport) IsProtocolInitialized() bool {
	t.mu.RLock()
	protocol := t.protocol
	t.mu.RUnlock()

	if protocol == nil {
		return false
	}

	return protocol.IsInitialized()
}

// Protocol returns the underlying control protocol for advanced usage.
// Returns nil if protocol is not active.
func (t *Transport) Protocol() *Protocol {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.protocol
}