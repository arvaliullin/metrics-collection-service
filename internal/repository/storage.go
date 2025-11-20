package repository

import (
	"context"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// MetricStorage определяет общий контракт хранилища метрик.
//
//go:generate mockgen -source=storage.go -destination=mock/storage_mock.go -package=repomock
type MetricStorage interface {
	UpdateGauge(ctx context.Context, id string, newGauge float64)
	GetGauge(ctx context.Context, id string) (float64, error)
	GetCounter(ctx context.Context, id string) (int64, error)
	AddCounter(ctx context.Context, id string, newCounter int64)
	AddCounterValue(ctx context.Context, id string, delta int64)
	BatchUpdate(ctx context.Context, metrics []models.Metrics) error
	AllCounters(ctx context.Context) []models.Metrics
	AllGauges(ctx context.Context) []models.Metrics
	ResetCounter(ctx context.Context, id string)
	Ping(ctx context.Context) error
}
