package shared

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

func TestGetValue(t *testing.T) {
	const testKey ContextKey = "test_key"

	tests := []struct {
		name     string
		setup    func() context.Context
		wantVal  string
		wantOK   bool
	}{
		{
			name: "value exists",
			setup: func() context.Context {
				return WithValue(context.Background(), testKey, "hello")
			},
			wantVal: "hello",
			wantOK:  true,
		},
		{
			name: "value missing",
			setup: func() context.Context {
				return context.Background()
			},
			wantVal: "",
			wantOK:  false,
		},
		{
			name: "wrong type",
			setup: func() context.Context {
				return context.WithValue(context.Background(), testKey, 123)
			},
			wantVal: "",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.setup()
			val, ok := GetValue[string](ctx, testKey)
			if val != tt.wantVal {
				t.Errorf("GetValue() value = %v, want %v", val, tt.wantVal)
			}
			if ok != tt.wantOK {
				t.Errorf("GetValue() ok = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestMustGetValue(t *testing.T) {
	const testKey ContextKey = "test_key"

	t.Run("value exists", func(t *testing.T) {
		ctx := WithValue(context.Background(), testKey, "hello")
		val := MustGetValue[string](ctx, testKey)
		if val != "hello" {
			t.Errorf("MustGetValue() = %v, want %v", val, "hello")
		}
	})

	t.Run("panics on missing", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustGetValue() did not panic on missing value")
			}
		}()
		ctx := context.Background()
		MustGetValue[string](ctx, testKey)
	})
}

type mockCloser struct {
	closed bool
}

func (m *mockCloser) Close() error {
	m.closed = true
	return nil
}

func TestCloseOnCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	closer := &mockCloser{}

	wg := CloseOnCancel(ctx, closer)

	if closer.closed {
		t.Error("closer should not be closed yet")
	}

	cancel()
	wg.Wait()

	if !closer.closed {
		t.Error("closer should be closed after context cancellation")
	}
}

func TestCloseOnCancelFunc(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var cleanupCalled bool

	wg := CloseOnCancelFunc(ctx, func() {
		cleanupCalled = true
	})

	if cleanupCalled {
		t.Error("cleanup should not be called yet")
	}

	cancel()
	wg.Wait()

	if !cleanupCalled {
		t.Error("cleanup should be called after context cancellation")
	}
}

func TestSelectWithContext(t *testing.T) {
	t.Run("receives value", func(t *testing.T) {
		ctx := context.Background()
		ch := make(chan int, 1)
		ch <- 42

		result := SelectWithContext(ctx, ch)
		if result.Err != nil {
			t.Errorf("unexpected error: %v", result.Err)
		}
		if result.Value != 42 {
			t.Errorf("SelectWithContext() = %v, want %v", result.Value, 42)
		}
	})

	t.Run("channel closed", func(t *testing.T) {
		ctx := context.Background()
		ch := make(chan int)
		close(ch)

		result := SelectWithContext(ctx, ch)
		if result.Err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", result.Err)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ch := make(chan int)

		result := SelectWithContext(ctx, ch)
		if result.Err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", result.Err)
		}
	})
}

func TestSelectWithContextAndError(t *testing.T) {
	t.Run("receives value", func(t *testing.T) {
		ctx := context.Background()
		valCh := make(chan string, 1)
		errCh := make(chan error)
		valCh <- "hello"

		result := SelectWithContextAndError(ctx, valCh, errCh)
		if result.Err != nil {
			t.Errorf("unexpected error: %v", result.Err)
		}
		if result.Value != "hello" {
			t.Errorf("SelectWithContextAndError() = %v, want %v", result.Value, "hello")
		}
	})

	t.Run("receives error", func(t *testing.T) {
		ctx := context.Background()
		valCh := make(chan string)
		errCh := make(chan error, 1)
		testErr := errors.New("test error")
		errCh <- testErr

		result := SelectWithContextAndError(ctx, valCh, errCh)
		if result.Err != testErr {
			t.Errorf("expected test error, got %v", result.Err)
		}
	})

	t.Run("context cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		valCh := make(chan string)
		errCh := make(chan error)

		result := SelectWithContextAndError(ctx, valCh, errCh)
		if result.Err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", result.Err)
		}
	})
}

func TestIsDone(t *testing.T) {
	t.Run("not cancelled", func(t *testing.T) {
		ctx := context.Background()
		if IsDone(ctx) {
			t.Error("IsDone() should return false for active context")
		}
	})

	t.Run("cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if !IsDone(ctx) {
			t.Error("IsDone() should return true for cancelled context")
		}
	})
}

func TestDoneErr(t *testing.T) {
	t.Run("not cancelled", func(t *testing.T) {
		ctx := context.Background()
		if err := DoneErr(ctx); err != nil {
			t.Errorf("DoneErr() = %v, want nil", err)
		}
	})

	t.Run("cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := DoneErr(ctx); err != context.Canceled {
			t.Errorf("DoneErr() = %v, want context.Canceled", err)
		}
	})

	t.Run("deadline exceeded", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		time.Sleep(10 * time.Millisecond)
		err := DoneErr(ctx)
		if err != context.DeadlineExceeded {
			t.Errorf("DoneErr() = %v, want context.DeadlineExceeded", err)
		}
	})
}

// Verify mockCloser satisfies io.Closer
var _ io.Closer = (*mockCloser)(nil)
