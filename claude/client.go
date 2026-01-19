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
		CanUseTool:   c.options.CanUseTool,
	}

	// Convert hooks from shared.HookConfig to transport's ProtocolHookMatcher
	if len(c.options.Hooks) > 0 {
		transportConfig.ProtocolHooks = convertHooksToProtocolFormat(c.options.Hooks)
		transportConfig.EnableControlProtocol = true
	}

	// Enable control protocol if permission callback is set
	if c.options.CanUseTool != nil {
		transportConfig.EnableControlProtocol = true
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

// convertHooksToProtocolFormat converts shared.HookConfig to subprocess.ProtocolHookMatcher.
// This bridges the client-level hook API to the transport-level protocol format.
func convertHooksToProtocolFormat(hooks map[shared.HookEvent][]shared.HookConfig) map[shared.HookEvent][]subprocess.ProtocolHookMatcher {
	result := make(map[shared.HookEvent][]subprocess.ProtocolHookMatcher)

	for event, configs := range hooks {
		var matchers []subprocess.ProtocolHookMatcher

		for _, config := range configs {
			// Capture handler in local variable for closure
			handler := config.Handler
			configMatcher := config.Matcher
			configTimeout := config.Timeout

			// Convert each HookConfig to a ProtocolHookMatcher
			matcher := subprocess.ProtocolHookMatcher{
				Matcher: configMatcher,
				Hooks: []subprocess.ProtocolHookCallback{
					func(ctx context.Context, input any, toolUseID *string) (*shared.SyncHookOutput, error) {
						return handler(ctx, input)
					},
				},
			}

			// Convert timeout if set
			if configTimeout > 0 {
				timeoutSec := configTimeout.Seconds()
				matcher.Timeout = &timeoutSec
			}

			matchers = append(matchers, matcher)
		}

		if len(matchers) > 0 {
			result[event] = matchers
		}
	}

	return result
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

// ReceiveResponseIterator returns an iterator for receiving messages.
// This follows the same pattern as severity1's implementation for easy migration.
//
// Example:
//
//	iter := client.ReceiveResponseIterator(ctx)
//	defer iter.Close()
//	for {
//	    msg, err := iter.Next(ctx)
//	    if errors.Is(err, claude.ErrNoMoreMessages) {
//	        break
//	    }
//	    if err != nil {
//	        return err
//	    }
//	    // Process message
//	}
func (c *ClientImpl) ReceiveResponseIterator(ctx context.Context) MessageIterator {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return &clientIterator{closed: true}
	}

	msgChan, errChan := transport.ReceiveMessages(ctx)
	return &clientIterator{
		msgChan: msgChan,
		errChan: errChan,
	}
}

// clientIterator implements MessageIterator for client message reception.
type clientIterator struct {
	msgChan <-chan shared.Message
	errChan <-chan error
	closed  bool
}

// Next returns the next message or an error.
func (ci *clientIterator) Next(ctx context.Context) (Message, error) {
	if ci.closed {
		return nil, ErrNoMoreMessages
	}

	select {
	case msg, ok := <-ci.msgChan:
		if !ok {
			ci.closed = true
			return nil, ErrNoMoreMessages
		}
		return msg, nil
	case err, ok := <-ci.errChan:
		if !ok {
			ci.closed = true
			return nil, ErrNoMoreMessages
		}
		ci.closed = true
		return nil, err
	case <-ctx.Done():
		ci.closed = true
		return nil, ctx.Err()
	}
}

// Close releases resources associated with the iterator.
func (ci *clientIterator) Close() error {
	ci.closed = true
	return nil
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
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return nil, fmt.Errorf("not connected")
	}

	protocol := transport.Protocol()
	if protocol == nil {
		return nil, fmt.Errorf("control protocol not active: connect with hooks or permissions enabled")
	}

	rawStatuses, err := protocol.GetMcpServerStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("get MCP server status: %w", err)
	}

	var statuses []shared.McpServerStatus
	for _, raw := range rawStatuses {
		status := shared.McpServerStatus{
			Name: getStringField(raw, "name"),
		}
		if s, ok := raw["status"].(string); ok {
			status.Status = s
		}
		if errMsg, ok := raw["error"].(string); ok {
			status.Error = errMsg
		}
		// Parse serverInfo if present
		if serverInfoRaw, ok := raw["serverInfo"].(map[string]any); ok {
			status.ServerInfo = &shared.McpServerInfo{
				Name:    getStringField(serverInfoRaw, "name"),
				Version: getStringField(serverInfoRaw, "version"),
			}
		}
		statuses = append(statuses, status)
	}

	return statuses, nil
}

// SupportedCommands returns the list of available slash commands.
// Requires an active control protocol connection.
func (c *ClientImpl) SupportedCommands(ctx context.Context) ([]shared.SlashCommand, error) {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return nil, fmt.Errorf("not connected")
	}

	protocol := transport.Protocol()
	if protocol == nil {
		return nil, fmt.Errorf("control protocol not active: connect with hooks or permissions enabled")
	}

	rawCommands, err := protocol.GetCommands(ctx)
	if err != nil {
		return nil, fmt.Errorf("get commands: %w", err)
	}

	var commands []shared.SlashCommand
	for _, raw := range rawCommands {
		cmd := shared.SlashCommand{
			Name:         getStringField(raw, "name"),
			Description:  getStringField(raw, "description"),
			ArgumentHint: getStringField(raw, "argumentHint"),
		}
		commands = append(commands, cmd)
	}

	return commands, nil
}

// SupportedModels returns the list of available models.
// Requires an active control protocol connection.
func (c *ClientImpl) SupportedModels(ctx context.Context) ([]shared.ModelInfo, error) {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return nil, fmt.Errorf("not connected")
	}

	protocol := transport.Protocol()
	if protocol == nil {
		return nil, fmt.Errorf("control protocol not active: connect with hooks or permissions enabled")
	}

	rawModels, err := protocol.GetModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("get models: %w", err)
	}

	var models []shared.ModelInfo
	for _, raw := range rawModels {
		model := shared.ModelInfo{
			Value:       getStringField(raw, "value"),
			DisplayName: getStringField(raw, "displayName"),
			Description: getStringField(raw, "description"),
		}
		models = append(models, model)
	}

	return models, nil
}

// AccountInfo returns information about the current user's account.
// Requires an active control protocol connection.
func (c *ClientImpl) AccountInfo(ctx context.Context) (*shared.AccountInfo, error) {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	if transport == nil {
		return nil, fmt.Errorf("not connected")
	}

	protocol := transport.Protocol()
	if protocol == nil {
		return nil, fmt.Errorf("control protocol not active: connect with hooks or permissions enabled")
	}

	rawInfo, err := protocol.GetAccountInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf("get account info: %w", err)
	}

	info := &shared.AccountInfo{
		Email:            getStringField(rawInfo, "email"),
		Organization:     getStringField(rawInfo, "organization"),
		SubscriptionType: getStringField(rawInfo, "subscriptionType"),
		TokenSource:      getStringField(rawInfo, "tokenSource"),
		ApiKeySource:     getStringField(rawInfo, "apiKeySource"),
	}

	return info, nil
}

// getStringField safely extracts a string field from a map.
func getStringField(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// SetMcpServers dynamically sets MCP servers.
func (c *ClientImpl) SetMcpServers(ctx context.Context, servers map[string]shared.McpServerConfig) (*shared.McpSetServersResult, error) {
	c.mu.RLock()
	transport := c.transport
	c.mu.RUnlock()

	// Store the new MCP servers in options
	c.mu.Lock()
	c.options.McpServers = servers
	c.mu.Unlock()

	// If control protocol is active, use it to set servers dynamically
	if transport != nil {
		if protocol := transport.Protocol(); protocol != nil {
			// Convert servers to map[string]any for the protocol
			serversMap := make(map[string]any)
			for name, config := range servers {
				serversMap[name] = config
			}

			rawResult, err := protocol.SetMcpServers(ctx, serversMap)
			if err != nil {
				return nil, fmt.Errorf("set MCP servers via protocol: %w", err)
			}

			// Parse the result
			result := &shared.McpSetServersResult{
				Added:   []string{},
				Removed: []string{},
				Errors:  make(map[string]string),
			}

			if added, ok := rawResult["added"].([]any); ok {
				for _, a := range added {
					if s, ok := a.(string); ok {
						result.Added = append(result.Added, s)
					}
				}
			}
			if removed, ok := rawResult["removed"].([]any); ok {
				for _, r := range removed {
					if s, ok := r.(string); ok {
						result.Removed = append(result.Removed, s)
					}
				}
			}
			if errors, ok := rawResult["errors"].(map[string]any); ok {
				for k, v := range errors {
					if s, ok := v.(string); ok {
						result.Errors[k] = s
					}
				}
			}

			return result, nil
		}
	}

	// Fallback: return simple result indicating servers were stored
	// Note: Without control protocol, changes take effect on next connection
	result := &shared.McpSetServersResult{
		Added:   []string{},
		Removed: []string{},
		Errors:  make(map[string]string),
	}

	for name := range servers {
		result.Added = append(result.Added, name)
	}

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
