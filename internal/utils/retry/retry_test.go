package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewStrategyDefaults(t *testing.T) {
	strategy := NewStrategy(nil, nil)
	if strategy == nil {
		t.Fatal("strategy is nil")
	}

	expected := []time.Duration{
		time.Second,
		3 * time.Second,
		5 * time.Second,
	}
	if len(strategy.delays) != len(expected) {
		t.Fatalf("expected %d delays, got %d", len(expected), len(strategy.delays))
	}
	for i, v := range expected {
		if strategy.delays[i] != v {
			t.Fatalf("expected delay %s at index %d, got %s", v, i, strategy.delays[i])
		}
	}

	if !strategy.shouldRetry(errors.New("any")) {
		t.Fatal("default predicate should allow retry")
	}
}

func TestNewStrategyNormalizesDelays(t *testing.T) {
	strategy := NewStrategy([]time.Duration{-time.Second, 0}, nil)
	if strategy.delays[0] != 0 {
		t.Fatalf("expected normalized delay 0, got %s", strategy.delays[0])
	}
}

func TestDoWithRetryRetriesUntilSuccess(t *testing.T) {
	strategy := NewStrategy([]time.Duration{0, 0, 0}, func(error) bool { return true })
	attempts := 0
	err := strategy.DoWithRetry(context.Background(), func(context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("fail")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestDoWithRetryStopsWhenPredicateFalse(t *testing.T) {
	strategy := NewStrategy([]time.Duration{0, 0, 0}, func(error) bool { return false })
	attempts := 0
	testErr := errors.New("fail")
	err := strategy.DoWithRetry(context.Background(), func(context.Context) error {
		attempts++
		return testErr
	})
	if !errors.Is(err, testErr) {
		t.Fatalf("expected %v, got %v", testErr, err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt when predicate false, got %d", attempts)
	}
}

func TestDoWithRetryErrAttemptFuncNil(t *testing.T) {
	strategy := NewStrategy(nil, nil)
	if err := strategy.DoWithRetry(context.Background(), nil); !errors.Is(err, ErrAttemptFuncNil) {
		t.Fatalf("expected ErrAttemptFuncNil, got %v", err)
	}
}

func TestDoWithRetryNilStrategy(t *testing.T) {
	var strategy *Strategy
	err := strategy.DoWithRetry(context.Background(), func(context.Context) error {
		return nil
	})
	if !errors.Is(err, ErrStrategyNil) {
		t.Fatalf("expected ErrStrategyNil, got %v", err)
	}
}

func TestDoWithRetryContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	strategy := NewStrategy([]time.Duration{time.Hour}, func(error) bool { return true })
	attempts := 0
	err := strategy.DoWithRetry(ctx, func(context.Context) error {
		attempts++
		cancel()
		return errors.New("fail")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt before cancellation, got %d", attempts)
	}
}
