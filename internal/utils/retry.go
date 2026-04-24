package utils

import (
	"context"
	"errors"
	"time"
)

// RetryConfig controls retry behaviour for external service calls.
type RetryConfig struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// DefaultRetryConfig is a sensible starting point for HTTP client retries.
var DefaultRetryConfig = RetryConfig{
	MaxAttempts:  3,
	InitialDelay: 200 * time.Millisecond,
	MaxDelay:     5 * time.Second,
	Multiplier:   2.0,
}

// RetryableError wraps an error and signals whether the caller should retry.
type RetryableError struct {
	Err       error
	Retryable bool
}

func (e *RetryableError) Error() string { return e.Err.Error() }
func (e *RetryableError) Unwrap() error { return e.Err }

// Do calls fn up to cfg.MaxAttempts times, backing off exponentially.
// fn should return a *RetryableError with Retryable=false to stop retrying
// immediately (e.g. for 4xx responses).
func Do(ctx context.Context, cfg RetryConfig, fn func() error) error {
	delay := cfg.InitialDelay

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		var retryable *RetryableError
		if errors.As(err, &retryable) && !retryable.Retryable {
			return retryable.Err
		}

		if attempt == cfg.MaxAttempts {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		delay = time.Duration(float64(delay) * cfg.Multiplier)
		if delay > cfg.MaxDelay {
			delay = cfg.MaxDelay
		}
	}

	return nil
}
