package memory

import (
	"context"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// BatchUpdate применяет пакет метрик.
func (r *Repository) BatchUpdate(ctx context.Context, metrics []models.Metrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range metrics {
		metric := &metrics[i]
		switch metric.MType {
		case models.Gauge:
			r.updateGauge(metric)
		case models.Counter:
			r.updateCounter(metric)
		}
	}

	return nil
}

// updateGauge обновляет метрику типа Gauge, минимизируя аллокации.
func (r *Repository) updateGauge(metric *models.Metrics) {
	if metric.Value == nil {
		return
	}

	r.gauge[metric.ID] = *metric.Value
}

// updateCounter обновляет метрику типа Counter, минимизируя аллокации.
func (r *Repository) updateCounter(metric *models.Metrics) {
	if metric.Delta == nil && metric.Value == nil {
		return
	}

	delta := r.extractDelta(metric)
	r.counter[metric.ID] += delta
}

// extractDelta извлекает значение delta из метрики.
func (r *Repository) extractDelta(metric *models.Metrics) int64 {
	if metric.Delta != nil {
		return *metric.Delta
	}
	if metric.Value != nil {
		return int64(*metric.Value)
	}
	return 0
}
