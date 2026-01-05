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

	existing, exists := r.gauge[metric.ID]
	if exists && existing.Value != nil {
		*existing.Value = *metric.Value
		return
	}

	value := *metric.Value
	r.gauge[metric.ID] = models.Metrics{
		ID:    metric.ID,
		MType: models.Gauge,
		Value: &value,
	}
}

// updateCounter обновляет метрику типа Counter, минимизируя аллокации.
func (r *Repository) updateCounter(metric *models.Metrics) {
	if metric.Delta == nil && metric.Value == nil {
		return
	}

	delta := r.extractDelta(metric)

	existing, exists := r.counter[metric.ID]
	if !exists {
		r.counter[metric.ID] = models.Metrics{
			ID:    metric.ID,
			MType: models.Counter,
			Delta: &delta,
		}
		return
	}

	if existing.Delta != nil {
		*existing.Delta += delta
		r.counter[metric.ID] = existing
	} else if existing.Value != nil {
		current := int64(*existing.Value)
		newDelta := current + delta
		r.counter[metric.ID] = models.Metrics{
			ID:    metric.ID,
			MType: models.Counter,
			Delta: &newDelta,
		}
	} else {
		r.counter[metric.ID] = models.Metrics{
			ID:    metric.ID,
			MType: models.Counter,
			Delta: &delta,
		}
	}
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
