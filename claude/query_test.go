package claude

import (
	"context"
	"errors"
	"testing"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrNoMoreMessages(t *testing.T) {
	t.Parallel()

	// Verify ErrNoMoreMessages is a distinct error
	assert.NotNil(t, ErrNoMoreMessages)
	assert.Equal(t, "no more messages", ErrNoMoreMessages.Error())
}

func TestQueryIteratorNext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		messages   []shared.Message
		errs       []error
		wantMsgs   int
		wantErr    error
		wantFinal  error
	}{
		{
			name: "single message",
			messages: []shared.Message{
				&shared.AssistantMessage{
					MessageType: "assistant",
					Content: []shared.ContentBlock{
						&shared.TextBlock{Text: "Hello"},
					},
				},
			},
			wantMsgs:  1,
			wantFinal: ErrNoMoreMessages,
		},
		{
			name: "multiple messages",
			messages: []shared.Message{
				&shared.AssistantMessage{MessageType: "assistant"},
				&shared.AssistantMessage{MessageType: "assistant"},
				&shared.AssistantMessage{MessageType: "assistant"},
			},
			wantMsgs:  3,
			wantFinal: ErrNoMoreMessages,
		},
		{
			name:       "empty stream",
			messages:   []shared.Message{},
			wantMsgs:   0,
			wantFinal:  ErrNoMoreMessages,
		},
		{
			name:     "error in stream",
			messages: []shared.Message{},
			errs:     []error{errors.New("test error")},
			wantMsgs: 0,
			wantErr:  errors.New("test error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create buffered channels
			msgChan := make(chan shared.Message, len(tt.messages)+1)
			errChan := make(chan error, len(tt.errs)+1)

			// Send messages
			for _, msg := range tt.messages {
				msgChan <- msg
			}
			close(msgChan)

			// Send errors
			for _, err := range tt.errs {
				errChan <- err
			}
			close(errChan)

			// Create iterator
			iter := &queryIterator{
				msgChan: msgChan,
				errChan: errChan,
			}
			defer iter.Close()

			// Read messages
			ctx := context.Background()
			gotMsgs := 0

			for {
				msg, err := iter.Next(ctx)

				if tt.wantErr != nil && err != nil && !errors.Is(err, ErrNoMoreMessages) {
					assert.Equal(t, tt.wantErr.Error(), err.Error())
					return
				}

				if errors.Is(err, ErrNoMoreMessages) {
					assert.Equal(t, tt.wantFinal, err)
					break
				}

				require.NoError(t, err)
				assert.NotNil(t, msg)
				gotMsgs++
			}

			assert.Equal(t, tt.wantMsgs, gotMsgs)
		})
	}
}

func TestQueryIteratorClose(t *testing.T) {
	t.Parallel()

	msgChan := make(chan shared.Message)
	errChan := make(chan error)

	iter := &queryIterator{
		msgChan: msgChan,
		errChan: errChan,
	}

	// First close should succeed
	err := iter.Close()
	assert.NoError(t, err)

	// Second close should be idempotent
	err = iter.Close()
	assert.NoError(t, err)

	// After close, Next should return ErrNoMoreMessages
	msg, err := iter.Next(context.Background())
	assert.Nil(t, msg)
	assert.ErrorIs(t, err, ErrNoMoreMessages)
}

func TestQueryIteratorContextCancellation(t *testing.T) {
	t.Parallel()

	msgChan := make(chan shared.Message)
	errChan := make(chan error)

	iter := &queryIterator{
		msgChan: msgChan,
		errChan: errChan,
	}
	defer iter.Close()

	// Cancel context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Next should return context error
	_, err := iter.Next(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestQueryWithTransport(t *testing.T) {
	t.Parallel()

	// Create channels
	msgChan := make(chan shared.Message, 1)
	errChan := make(chan error, 1)

	// Send a message
	msgChan <- &shared.AssistantMessage{
		MessageType: "assistant",
		Content: []shared.ContentBlock{
			&shared.TextBlock{Text: "Test response"},
		},
	}
	close(msgChan)
	close(errChan)

	// Create iterator with nil transport (acceptable for testing)
	iter := QueryWithTransport(nil, msgChan, errChan)
	defer iter.Close()

	// Read the message
	msg, err := iter.Next(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, msg)

	// Should be done
	_, err = iter.Next(context.Background())
	assert.ErrorIs(t, err, ErrNoMoreMessages)
}

func TestMessageIteratorInterface(t *testing.T) {
	t.Parallel()

	// Verify queryIterator implements MessageIterator
	var _ MessageIterator = (*queryIterator)(nil)
}
