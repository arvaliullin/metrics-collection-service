package postgres

import (
	"context"
	"fmt"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// BatchUpdate применяет пакет обновлений метрик в одной транзакции.
func (r *Repository) BatchUpdate(ctx context.Context, metrics []models.Metrics) (err error) {
	if len(metrics) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer tx.Rollback(ctx)

	const (
		upsertGauge = `
INSERT INTO metrics_gauge (id, value)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET
	value = EXCLUDED.value`
		upsertCounter = `
INSERT INTO metrics_counter (id, value)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET
	value = metrics_counter.value + EXCLUDED.value`
	)

	for _, metric := range metrics {
		switch metric.MType {
		case models.Gauge:
			if metric.Value == nil {
				return fmt.Errorf("gauge %q has no value", metric.ID)
			}
			if _, err = tx.Exec(ctx, upsertGauge, metric.ID, *metric.Value); err != nil {
				return fmt.Errorf("upsert gauge %q: %w", metric.ID, err)
			}
		case models.Counter:
			var delta int64
			switch {
			case metric.Delta != nil:
				delta = *metric.Delta
			case metric.Value != nil:
				delta = int64(*metric.Value)
			default:
				return fmt.Errorf("counter %q has no delta", metric.ID)
			}
			if _, err = tx.Exec(ctx, upsertCounter, metric.ID, delta); err != nil {
				return fmt.Errorf("upsert counter %q: %w", metric.ID, err)
			}
		default:
			return fmt.Errorf("unsupported metric type %q", metric.MType)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}
