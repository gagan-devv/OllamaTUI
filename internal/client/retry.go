package client

import (
	"context"
	"fmt"
	"time"
)

// RetryPolicy defines the retry behavior for failed operations
type RetryPolicy struct {
	MaxAttempts  int           // Maximum number of retry attempts (default 3)
	InitialDelay time.Duration // Initial delay between retries (default 2s)
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Multiplier for exponential backoff
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// RetryableFuncWithContext is a function with context that can be retried
type RetryableFuncWithContext func(ctx context.Context) error

// RetryNotifier is called before each retry attempt
type RetryNotifier func(attempt int, err error, delay time.Duration)

// NewRetryPolicy creates a new RetryPolicy with default values
func NewRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:  3,
		InitialDelay: 2 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// Execute executes the given function with retry logic
// Returns the error from the last attempt if all retries fail
func (p *RetryPolicy) Execute(fn RetryableFunc) error {
	return p.ExecuteWithNotifier(fn, nil)
}

// ExecuteWithNotifier executes the function with retry logic and calls the notifier before each retry
func (p *RetryPolicy) ExecuteWithNotifier(fn RetryableFunc, notifier RetryNotifier) error {
	var lastErr error
	delay := p.InitialDelay

	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if this is a retryable error
		if !isRetryable(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// If this was the last attempt, don't wait
		if attempt >= p.MaxAttempts {
			break
		}

		// Notify before retry
		if notifier != nil {
			notifier(attempt, err, delay)
		}

		// Wait before retrying
		time.Sleep(delay)

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * p.Multiplier)
		if delay > p.MaxDelay {
			delay = p.MaxDelay
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", p.MaxAttempts, lastErr)
}

// ExecuteWithContext executes the function with context and retry logic
func (p *RetryPolicy) ExecuteWithContext(ctx context.Context, fn RetryableFuncWithContext) error {
	return p.ExecuteWithContextAndNotifier(ctx, fn, nil)
}

// ExecuteWithContextAndNotifier executes the function with context, retry logic, and notifier
func (p *RetryPolicy) ExecuteWithContextAndNotifier(ctx context.Context, fn RetryableFuncWithContext, notifier RetryNotifier) error {
	var lastErr error
	delay := p.InitialDelay

	for attempt := 1; attempt <= p.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if this is a retryable error
		if !isRetryable(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// If this was the last attempt, don't wait
		if attempt >= p.MaxAttempts {
			break
		}

		// Notify before retry
		if notifier != nil {
			notifier(attempt, err, delay)
		}

		// Wait before retrying, respecting context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * p.Multiplier)
		if delay > p.MaxDelay {
			delay = p.MaxDelay
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", p.MaxAttempts, lastErr)
}

// isRetryable determines if an error should trigger a retry
// Network errors, timeouts, and temporary errors are retryable
// Validation errors and permanent errors are not retryable
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for common retryable error patterns
	errStr := err.Error()
	
	// Network-related errors are retryable
	retryablePatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"network is unreachable",
		"no such host",
		"EOF",
		"broken pipe",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	// Non-retryable error patterns
	nonRetryablePatterns := []string{
		"validation",
		"invalid",
		"not found",
		"unauthorized",
		"forbidden",
		"bad request",
	}

	for _, pattern := range nonRetryablePatterns {
		if contains(errStr, pattern) {
			return false
		}
	}

	// Default to retryable for unknown errors
	return true
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if matchesAt(s, substr, i) {
			return true
		}
	}
	return false
}

func matchesAt(s, substr string, pos int) bool {
	for i := 0; i < len(substr); i++ {
		if toLower(s[pos+i]) != toLower(substr[i]) {
			return false
		}
	}
	return true
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}
