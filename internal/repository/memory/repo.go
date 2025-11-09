package memory

import (
	"context"
	"fmt"
	"sync"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// Repository предоставляет in-memory реализацию MetricStorage.
type Repository struct {
	mu      sync.RWMutex
	gauge   map[string]models.Metrics
	counter map[string]models.Metrics
}

// NewRepository создает пустое in-memory хранилище метрик.
func NewRepository() *Repository {
	return &Repository{
		gauge:   make(map[string]models.Metrics),
		counter: make(map[string]models.Metrics),
	}
}

// UpdateGauge замещает значение метрики типа gauge значением newGauge.
func (r *Repository) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	metric := models.Metrics{
		ID:    id,
		MType: models.Gauge,
		Value: &newGauge,
	}
	r.gauge[id] = metric
}

// GetGauge возвращает значение метрики gauge для id.
func (r *Repository) GetGauge(ctx context.Context, id string) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if metric, exists := r.gauge[id]; exists {
		return *metric.Value, nil
	}
	return 0, fmt.Errorf("для %s не задано значение Gauage", id)
}

// GetCounter возвращает значение метрики counter для id.
func (r *Repository) GetCounter(ctx context.Context, id string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if metric, exists := r.counter[id]; exists {
		if metric.Value != nil {
			return int64(*metric.Value), nil
		} else if metric.Delta != nil {
			return *metric.Delta, nil
		}
	}
	return 0, fmt.Errorf("для %s не задано значение Counter", id)
}

// AddCounter добавляет к предыдущему значению newCounter.
func (r *Repository) AddCounter(ctx context.Context, id string, newCounter int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var currentDelta int64
	if metric, exists := r.counter[id]; exists {
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

	r.counter[id] = newMetric
}

// ResetCounter сбрасывает значение счетчика до нуля.
func (r *Repository) ResetCounter(ctx context.Context, id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	zeroDelta := int64(0)
	newMetric := models.Metrics{
		ID:    id,
		MType: models.Counter,
		Delta: &zeroDelta,
	}

	r.counter[id] = newMetric
}

// AddCounterValue добавляет значение delta к счетчику.
func (r *Repository) AddCounterValue(ctx context.Context, id string, delta int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var currentValue int64
	if metric, exists := r.counter[id]; exists {
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

	r.counter[id] = newMetric
}

// AllCounters возвращает все счетчики.
func (r *Repository) AllCounters(ctx context.Context) []models.Metrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(r.counter))

	for _, value := range r.counter {
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
func (r *Repository) AllGauges(ctx context.Context) []models.Metrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(r.gauge))

	for _, value := range r.gauge {
		metrics = append(metrics, value)
	}

	return metrics
}


