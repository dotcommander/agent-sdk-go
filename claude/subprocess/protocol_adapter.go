// Package subprocess provides subprocess communication with the Claude CLI.
// This file provides an adapter to wrap stdin for the control protocol.
package subprocess

import (
	"context"
	"io"
	"sync"
)

// ProtocolAdapter wraps stdin to implement ControlTransport interface.
// It provides a bridge between the Transport's stdin and the Protocol.
type ProtocolAdapter struct {
	stdin  io.Writer
	mu     sync.Mutex
	closed bool
}

// NewProtocolAdapter creates a new protocol adapter wrapping stdin.
func NewProtocolAdapter(stdin io.Writer) *ProtocolAdapter {
	return &ProtocolAdapter{
		stdin: stdin,
	}
}

// Write sends data to the CLI stdin.
func (a *ProtocolAdapter) Write(ctx context.Context, data []byte) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.closed {
		return io.ErrClosedPipe
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	_, err := a.stdin.Write(data)
	return err
}

// Read returns a closed channel since we route messages directly from handleStdout.
// The Protocol's readLoop will exit immediately when this channel is closed.
func (a *ProtocolAdapter) Read(ctx context.Context) <-chan []byte {
	// Return nil to indicate that messages are routed directly
	// rather than through this channel
	return nil
}

// Close marks the adapter as closed.
// It does NOT close the underlying stdin - that's managed by Transport.
func (a *ProtocolAdapter) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.closed = true
	return nil
}

// IsClosed returns whether the adapter is closed.
func (a *ProtocolAdapter) IsClosed() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.closed
}
