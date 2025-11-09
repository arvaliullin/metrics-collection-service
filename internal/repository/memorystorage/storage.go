package memorystorage

import (
	"context"
	"fmt"
	"sync"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// Storage предоставляет in-memory реализацию MetricStorage.
type Storage struct {
	mu      sync.RWMutex
	gauge   map[string]models.Metrics
	counter map[string]models.Metrics
}

// New создает пустое in-memory хранилище метрик.
func New() *Storage {
	return &Storage{
		gauge:   make(map[string]models.Metrics),
		counter: make(map[string]models.Metrics),
	}
}

// UpdateGauge замещает значение метрики типа gauge значением newGauge.
func (s *Storage) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metric := models.Metrics{
		ID:    id,
		MType: models.Gauge,
		Value: &newGauge,
	}
	s.gauge[id] = metric
}

// GetGauge возвращает значение метрики gauge для id.
func (s *Storage) GetGauge(ctx context.Context, id string) (float64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if metric, exists := s.gauge[id]; exists {
		return *metric.Value, nil
	}
	return 0, fmt.Errorf("для %s не задано значение Gauage", id)
}

// GetCounter возвращает значение метрики counter для id.
func (s *Storage) GetCounter(ctx context.Context, id string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if metric, exists := s.counter[id]; exists {
		if metric.Value != nil {
			return int64(*metric.Value), nil
		} else if metric.Delta != nil {
			return *metric.Delta, nil
		}
	}
	return 0, fmt.Errorf("для %s не задано значение Counter", id)
}

// AddCounter добавляет к предыдущему значению newCounter.
func (s *Storage) AddCounter(ctx context.Context, id string, newCounter int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var currentDelta int64
	if metric, exists := s.counter[id]; exists {
		if metric.Delta != nil {
			currentDelta = *metric.Delta
		} else if metric.Value != nil {
			currentDelta = int64(*metric.Value)
		}
	}

	newDelta := currentDelta + newCounter
	newMetric := models.Metrics{
		ID:    id,
		MType: models.Counter,
		Delta: &newDelta,
	}

	s.counter[id] = newMetric
}

// ResetCounter сбрасывает значение счетчика до нуля.
func (s *Storage) ResetCounter(ctx context.Context, id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	zeroDelta := int64(0)
	newMetric := models.Metrics{
		ID:    id,
		MType: models.Counter,
		Delta: &zeroDelta,
	}

	s.counter[id] = newMetric
}

// AddCounterValue добавляет значение delta к счетчику.
func (s *Storage) AddCounterValue(ctx context.Context, id string, delta int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var currentValue int64
	if metric, exists := s.counter[id]; exists {
		if metric.Value != nil {
			currentValue = int64(*metric.Value)
		}
	}

	newValue := float64(currentValue + delta)
	newMetric := models.Metrics{
		ID:    id,
		MType: models.Counter,
		Value: &newValue,
	}

	s.counter[id] = newMetric
}

// AllCounters возвращает все счетчики.
func (s *Storage) AllCounters(ctx context.Context) []models.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(s.counter))

	for _, value := range s.counter {
		newMetric := models.Metrics{
			ID:    value.ID,
			MType: value.MType,
		}
		if value.Delta != nil {
			newMetric.Delta = value.Delta
		} else if value.Value != nil {
			delta := int64(*value.Value)
			newMetric.Delta = &delta
		}
		metrics = append(metrics, newMetric)
	}

	return metrics
}

// AllGauges возвращает все gauge-метрики.
func (s *Storage) AllGauges(ctx context.Context) []models.Metrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(s.gauge))

	for _, value := range s.gauge {
		metrics = append(metrics, value)
	}

	return metrics
}
