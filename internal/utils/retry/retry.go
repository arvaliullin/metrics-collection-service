package retry

import (
	"context"
	"fmt"
	"time"
)

// ErrAttemptFuncNil сообщает о попытке выполнить стратегию без переданного действия.
var ErrAttemptFuncNil = fmt.Errorf("функция повтора не задана")

// ErrStrategyNil сообщает о попытке выполнить действие с nil стратегией повторов.
var ErrStrategyNil = fmt.Errorf("стратегия повторов не задана")

// DefaultDelays задаёт интервалы между повторными попытками по умолчанию.
var DefaultDelays = []time.Duration{
	time.Second,
	3 * time.Second,
	5 * time.Second,
}

// AttemptFunc описывает действие, которое должно быть выполнено со стратегией повторов.
type AttemptFunc func(ctx context.Context) error

// ShouldRetryFunc определяет, подлежит ли ошибка повторному выполнению.
type ShouldRetryFunc func(err error) bool

// Strategy задаёт реализацию механизма повторов.
type Strategy struct {
	delays      []time.Duration
	shouldRetry ShouldRetryFunc
}

// NewStrategy создаёт стратегию повторов с указанными задержками и предикатом ошибок.
func NewStrategy(delays []time.Duration, shouldRetry ShouldRetryFunc) *Strategy {
	delayCopy := make([]time.Duration, len(delays))
	copy(delayCopy, delays)
	if len(delayCopy) == 0 {
		delayCopy = append(delayCopy, DefaultDelays...)
	}

	predicate := shouldRetry
	if predicate == nil {
		predicate = func(error) bool {
			return true
		}
	}

	for i := range delayCopy {
		if delayCopy[i] < 0 {
			delayCopy[i] = 0
		}
	}

	return &Strategy{
		delays:      delayCopy,
		shouldRetry: predicate,
	}
}

// DoWithRetry выполняет действие с ограниченным числом повторов и заданными задержками.
func (s *Strategy) DoWithRetry(ctx context.Context, attempt AttemptFunc) error {
	if s == nil {
		return ErrStrategyNil
	}

	if attempt == nil {
		return ErrAttemptFuncNil
	}

	var lastErr error

	for idx := 0; idx <= len(s.delays); idx++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := attempt(ctx)
		if err == nil {
			return nil
		}

		lastErr = err
		if !s.shouldRetry(err) || idx == len(s.delays) {
			return err
		}

		delay := s.delays[idx]
		if delay <= 0 {
			continue
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}

	return lastErr
}
