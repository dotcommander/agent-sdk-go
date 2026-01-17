package v2

import (
	"context"
	"fmt"

	"github.com/dotcommander/agent-sdk-go/claude/shared"
)

// withSessionCleanup is an internal helper that sets up context-based session cleanup.
// It starts a goroutine that closes the session when the context is done.
// This extracts the repeated pattern from all WithSession* functions.
func withSessionCleanup(ctx context.Context, session V2Session) {
	shared.CloseOnCancelFunc(ctx, func() {
		session.Close()
	})
}

// withSessionCancelFunc is an internal helper that creates a cancel function
// that both closes the session and cancels the context.
// This extracts the repeated pattern from WithSession*WithError functions.
func withSessionCancelFunc(session V2Session, cancel context.CancelFunc) context.CancelFunc {
	return func() {
		session.Close()
		cancel()
	}
}

// WithSession creates a new session and returns it along with a context
// that will automatically close the session when the context is cancelled.
// This provides a convenient way to ensure sessions are properly cleaned up.
//
// Example:
//
//	session, ctx, err := v2.WithSession(context.Background(),
//	    v2.WithModel("claude-sonnet-4-5-20250929"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer session.Close(ctx)
//
//	session.Send(ctx, "Hello!")
//	for msg := range session.Receive(ctx) {
//	    // ... handle messages ...
//	}
func WithSession(ctx context.Context, opts ...SessionOption) (V2Session, context.Context, error) {
	session, err := CreateSession(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	withSessionCleanup(ctx, session)
	return session, ctx, nil
}

// WithSessionWithError creates a new session and returns it along with a context
// and a cancel function. This provides more control over session cleanup.
//
// The cancel function should be called to close the session and cancel the context.
// If the context is cancelled (e.g., via timeout), the session will be automatically closed.
//
// Example:
//
//	session, ctx, cancel, err := v2.WithSessionWithError(context.Background(),
//	    v2.WithModel("claude-sonnet-4-5-20250929"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cancel()
//
//	session.Send(ctx, "Hello!")
//	for msg := range session.Receive(ctx) {
//	    // ... handle messages ...
//	}
func WithSessionWithError(parentCtx context.Context, opts ...SessionOption) (V2Session, context.Context, context.CancelFunc, error) {
	session, err := CreateSession(parentCtx, opts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create session: %w", err)
	}

	ctx, cancel := context.WithCancel(parentCtx)
	cancelFunc := withSessionCancelFunc(session, cancel)
	withSessionCleanup(ctx, session)

	return session, ctx, cancelFunc, nil
}

// WithSessionResume resumes an existing session and returns it along with a context
// that will automatically close the session when the context is cancelled.
//
// Example:
//
//	session, ctx, err := v2.WithSessionResume(context.Background(), sessionID,
//	    v2.WithModel("claude-sonnet-4-5-20250929"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer session.Close(ctx)
//
//	session.Send(ctx, "Continue our conversation")
func WithSessionResume(ctx context.Context, sessionID string, opts ...SessionOption) (V2Session, context.Context, error) {
	session, err := ResumeSession(ctx, sessionID, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("resume session: %w", err)
	}

	withSessionCleanup(ctx, session)
	return session, ctx, nil
}

// WithSessionResumeWithError resumes an existing session and returns it along
// with a context and a cancel function for more control over session cleanup.
//
// Example:
//
//	session, ctx, cancel, err := v2.WithSessionResumeWithError(context.Background(), sessionID,
//	    v2.WithModel("claude-sonnet-4-5-20250929"))
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cancel()
//
//	session.Send(ctx, "Continue our conversation")
func WithSessionResumeWithError(parentCtx context.Context, sessionID string, opts ...SessionOption) (V2Session, context.Context, context.CancelFunc, error) {
	session, err := ResumeSession(parentCtx, sessionID, opts...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("resume session: %w", err)
	}

	ctx, cancel := context.WithCancel(parentCtx)
	cancelFunc := withSessionCancelFunc(session, cancel)
	withSessionCleanup(ctx, session)

	return session, ctx, cancelFunc, nil
}
