package memory

import (
	"context"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// BatchUpdate применяет пакет метрик.
func (r *Repository) BatchUpdate(ctx context.Context, metrics []models.Metrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				continue
			}
			value := *metric.Value
			r.gauge[metric.ID] = models.Metrics{
				ID:    metric.ID,
				MType: models.Gauge,
				Value: &value,
			}
		case models.Counter:
			var delta int64
			switch {
			case metric.Delta != nil:
				delta = *metric.Delta
			case metric.Value != nil:
				delta = int64(*metric.Value)
			default:
				continue
			}

			var current int64
			if existing, ok := r.counter[metric.ID]; ok {
				if existing.Delta != nil {
					current = *existing.Delta
				} else if existing.Value != nil {
					current = int64(*existing.Value)
				}
			}

			newDelta := current + delta
			r.counter[metric.ID] = models.Metrics{
				ID:    metric.ID,
				MType: models.Counter,
				Delta: &newDelta,
			}
		}
	}

	return nil
}
