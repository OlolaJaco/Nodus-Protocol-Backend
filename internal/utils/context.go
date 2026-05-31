package utils

import (
	"context"
	"time"
)

// ContextWithTimeout returns a context with the given timeout and its cancel
// function. Callers must defer the cancel to avoid leaking resources.
func ContextWithTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(parent, d)
}

// ShortContext returns a context suitable for quick read operations (2s).
func ShortContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 2*time.Second)
}

// StandardContext returns a context suitable for normal API calls (10s).
func StandardContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

// LongContext returns a context suitable for batch operations (60s).
func LongContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 60*time.Second)
}
