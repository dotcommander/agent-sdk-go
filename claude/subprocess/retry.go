package subprocess

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/dotcommander/agent-sdk-go/internal/shared"
)

const (
	// maxRetries is the maximum number of retry attempts for transient failures.
	maxRetries = 3
	// baseDelay is the base delay for exponential backoff (100ms).
	baseDelay = 100 * time.Millisecond
)

// withRetry executes a function with exponential backoff retry logic.
// Retries transient failures (connection errors, timeouts, process errors).
func withRetry(ctx context.Context, operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn()
		if err == nil {
			return nil // Success
		}

		// Check if error is retryable
		if !isRetryableError(err) {
			return fmt.Errorf("%s: non-retryable error: %w", operation, err)
		}

		lastErr = fmt.Errorf("%s (attempt %d/%d): %w", operation, attempt+1, maxRetries, err)

		// Calculate delay with exponential backoff and jitter
		if attempt < maxRetries-1 {
			delay := calculateDelay(attempt)
			select {
			case <-time.After(delay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("%s failed after %d attempts: %w", operation, maxRetries, lastErr)
}

// isRetryableError determines if an error should be retried.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// Check specific error types that are retryable
	if shared.IsConnectionError(err) {
		return true
	}

	if shared.IsTimeoutError(err) {
		return true
	}

	// Check for process-related errors
	if _, ok := err.(*shared.ProcessError); ok {
		return true
	}

	// Check for transient I/O errors
	if strings.Contains(err.Error(), "resource temporarily unavailable") ||
		strings.Contains(err.Error(), "connection refused") ||
		strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "timeout") ||
		strings.Contains(err.Error(), "EOF") {
		return true
	}

	return false
}

// calculateDelay calculates the delay with exponential backoff and jitter.
func calculateDelay(attempt int) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	delay := float64(baseDelay) * float64(uint(1)<<uint(attempt))

	// Add jitter (random factor between 0.5x and 1.5x)
	jitter := 0.5 + rand.Float64()*1.0
	delay *= jitter

	// Cap at 5 seconds to avoid excessive delays
	cappedDelay := time.Duration(delay)
	if cappedDelay > 5*time.Second {
		cappedDelay = 5 * time.Second
	}

	return cappedDelay
}
