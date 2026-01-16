package v2

import (
	"context"
	"fmt"
	"sync"
	"time"

	"agent-sdk-go/internal/claude"
	"agent-sdk-go/internal/claude/cli"
	"agent-sdk-go/internal/claude/shared"
)

// sessionConfig holds configuration for session creation.
type sessionConfig struct {
	sessionID string // empty for new sessions, set for resumed sessions
	resumed   bool   // true if resuming an existing session
}

// newSession is the internal factory for creating sessions.
// Both CreateSession and ResumeSession delegate to this function.
func newSession(ctx context.Context, cfg sessionConfig, opts ...SessionOption) (V2Session, error) {
	// Apply options
	options := DefaultSessionOptions()
	for _, opt := range opts {
		opt(options)
	}

	// Validate options
	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Check if Claude CLI is available (using injected checker for testability)
	cliChecker := options.cliChecker
	if cliChecker == nil {
		cliChecker = shared.CLICheckerFunc(cli.IsCLIAvailable)
	}
	if !cliChecker.IsCLIAvailable() {
		return nil, fmt.Errorf("claude CLI not found. Please install it first")
	}

	// Get the client factory (DIP: depend on abstraction, not concrete NewClient)
	factory := options.clientFactory
	if factory == nil {
		factory = DefaultClientFactory()
	}

	// Create the underlying client using the factory
	client, err := factory.NewClient(
		claude.WithModel(options.Model),
		claude.WithTimeout(options.Timeout.String()),
	)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	// Determine session ID
	sessionID := cfg.sessionID
	if sessionID == "" {
		sessionID = generateSessionID()
	} else {
		// Set the session ID on the client for resumed sessions
		if clientImpl, ok := client.(*claude.ClientImpl); ok {
			clientImpl.SetSessionID(sessionID)
		}
	}

	// Create session
	session := &v2SessionImpl{
		client:      client,
		options:     options,
		sessionID:   sessionID,
		mu:          sync.RWMutex{},
		closed:      false,
		pendingSend: nil,
		resumed:     cfg.resumed,
	}

	// Connect to Claude CLI
	if err := session.client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	return session, nil
}

// CreateSession creates a new V2 session with the provided options.
// This is equivalent to unstable_v2_createSession() in the TypeScript SDK.
//
// The session maintains a connection to Claude CLI and allows for
// multi-turn conversations using the Send/Receive pattern.
//
// Example:
//
//	session, err := v2.CreateSession(ctx,
//	    v2.WithModel("claude-sonnet-4-5-20250929"),
//	    v2.WithTimeout(30*time.Second))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer session.Close()
//
//	session.Send(ctx, "Hello!")
//	for msg := range session.Receive(ctx) {
//	    if msg.Type() == v2.V2EventTypeAssistant {
//	        fmt.Println(v2.ExtractText(msg))
//	    }
//	}
func CreateSession(ctx context.Context, opts ...SessionOption) (V2Session, error) {
	return newSession(ctx, sessionConfig{}, opts...)
}

// ResumeSession resumes an existing session by ID.
// This is equivalent to unstable_v2_resumeSession() in the TypeScript SDK.
//
// Example:
//
//	session, err := v2.ResumeSession(ctx, sessionID,
//	    v2.WithModel("claude-sonnet-4-5-20250929"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer session.Close()
//
//	session.Send(ctx, "What number did I ask you to remember?")
func ResumeSession(ctx context.Context, sessionID string, opts ...SessionOption) (V2Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}
	return newSession(ctx, sessionConfig{sessionID: sessionID, resumed: true}, opts...)
}

// v2SessionImpl implements the V2Session interface.
type v2SessionImpl struct {
	client      claude.Client
	options     *V2SessionOptions
	sessionID   string
	mu          sync.RWMutex
	closed      bool
	pendingSend *pendingSendData
	resumed     bool
}

// pendingSendData holds data from a Send operation that will be consumed by Receive.
type pendingSendData struct {
	message  string
	sentAt   time.Time
	consumed bool
}

// Send sends a message to Claude in this session.
// The message is queued and will be sent when Receive is called.
func (s *v2SessionImpl) Send(ctx context.Context, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return fmt.Errorf("session is closed")
	}

	// Store the pending message
	s.pendingSend = &pendingSendData{
		message:  message,
		sentAt:   time.Now(),
		consumed: false,
	}

	return nil
}

// Receive returns a channel of V2Message responses.
// The channel closes when the response is complete.
// If Send was called, the message is sent first.
func (s *v2SessionImpl) Receive(ctx context.Context) <-chan V2Message {
	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()
		ch := make(chan V2Message)
		close(ch)
		return ch
	}

	// Check if there's a pending send and take a copy while holding the lock
	pendingData := s.pendingSend
	pendingDataCopy := pendingData

	s.mu.Unlock()

	// If there's a pending send, send it first
	if pendingDataCopy != nil && !pendingDataCopy.consumed {
		// Send the message
		msgChan, errChan := s.client.QueryStream(ctx, pendingDataCopy.message)

		// Mark as consumed (acquire lock again)
		s.mu.Lock()
		if s.pendingSend == pendingDataCopy {
			s.pendingSend.consumed = true
		}
		s.mu.Unlock()

		// Wrap the messages as V2 messages
		return s.wrapMessageChannel(ctx, msgChan, errChan)
	}

	// Otherwise, just receive messages from the client
	msgChan, errChan := s.client.ReceiveMessages(ctx)
	return s.wrapMessageChannel(ctx, msgChan, errChan)
}

// ReceiveIterator returns an iterator for receiving messages.
// This provides an alternative to the channel-based Receive().
func (s *v2SessionImpl) ReceiveIterator(ctx context.Context) V2MessageIterator {
	return &v2MessageIteratorImpl{
		session: s,
		ctx:     ctx,
		ch:      s.Receive(ctx),
		closed:  false,
	}
}

// Close closes the session and releases resources.
func (s *v2SessionImpl) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil
	}

	s.closed = true

	// Disconnect the client
	return s.client.Disconnect()
}

// SessionID returns the unique session identifier.
func (s *v2SessionImpl) SessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.sessionID
}

// wrapMessageChannel wraps the client's message channel and converts messages to V2 format.
func (s *v2SessionImpl) wrapMessageChannel(ctx context.Context, msgChan <-chan claude.Message, errChan <-chan error) <-chan V2Message {
	out := make(chan V2Message, 100)

	go func() {
		defer close(out)

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgChan:
				if !ok {
					return
				}
				// Convert to V2 message
				v2Msg := convertToV2Message(msg, s.sessionID)
				if v2Msg != nil {
					select {
					case out <- v2Msg:
					case <-ctx.Done():
						return
					}
				}
			case err, ok := <-errChan:
				if !ok {
					return
				}
				// Convert error to V2 error message
				v2Err := &V2Error{
					TypeField:  V2EventTypeError,
					ErrorField: err.Error(),
					SessionID:  s.sessionID,
				}
				select {
				case out <- v2Err:
				case <-ctx.Done():
					return
				}
				return
			}
		}
	}()

	return out
}

// convertToV2Message converts a client Message to a V2Message.
func convertToV2Message(msg claude.Message, sessionID string) V2Message {
	switch m := msg.(type) {
	case *shared.AssistantMessage:
		return &V2AssistantMessage{
			TypeField: V2EventTypeAssistant,
			Message: AssistantMessageContent{
				Content: m.Content,
				Model:   m.Model,
			},
			SessionID: sessionID,
		}
	case *shared.ResultMessage:
		result := ""
		if m.Result != nil {
			result = *m.Result
		}
		return &V2ResultMessage{
			TypeField: V2EventTypeResult,
			Result:    result,
			SessionID: sessionID,
		}
	case *shared.StreamEvent:
		return &V2StreamDelta{
			TypeField: V2EventTypeStreamDelta,
			Delta:     m.Event,
			SessionID: sessionID,
		}
	default:
		return nil
	}
}

// generateSessionID generates a new unique session ID.
func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}

// IsResumed returns true if this session was resumed.
func (s *v2SessionImpl) IsResumed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.resumed
}

// GetClient returns the underlying client (for advanced usage).
func (s *v2SessionImpl) GetClient() claude.Client {
	return s.client
}

// String returns a string representation of the session.
func (s *v2SessionImpl) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := "active"
	if s.closed {
		status = "closed"
	}

	return fmt.Sprintf("V2Session{%s: id=%q, model=%s}",
		status, s.sessionID, s.options.Model)
}

// Interrupt stops the current query execution.
func (s *v2SessionImpl) Interrupt(ctx context.Context) error {
	return fmt.Errorf("Interrupt not implemented")
}

// SetPermissionMode changes the permission mode.
func (s *v2SessionImpl) SetPermissionMode(ctx context.Context, mode shared.PermissionMode) error {
	return fmt.Errorf("SetPermissionMode not implemented")
}

// SetModel changes the model used.
func (s *v2SessionImpl) SetModel(ctx context.Context, model string) error {
	return fmt.Errorf("SetModel not implemented")
}

// SetMaxThinkingTokens adjusts the thinking token limit.
// Pass nil to clear the limit.
func (s *v2SessionImpl) SetMaxThinkingTokens(ctx context.Context, tokens *int) error {
	return fmt.Errorf("SetMaxThinkingTokens not implemented")
}

// SupportedCommands returns available slash commands.
func (s *v2SessionImpl) SupportedCommands(ctx context.Context) ([]shared.SlashCommand, error) {
	return nil, fmt.Errorf("SupportedCommands not implemented")
}

// SupportedModels returns available models.
func (s *v2SessionImpl) SupportedModels(ctx context.Context) ([]shared.ModelInfo, error) {
	return nil, fmt.Errorf("SupportedModels not implemented")
}

// McpServerStatus returns MCP server statuses.
func (s *v2SessionImpl) McpServerStatus(ctx context.Context) ([]shared.McpServerStatus, error) {
	return nil, fmt.Errorf("McpServerStatus not implemented")
}

// AccountInfo returns account information.
func (s *v2SessionImpl) AccountInfo(ctx context.Context) (*shared.AccountInfo, error) {
	return nil, fmt.Errorf("AccountInfo not implemented")
}

// RewindFiles rewinds files to a specific message state.
func (s *v2SessionImpl) RewindFiles(ctx context.Context, userMessageID string, opts *RewindFilesOptions) (*shared.RewindFilesResult, error) {
	return nil, fmt.Errorf("RewindFiles not implemented")
}

// SetMcpServers dynamically sets MCP servers.
func (s *v2SessionImpl) SetMcpServers(ctx context.Context, servers map[string]shared.McpServerConfig) (*shared.McpSetServersResult, error) {
	return nil, fmt.Errorf("SetMcpServers not implemented")
}

// v2MessageIteratorImpl implements the V2MessageIterator interface.
type v2MessageIteratorImpl struct {
	session *v2SessionImpl
	ctx     context.Context
	ch      <-chan V2Message
	closed  bool
	mu      sync.Mutex
}

// Next advances to the next message.
// Returns ErrNoMoreMessages when iteration is complete.
func (it *v2MessageIteratorImpl) Next(ctx context.Context) (V2Message, error) {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.closed {
		return nil, ErrNoMoreMessages
	}

	select {
	case msg, ok := <-it.ch:
		if !ok {
			it.closed = true
			return nil, ErrNoMoreMessages
		}
		return msg, nil
	case <-ctx.Done():
		it.closed = true
		return nil, ctx.Err()
	case <-it.ctx.Done():
		it.closed = true
		return nil, it.ctx.Err()
	}
}

// Close closes the iterator and releases resources.
func (it *v2MessageIteratorImpl) Close() error {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.closed {
		return nil
	}

	it.closed = true
	return nil
}

// IsClosed returns true if the iterator is closed.
func (it *v2MessageIteratorImpl) IsClosed() bool {
	it.mu.Lock()
	defer it.mu.Unlock()
	return it.closed
}
