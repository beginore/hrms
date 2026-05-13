package retry

import (
	"context"
	"log"
	"time"
)

const (
	maxAttempts = 3
	baseDelay   = 500 * time.Millisecond
)

// Do calls fn up to maxAttempts times, doubling the delay after each failure.
// It stops immediately if ctx is cancelled or fn succeeds.
func Do(ctx context.Context, label string, fn func() error) error {
	var err error
	delay := baseDelay
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err = fn(); err == nil {
			return nil
		}
		if attempt == maxAttempts {
			break
		}
		log.Printf("[retry] %s failed (attempt %d/%d): %v — retrying in %s", label, attempt, maxAttempts, err, delay)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay *= 2
	}
	return err
}
