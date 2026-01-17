// Package claude provides the client implementation for communicating with the Claude CLI.
package claude

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"agent-sdk-go/claude/shared"
	"agent-sdk-go/claude/subprocess"
)

// ClientImpl implements the Client interface using subprocess transport.
type ClientImpl struct {
	transport *subprocess.Transport
	options   *ClientOptions
	sessionID string
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

// Interrupt forcibly interrupts the current operation.
func (c *ClientImpl) Interrupt() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.transport == nil {
		return nil
	}

	return c.transport.Close()
}

// SetModel sets the Claude model to use.
func (c *ClientImpl) SetModel(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.options.Model = model
}

// SetPermissionMode sets the permission mode for the session.
func (c *ClientImpl) SetPermissionMode(mode string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.options.PermissionMode = mode
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

// RewindFiles adds files to the context for the next query.
func (c *ClientImpl) RewindFiles(ctx context.Context, files []string) error {
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
