package repository

import (
	"context"
	"fmt"
	"sync"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

type MemStorage interface {
	UpdateGauge(ctx context.Context, id string, newGauge float64)
	GetGauge(ctx context.Context, id string) (float64, error)
	GetCounter(ctx context.Context, id string) (int64, error)
	AddCounter(ctx context.Context, id string, newConter int64)
	AllCounters(ctx context.Context) []models.Metrics
	AllGauges(ctx context.Context) []models.Metrics
}

type memStorage struct {
	mu      sync.RWMutex
	gauge   map[string]models.Metrics
	counter map[string]models.Metrics
}

// NewEmptyMemStorage создает пустое
func NewEmptyMemStorage() *memStorage {
	return &memStorage{
		gauge:   make(map[string]models.Metrics),
		counter: make(map[string]models.Metrics),
	}
}

// UpdateGauge замещает значение метрики типа Gauge значением newGauge
func (ms *memStorage) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	metrics := models.Metrics{
		ID:    id,
		MType: models.Gauge,
		Value: &newGauge,
	}
	ms.gauge[id] = metrics
}

// GetGauge возвращает значение метрики gauge для id
func (ms *memStorage) GetGauge(ctx context.Context, id string) (float64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if metric, exists := ms.gauge[id]; exists {
		return *metric.Value, nil
	}
	return 0, fmt.Errorf("для %s не задано значение Gauage", id)
}

// GetCounter возвращает значение метрики counter для id
func (ms *memStorage) GetCounter(ctx context.Context, id string) (int64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if metric, exists := ms.counter[id]; exists {
		return int64(*metric.Value), nil
	}
	return 0, fmt.Errorf("для %s не задано значение Counter", id)
}

// AddCounter добавляет к предыдущему значению newCounter
func (ms *memStorage) AddCounter(ctx context.Context, id string, newConter int64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	value := float64(newConter)
	newMetric := models.Metrics{
		ID:    id,
		MType: models.Counter,
		Value: &value,
	}

	if metric, exists := ms.counter[id]; exists {
		value += *metric.Value
		newMetric = metric
		newMetric.Value = &value
		ms.counter[id] = newMetric
	}

	ms.counter[id] = newMetric
}

func (ms *memStorage) AllCounters(ctx context.Context) []models.Metrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(ms.counter))

	for _, value := range ms.counter {
		metrics = append(metrics, value)
	}

	return metrics
}

func (ms *memStorage) AllGauges(ctx context.Context) []models.Metrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(ms.gauge))

	for _, value := range ms.gauge {
		metrics = append(metrics, value)
	}

	return metrics
}
