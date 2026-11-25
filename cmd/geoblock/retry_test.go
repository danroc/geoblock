package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

// errTest is the error returned by flakyFn.
var errTest = errors.New("operation failed")

// flakyFn fails the first `failures` calls, then succeeds.
type flakyFn struct {
	failures int
	calls    int
}

func (f *flakyFn) run() error {
	f.calls++
	if f.calls <= f.failures {
		return errTest
	}
	return nil
}

// blockingFn returns a function that blocks until the context is canceled.
func blockingFn(ctx context.Context) func() error {
	return func() error {
		<-ctx.Done()
		return ctx.Err()
	}
}

// retryCall records one onRetry notification.
type retryCall struct {
	err   error
	delay time.Duration
}

func TestRetryWithBackoff(t *testing.T) {
	// Speed up the backoff delays for testing.
	originalBase, originalMax := retryBaseDelay, retryMaxDelay
	retryBaseDelay = time.Millisecond
	retryMaxDelay = 2 * time.Millisecond
	t.Cleanup(func() {
		retryBaseDelay, retryMaxDelay = originalBase, originalMax
	})

	// collectRetry returns an onRetry callback that records notifications.
	collectRetry := func(calls *[]retryCall) func(error, time.Duration) {
		return func(err error, delay time.Duration) {
			*calls = append(*calls, retryCall{err: err, delay: delay})
		}
	}

	t.Run("succeeds on first attempt", func(t *testing.T) {
		fn := &flakyFn{failures: 0}
		var calls []retryCall
		err := retryWithBackoff(
			context.Background(),
			fn.run,
			collectRetry(&calls),
		)
		if err != nil {
			t.Errorf("retryWithBackoff() unexpected error: %v", err)
		}
		if fn.calls != 1 {
			t.Errorf("calls = %d, want 1", fn.calls)
		}
		if len(calls) != 0 {
			t.Errorf("onRetry calls = %d, want 0", len(calls))
		}
	})

	t.Run("succeeds after transient failures", func(t *testing.T) {
		fn := &flakyFn{failures: retryAttempts - 1}
		var calls []retryCall
		err := retryWithBackoff(
			context.Background(),
			fn.run,
			collectRetry(&calls),
		)
		if err != nil {
			t.Errorf("retryWithBackoff() unexpected error: %v", err)
		}
		if fn.calls != retryAttempts {
			t.Errorf("calls = %d, want %d", fn.calls, retryAttempts)
		}
		if len(calls) != retryAttempts-1 {
			t.Errorf("onRetry calls = %d, want %d", len(calls), retryAttempts-1)
		}
	})

	t.Run("fails after all attempts and reports the last error", func(t *testing.T) {
		fn := &flakyFn{failures: retryAttempts}
		var calls []retryCall
		err := retryWithBackoff(
			context.Background(),
			fn.run,
			collectRetry(&calls),
		)
		if !errors.Is(err, errTest) {
			t.Errorf("retryWithBackoff() error = %v, want %v", err, errTest)
		}
		if fn.calls != retryAttempts {
			t.Errorf("calls = %d, want %d", fn.calls, retryAttempts)
		}

		wantDelays := []time.Duration{
			retryBaseDelay,
			min(2*retryBaseDelay, retryMaxDelay),
		}
		if len(calls) != len(wantDelays) {
			t.Fatalf("onRetry calls = %d, want %d", len(calls), len(wantDelays))
		}
		for i, call := range calls {
			if call.delay != wantDelays[i] {
				t.Errorf("onRetry delay %d = %v, want %v", i, call.delay, wantDelays[i])
			}
			if call.err == nil {
				t.Errorf("onRetry error %d = nil, want non-nil", i)
			}
		}
	})

	t.Run("stops when context is canceled", func(t *testing.T) {
		testContextCancellation(t, func(ctx context.Context) {
			if err := retryWithBackoff(
				ctx,
				blockingFn(ctx),
				warnRetry,
			); !errors.Is(err, ctx.Err()) {
				t.Errorf("retryWithBackoff() error = %v, want %v", err, ctx.Err())
			}
		})
	})
}
