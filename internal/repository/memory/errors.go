package memory

import (
	"errors"
	"fmt"
)

// ErrGaugeNotFound сообщает о том, что значение gauge метрики не найдено.
var ErrGaugeNotFound = errors.New("gauge metric value not found")

// ErrCounterNotFound сообщает о том, что значение counter метрики не найдено.
var ErrCounterNotFound = errors.New("counter metric value not found")

// ErrPingNotAvailable сообщает о том, что проверка соединения с БД недоступна для in-memory хранилища.
var ErrPingNotAvailable = errors.New("проверка соединения с БД недоступна: используется in-memory хранилище")

// GaugeNotFoundError представляет ошибку отсутствия gauge метрики с указанным ID.
type GaugeNotFoundError struct {
	ID string
}

func (e *GaugeNotFoundError) Error() string {
	return fmt.Sprintf("для %s не задано значение Gauge", e.ID)
}

func (e *GaugeNotFoundError) Is(target error) bool {
	return target == ErrGaugeNotFound
}

// CounterNotFoundError представляет ошибку отсутствия counter метрики с указанным ID.
type CounterNotFoundError struct {
	ID string
}

func (e *CounterNotFoundError) Error() string {
	return fmt.Sprintf("для %s не задано значение Counter", e.ID)
}

func (e *CounterNotFoundError) Is(target error) bool {
	return target == ErrCounterNotFound
}
