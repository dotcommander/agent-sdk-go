package shared

import (
	"context"
	"io"
	"sync"
)

// ContextKey is a type for context value keys to avoid collisions.
type ContextKey string

// GetValue retrieves a typed value from context with type assertion.
// Returns the zero value and false if the key doesn't exist or type doesn't match.
func GetValue[T any](ctx context.Context, key ContextKey) (T, bool) {
	v := ctx.Value(key)
	if v == nil {
		var zero T
		return zero, false
	}
	val, ok := v.(T)
	return val, ok
}

// MustGetValue retrieves a typed value from context, panicking if not found.
// Use this only when the value is guaranteed to exist.
func MustGetValue[T any](ctx context.Context, key ContextKey) T {
	val, ok := GetValue[T](ctx, key)
	if !ok {
		panic("context value not found for key: " + string(key))
	}
	return val
}

// WithValue adds a typed value to context.
func WithValue[T any](ctx context.Context, key ContextKey, val T) context.Context {
	return context.WithValue(ctx, key, val)
}

// CloseOnCancel starts a goroutine that closes a resource when context is cancelled.
// Returns a WaitGroup that completes when the cleanup goroutine finishes.
// This is a common pattern for ensuring resources are cleaned up on context cancellation.
//
// Example:
//
//	wg := shared.CloseOnCancel(ctx, conn)
//	// ... use conn ...
//	wg.Wait() // ensure cleanup completed
func CloseOnCancel(ctx context.Context, closer io.Closer) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		closer.Close()
	}()
	return &wg
}

// CloseOnCancelFunc starts a goroutine that calls a cleanup function when context is cancelled.
// This is useful when you need custom cleanup logic beyond io.Closer.
//
// Example:
//
//	wg := shared.CloseOnCancelFunc(ctx, func() {
//	    session.Close()
//	    conn.Disconnect()
//	})
func CloseOnCancelFunc(ctx context.Context, cleanup func()) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		cleanup()
	}()
	return &wg
}

// ContextResult wraps an operation result with context-aware error handling.
// Useful for select statements that handle both results and context cancellation.
type ContextResult[T any] struct {
	Value T
	Err   error
}

// SelectWithContext helps implement the common pattern of selecting between
// a channel result and context cancellation. Returns the result or context error.
//
// Example:
//
//	result := shared.SelectWithContext(ctx, resultChan)
//	if result.Err != nil {
//	    return result.Err
//	}
//	return result.Value
func SelectWithContext[T any](ctx context.Context, ch <-chan T) ContextResult[T] {
	select {
	case val, ok := <-ch:
		if !ok {
			return ContextResult[T]{Err: context.Canceled}
		}
		return ContextResult[T]{Value: val}
	case <-ctx.Done():
		return ContextResult[T]{Err: ctx.Err()}
	}
}

// SelectWithContextAndError handles the common pattern of selecting between
// a value channel, an error channel, and context cancellation.
// This consolidates the repetitive three-way select pattern found throughout the codebase.
//
// Example:
//
//	result := shared.SelectWithContextAndError(ctx, msgChan, errChan)
//	if result.Err != nil {
//	    return result.Err
//	}
//	msg := result.Value
func SelectWithContextAndError[T any](ctx context.Context, valueChan <-chan T, errChan <-chan error) ContextResult[T] {
	select {
	case val, ok := <-valueChan:
		if !ok {
			return ContextResult[T]{Err: context.Canceled}
		}
		return ContextResult[T]{Value: val}
	case err, ok := <-errChan:
		if !ok {
			return ContextResult[T]{Err: context.Canceled}
		}
		return ContextResult[T]{Err: err}
	case <-ctx.Done():
		return ContextResult[T]{Err: ctx.Err()}
	}
}

// IsDone checks if the context has been cancelled without blocking.
// Useful for early exit in loops.
func IsDone(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

// DoneErr returns the context error if cancelled, nil otherwise.
// Non-blocking version of ctx.Err() for quick checks.
func DoneErr(ctx context.Context) error {
	if IsDone(ctx) {
		return ctx.Err()
	}
	return nil
}
