package utils

import (
	"context"
	"time"
)

// WithTimeout creates a context with a specified timeout
func WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// WithStandardTimeout creates a context with the standard 10-second timeout
func WithStandardTimeout() (context.Context, context.CancelFunc) {
	return WithTimeout(10 * time.Second)
}

// WithDBTimeout creates a context with database-specific timeout
func WithDBTimeout() (context.Context, context.CancelFunc) {
	return WithTimeout(10 * time.Second)
}
