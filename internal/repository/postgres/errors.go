package postgres

import (
	"errors"
	"fmt"
)

// ErrGaugeNotFound сообщает о том, что gauge метрика не найдена.
var ErrGaugeNotFound = errors.New("gauge not found")

// ErrCounterNotFound сообщает о том, что counter метрика не найдена.
var ErrCounterNotFound = errors.New("counter not found")

// GaugeNotFoundError представляет ошибку отсутствия gauge метрики с указанным ID.
type GaugeNotFoundError struct {
	ID string
}

func (e *GaugeNotFoundError) Error() string {
	return fmt.Sprintf("gauge %q not found", e.ID)
}

func (e *GaugeNotFoundError) Is(target error) bool {
	return target == ErrGaugeNotFound
}

// CounterNotFoundError представляет ошибку отсутствия counter метрики с указанным ID.
type CounterNotFoundError struct {
	ID string
}

func (e *CounterNotFoundError) Error() string {
	return fmt.Sprintf("counter %q not found", e.ID)
}

func (e *CounterNotFoundError) Is(target error) bool {
	return target == ErrCounterNotFound
}

// ErrGaugeHasNoValue сообщает о том, что gauge метрика не имеет значения.
var ErrGaugeHasNoValue = errors.New("gauge has no value")

// ErrCounterHasNoDelta сообщает о том, что counter метрика не имеет дельты.
var ErrCounterHasNoDelta = errors.New("counter has no delta")

// ErrUnsupportedMetricType сообщает о неподдерживаемом типе метрики.
var ErrUnsupportedMetricType = errors.New("unsupported metric type")

// GaugeHasNoValueError представляет ошибку отсутствия значения у gauge метрики.
type GaugeHasNoValueError struct {
	ID string
}

func (e *GaugeHasNoValueError) Error() string {
	return fmt.Sprintf("gauge %q has no value", e.ID)
}

func (e *GaugeHasNoValueError) Is(target error) bool {
	return target == ErrGaugeHasNoValue
}

// CounterHasNoDeltaError представляет ошибку отсутствия дельты у counter метрики.
type CounterHasNoDeltaError struct {
	ID string
}

func (e *CounterHasNoDeltaError) Error() string {
	return fmt.Sprintf("counter %q has no delta", e.ID)
}

func (e *CounterHasNoDeltaError) Is(target error) bool {
	return target == ErrCounterHasNoDelta
}

// UnsupportedMetricTypeError представляет ошибку неподдерживаемого типа метрики.
type UnsupportedMetricTypeError struct {
	Type string
}

func (e *UnsupportedMetricTypeError) Error() string {
	return fmt.Sprintf("unsupported metric type %q", e.Type)
}

func (e *UnsupportedMetricTypeError) Is(target error) bool {
	return target == ErrUnsupportedMetricType
}
