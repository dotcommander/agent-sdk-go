// Package claude provides the Query function for one-shot queries to Claude CLI.
package claude

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
	"github.com/dotcommander/agent-sdk-go/claude/subprocess"
)

// ErrNoMoreMessages is returned by MessageIterator.Next when there are no more messages.
// Re-exported from shared for convenience.
var ErrNoMoreMessages = shared.ErrNoMoreMessages

// MessageIterator provides sequential access to messages from a Claude session.
// It follows Go's io.Reader pattern with Next returning (message, error).
// Re-exported from shared for convenience.
//
// Example:
//
//	iter, err := claude.Query(ctx, "What is 2+2?")
//	if err != nil {
//	    return err
//	}
//	defer iter.Close()
//
//	for {
//	    msg, err := iter.Next(ctx)
//	    if errors.Is(err, claude.ErrNoMoreMessages) {
//	        break
//	    }
//	    if err != nil {
//	        return err
//	    }
//	    fmt.Printf("%T: %+v\n", msg, msg)
//	}
type MessageIterator = shared.MessageIterator

// queryIterator implements MessageIterator with lazy initialization.
type queryIterator struct {
	mu        sync.Mutex
	transport *subprocess.Transport
	msgChan   <-chan shared.Message
	errChan   <-chan error
	closed    bool
	lastErr   error
}

// Next returns the next message from the stream.
// Returns ErrNoMoreMessages when the stream is exhausted.
func (q *queryIterator) Next(ctx context.Context) (Message, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil, ErrNoMoreMessages
	}

	select {
	case msg, ok := <-q.msgChan:
		if !ok {
			q.closed = true
			return nil, ErrNoMoreMessages
		}
		return msg, nil

	case err, ok := <-q.errChan:
		if !ok {
			q.closed = true
			return nil, ErrNoMoreMessages
		}
		q.lastErr = err
		return nil, err

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close releases resources associated with the iterator.
func (q *queryIterator) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil
	}
	q.closed = true

	if q.transport != nil {
		return q.transport.Close()
	}
	return nil
}

// Query sends a one-shot query to Claude CLI and returns a MessageIterator.
// This is the primary entry point for simple queries that don't need session management.
//
// The iterator streams messages as they arrive from the CLI. Callers must call Close()
// on the iterator when done to release resources.
//
// Example:
//
//	iter, err := claude.Query(ctx, "Explain Go interfaces")
//	if err != nil {
//	    return err
//	}
//	defer iter.Close()
//
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
func Query(ctx context.Context, prompt string, opts ...ClientOption) (MessageIterator, error) {
	options := DefaultClientOptions()
	for _, opt := range opts {
		opt(options)
	}

	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Create transport config from client options
	transportConfig := &subprocess.TransportConfig{
		CLIPath:      options.CLIPath,
		CLICommand:   options.CLICommand,
		Model:        options.Model,
		Timeout:      parseTimeout(options.Timeout),
		SystemPrompt: "",
		CustomArgs:   options.CustomArgs,
		Env:          options.Env,
	}

	// Create one-shot transport
	transport, err := subprocess.NewTransportWithPrompt(transportConfig, prompt)
	if err != nil {
		return nil, fmt.Errorf("create transport: %w", err)
	}

	// Connect to start the subprocess
	if err := transport.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	// Get message channels
	msgChan, errChan := transport.ReceiveMessages(ctx)

	return &queryIterator{
		transport: transport,
		msgChan:   msgChan,
		errChan:   errChan,
	}, nil
}

// QueryWithTransport creates a query iterator with a custom transport.
// This is primarily for testing purposes to inject mock transports.
//
// Example (testing):
//
//	mockTransport := &MockTransport{...}
//	iter := claude.QueryWithTransport(mockTransport, msgChan, errChan)
//	defer iter.Close()
func QueryWithTransport(transport *subprocess.Transport, msgChan <-chan shared.Message, errChan <-chan error) MessageIterator {
	return &queryIterator{
		transport: transport,
		msgChan:   msgChan,
		errChan:   errChan,
	}
}

// QueryText sends a one-shot query and returns the concatenated text response.
// This is a convenience wrapper around Query that collects all text content.
//
// Example:
//
//	text, err := claude.QueryText(ctx, "What is 2+2?")
//	if err != nil {
//	    return err
//	}
//	fmt.Println(text)
func QueryText(ctx context.Context, prompt string, opts ...ClientOption) (string, error) {
	iter, err := Query(ctx, prompt, opts...)
	if err != nil {
		return "", err
	}
	defer iter.Close()

	var result string
	for {
		msg, err := iter.Next(ctx)
		if errors.Is(err, ErrNoMoreMessages) {
			break
		}
		if err != nil {
			return result, err
		}

		// Extract text content from assistant messages
		if text := shared.GetContentText(msg); text != "" {
			result += text
		}
	}

	return result, nil
}
