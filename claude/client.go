// Package claude provides the client implementation for communicating with the Claude CLI.
package claude

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
	"github.com/dotcommander/agent-sdk-go/claude/subprocess"
)

// ClientImpl implements the Client interface using subprocess transport.
type ClientImpl struct {
	transport *subprocess.Transport
	options   *ClientOptions
	sessionID string
	validator *shared.StreamValidator
	mu        sync.RWMutex
}

// NewClient creates a new Client with the given options.
func NewClient(opts ...ClientOption) (Client, error) {
	options := DefaultClientOptions()

	// Apply options
	for _, opt := range opts {
		opt(options)
	}

	// Validate options
	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	return &ClientImpl{
		options: options,
	}, nil
}

// Connect establishes a connection to Claude CLI.
func (c *ClientImpl) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport != nil {
		return fmt.Errorf("already connected")
	}

	// Create transport config
	transportConfig := &subprocess.TransportConfig{
		CLIPath:      c.options.CLIPath,
		CLICommand:   c.options.CLICommand,
		Model:        c.options.Model,
		Timeout:      parseTimeout(c.options.Timeout),
		SystemPrompt: "", // Can be added from options if needed
		CustomArgs:   c.options.CustomArgs,
		Env:          c.options.Env,
		McpServers:   c.options.McpServers,
	}

	var err error
	c.transport, err = subprocess.NewTransport(transportConfig)
	if err != nil {
		return fmt.Errorf("create transport: %w", err)
	}

	if err := c.transport.Connect(ctx); err != nil {
		c.transport = nil
		return fmt.Errorf("connect transport: %w", err)
	}

	// Initialize validator for stream tracking
	c.validator = shared.NewStreamValidator()

	return nil
}

// Disconnect closes the connection to Claude CLI.
func (c *ClientImpl) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport == nil {
		return nil
	}

	err := c.transport.Close()
	c.transport = nil
	return err
}

// Query sends a one-shot query to Claude CLI and returns the response.
func (c *ClientImpl) Query(ctx context.Context, prompt string) (string, error) {
	msgChan, errChan := c.QueryStream(ctx, prompt)

	var result strings.Builder
	for {
		select {
		case msg, ok := <-msgChan:
			if !ok {
				return result.String(), nil
			}
			if text := shared.GetContentText(msg); text != "" {
				result.WriteString(text)
			}
		case err, ok := <-errChan:
			if !ok {
				return result.String(), nil
			}
			return result.String(), err
		case <-ctx.Done():
			return result.String(), ctx.Err()
		}
	}
}

// QueryWithSession sends a query with session context.
func (c *ClientImpl) QueryWithSession(ctx context.Context, sessionID string, prompt string) (string, error) {
	// Set session ID before querying
	c.SetSessionID(sessionID)
	return c.Query(ctx, prompt)
}

// QueryStream sends a one-shot query and streams the response.
func (c *ClientImpl) QueryStream(ctx context.Context, prompt string) (<-chan Message, <-chan error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a new transport for one-shot query
	transportConfig := &subprocess.TransportConfig{
		CLIPath:      c.options.CLIPath,
		CLICommand:   c.options.CLICommand,
		Model:        c.options.Model,
		Timeout:      parseTimeout(c.options.Timeout),
		SystemPrompt: "",
		CustomArgs:   c.options.CustomArgs,
		Env:          c.options.Env,
		PromptArg:    &prompt,
	}

	transport, err := subprocess.NewTransportWithPrompt(transportConfig, prompt)
	if err != nil {
		msgChan := make(chan Message)
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("create one-shot transport: %w", err)
		close(msgChan)
		close(errChan)
		return msgChan, errChan
	}

	if err := transport.Connect(ctx); err != nil {
		msgChan := make(chan Message)
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("connect one-shot transport: %w", err)
		close(msgChan)
		close(errChan)
		return msgChan, errChan
	}

	msgChan, errChan := transport.ReceiveMessages(ctx)

	// Start cleanup goroutine
	go func() {
		<-ctx.Done()
		transport.Close()
	}()

	return msgChan, errChan
}

// ReceiveMessages receives messages from Claude CLI.
func (c *ClientImpl) ReceiveMessages(ctx context.Context) (<-chan Message, <-chan error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.transport == nil {
		msgChan := make(chan Message)
		errChan := make(chan error, 1)
		errChan <- fmt.Errorf("not connected")
		close(msgChan)
		close(errChan)
		return msgChan, errChan
	}

	return c.transport.ReceiveMessages(ctx)
}

// ReceiveResponse receives a single response from Claude CLI.
func (c *ClientImpl) ReceiveResponse(ctx context.Context) (Message, error) {
	msgChan, errChan := c.ReceiveMessages(ctx)

	select {
	case msg, ok := <-msgChan:
		if !ok {
			return nil, fmt.Errorf("no response")
		}
		return msg, nil
	case err, ok := <-errChan:
		if !ok {
			return nil, fmt.Errorf("no response")
		}
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Interrupt forcibly interrupts the current operation (process-level).
func (c *ClientImpl) Interrupt() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport == nil {
		return nil
	}

	return c.transport.Close()
}

// InterruptGraceful sends interrupt via control protocol (if active).
func (c *ClientImpl) InterruptGraceful(ctx context.Context) error {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return nil
	}

	return transport.InterruptProtocol(ctx)
}

// SetModel changes the AI model during a streaming session.
// Pass nil to reset to the default model.
// Only works when control protocol is active (connected streaming session).
func (c *ClientImpl) SetModel(ctx context.Context, model *string) error {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return fmt.Errorf("not connected")
	}

	return transport.SetModel(ctx, model)
}

// SetPermissionMode changes the permission mode during a streaming session.
// Valid modes: "default", "acceptEdits", "plan", "bypassPermissions", "delegate", "dontAsk"
// Only works when control protocol is active.
func (c *ClientImpl) SetPermissionMode(ctx context.Context, mode string) error {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return fmt.Errorf("not connected")
	}

	return transport.SetPermissionMode(ctx, mode)
}

// RewindFiles reverts tracked files to their state at a specific user message.
// The messageUUID should be the UUID from a UserMessage received during the session.
// Requires EnableFileCheckpointing option.
func (c *ClientImpl) RewindFiles(ctx context.Context, messageUUID string) error {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return fmt.Errorf("not connected")
	}

	return transport.RewindFiles(ctx, messageUUID)
}

// GetStreamIssues returns validation issues found in the message stream.
func (c *ClientImpl) GetStreamIssues() []shared.StreamIssue {
	c.mu.RLock()
	validator := c.validator
	c.mu.RUnlock()

	if validator == nil {
		return nil
	}

	return validator.GetIssues()
}

// GetStreamStats returns statistics about the message stream.
func (c *ClientImpl) GetStreamStats() shared.StreamStats {
	c.mu.RLock()
	validator := c.validator
	c.mu.RUnlock()

	if validator == nil {
		return shared.StreamStats{}
	}

	return validator.GetStats()
}

// GetServerInfo returns diagnostic information about the client connection.
func (c *ClientImpl) GetServerInfo(ctx context.Context) (map[string]any, error) {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return nil, fmt.Errorf("not connected")
	}

	return map[string]any{
		"connected":            true,
		"transport_type":       "subprocess",
		"protocol_active":      transport.IsProtocolActive(),
		"protocol_initialized": transport.IsProtocolInitialized(),
	}, nil
}

// IsProtocolActive returns whether the control protocol is active.
func (c *ClientImpl) IsProtocolActive() bool {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return false
	}

	return transport.IsProtocolActive()
}

// SetSessionID sets the session ID for this client.
// This is used by V2 session resume functionality.
func (c *ClientImpl) SetSessionID(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sessionID = sessionID
}

// GetSessionID returns the current session ID.
func (c *ClientImpl) GetSessionID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionID
}

// AddContextFiles adds files to the context for the next query.
func (c *ClientImpl) AddContextFiles(ctx context.Context, files []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.options.ContextFiles = append(c.options.ContextFiles, files...)
	return nil
}

// GetOptions returns a copy of the client options.
func (c *ClientImpl) GetOptions() *ClientOptions {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Create a copy
	copy := *c.options
	return &copy
}

// McpServerStatus returns MCP server statuses.
func (c *ClientImpl) McpServerStatus(ctx context.Context) ([]shared.McpServerStatus, error) {
	// TODO: Implement actual MCP server status retrieval
	// This would require communication with the CLI to query MCP server status
	return nil, fmt.Errorf("McpServerStatus not implemented")
}

// SetMcpServers dynamically sets MCP servers.
func (c *ClientImpl) SetMcpServers(ctx context.Context, servers map[string]shared.McpServerConfig) (*shared.McpSetServersResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store the new MCP servers in options
	c.options.McpServers = servers

	// TODO: Implement actual MCP server configuration update
	// This would require updating the transport configuration and potentially reconnecting

	// For now, return a simple result indicating servers were updated
	result := &shared.McpSetServersResult{
		Added:   []string{},
		Removed: []string{},
		Errors:  make(map[string]string),
	}

	// Extract server names for the "added" field
	for name := range servers {
		result.Added = append(result.Added, name)
	}

	// Note: This is a minimal implementation that just stores the config
	// The actual MCP server activation would require CLI communication
	return result, nil
}

// parseTimeout parses a timeout string into a time.Duration.
func parseTimeout(s string) time.Duration {
	if s == "" {
		return 30 * time.Second
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

// WithClient creates a client, passes it to the provided function, and ensures
// proper cleanup. This is the recommended pattern for using the client.
//
// The client is automatically connected before calling fn and disconnected after.
// Any error from fn or from connection/disconnection is returned.
//
// Example usage:
//
//	err := claude.WithClient(ctx, func(c claude.Client) error {
//	    response, err := c.Query(ctx, "Hello!")
//	    if err != nil {
//	        return err
//	    }
//	    fmt.Println(response)
//	    return nil
//	}, claude.WithModel("claude-sonnet-4-5-20250929"))
func WithClient(ctx context.Context, fn func(Client) error, opts ...ClientOption) error {
	client, err := NewClient(opts...)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer client.Disconnect()

	return fn(client)
}

// WithClientTransport creates a client with a custom transport for testing.
// This allows injecting mock transports for unit testing without subprocess overhead.
//
// The transport must already be connected; this function does not call Connect().
// Cleanup is the responsibility of the caller.
//
// Example usage (for testing):
//
//	mockTransport := &MockTransport{}
//	err := claude.WithClientTransport(ctx, mockTransport, func(c claude.Client) error {
//	    // Test client behavior
//	    return nil
//	})
func WithClientTransport(ctx context.Context, transport *subprocess.Transport, fn func(Client) error, opts ...ClientOption) error {
	options := DefaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	client := &ClientImpl{
		transport: transport,
		options:   options,
	}

	return fn(client)
}
