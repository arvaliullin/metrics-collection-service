package postgres

import (
	"context"
	"fmt"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/jackc/pgx/v5"
)

// UpdateGauge устанавливает значение метрики типа gauge по идентификатору.
func (r *Repository) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	const query = `
INSERT INTO metrics_gauge (id, value)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET
	value = EXCLUDED.value`
	_, _ = r.pool.Exec(ctx, query, id, newGauge)
}

// GetGauge возвращает значение метрики gauge по идентификатору.
func (r *Repository) GetGauge(ctx context.Context, id string) (float64, error) {
	const query = `SELECT value FROM metrics_gauge WHERE id = $1`
	var v float64
	if err := r.pool.QueryRow(ctx, query, id).Scan(&v); err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("gauge %q not found", id)
		}
		return 0, err
	}
	return v, nil
}

// GetCounter возвращает значение метрики counter по идентификатору.
func (r *Repository) GetCounter(ctx context.Context, id string) (int64, error) {
	const query = `SELECT value FROM metrics_counter WHERE id = $1`
	var v int64
	if err := r.pool.QueryRow(ctx, query, id).Scan(&v); err != nil {
		if err == pgx.ErrNoRows {
			return 0, fmt.Errorf("counter %q not found", id)
		}
		return 0, err
	}
	return v, nil
}

// AddCounter увеличивает значение счётчика на указанную величину.
func (r *Repository) AddCounter(ctx context.Context, id string, newCounter int64) {
	const query = `
INSERT INTO metrics_counter (id, value)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET
	value = metrics_counter.value + EXCLUDED.value`
	_, _ = r.pool.Exec(ctx, query, id, newCounter)
}

// AddCounterValue увеличивает значение счётчика на delta.
func (r *Repository) AddCounterValue(ctx context.Context, id string, delta int64) {
	r.AddCounter(ctx, id, delta)
}

// AllCounters возвращает все метрики типа counter с заполненным полем Delta.
func (r *Repository) AllCounters(ctx context.Context) []models.Metrics {
	const query = `SELECT id, value FROM metrics_counter`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	metrics := make([]models.Metrics, 0, 64)
	for rows.Next() {
		var (
			id string
			v  int64
		)
		if err := rows.Scan(&id, &v); err != nil {
			continue
		}
		val := v
		metrics = append(metrics, models.Metrics{
			ID:    id,
			MType: models.Counter,
			Delta: &val,
		})
	}
	return metrics
}

// AllGauges возвращает все метрики типа gauge с заполненным полем Value.
func (r *Repository) AllGauges(ctx context.Context) []models.Metrics {
	const query = `SELECT id, value FROM metrics_gauge`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil
	}
	defer rows.Close()

	metrics := make([]models.Metrics, 0, 64)
	for rows.Next() {
		var (
			id string
			v  float64
		)
		if err := rows.Scan(&id, &v); err != nil {
			continue
		}
		val := v
		metrics = append(metrics, models.Metrics{
			ID:    id,
			MType: models.Gauge,
			Value: &val,
		})
	}
	return metrics
}

// ResetCounter сбрасывает значение счётчика по идентификатору в 0.
func (r *Repository) ResetCounter(ctx context.Context, id string) {
	const query = `
INSERT INTO metrics_counter (id, value)
VALUES ($1, 0)
ON CONFLICT (id) DO UPDATE SET
	value = 0`
	_, _ = r.pool.Exec(ctx, query, id)
}
