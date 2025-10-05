package repository

import (
	"fmt"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

type MemStorage interface {
	UpdateGauge(id string, newGauge float64)
	GetGauge(id string) (float64, error)
	GetCounter(id string) (int64, error)
	AddCounter(id string, newConter int64)
}

type memStorage struct {
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
func (ms *memStorage) UpdateGauge(id string, newGauge float64) {
	metrics := models.Metrics{
		ID:    id,
		MType: models.Gauge,
		Value: &newGauge,
	}
	ms.gauge[id] = metrics
}

// GetGauge возвращает значение метрики gauge для id
func (ms *memStorage) GetGauge(id string) (float64, error) {
	if metric, exists := ms.gauge[id]; exists {
		return *metric.Value, nil
	}
	return 0, fmt.Errorf("Для %s не задано значение Gauage", id)
}

// GetCounter возвращает значение метрики counter для id
func (ms *memStorage) GetCounter(id string) (int64, error) {
	if metric, exists := ms.counter[id]; exists {
		return int64(*metric.Value), nil
	}
	return 0, fmt.Errorf("Для %s не задано значение Counter", id)
}

// AddCounter добавляет к предыдущему значению newCounter
func (ms *memStorage) AddCounter(id string, newConter int64) {

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
