package main

import (
	"context"
	"time"
)

// Retry policy. Variables so tests can speed up the delays.
var (
	retryAttempts  = 3
	retryBaseDelay = 2 * time.Second
	retryMaxDelay  = 30 * time.Second
)

// retryWithBackoff calls fn, retrying with exponential backoff while it returns an
// error. Before each retry it calls onRetry with the error and the upcoming delay.
// It stops when fn succeeds, the attempts are exhausted, or the context is canceled.
// The last error is returned, or the context error if cancellation came first.
func retryWithBackoff(
	ctx context.Context,
	fn func() error,
	onRetry func(err error, delay time.Duration),
) error {
	delay := retryBaseDelay

	var err error
	for attempt := 1; attempt <= retryAttempts; attempt++ {
		if err = fn(); err == nil {
			return nil
		}

		if attempt == retryAttempts {
			break
		}

		onRetry(err, delay)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		delay = min(2*delay, retryMaxDelay)
	}

	return err
}
