package server

import (
	"context"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

//go:generate mockgen -source=deps.go -destination=mock/storage_mock.go -package=server_mock
type MetricStorage interface {
	UpdateGauge(ctx context.Context, id string, newGauge float64)
	GetGauge(ctx context.Context, id string) (float64, error)
	GetCounter(ctx context.Context, id string) (int64, error)
	AddCounter(ctx context.Context, id string, newCounter int64)
	AddCounterValue(ctx context.Context, id string, delta int64)
	AllCounters(ctx context.Context) []models.Metrics
	AllGauges(ctx context.Context) []models.Metrics
	Ping(ctx context.Context) error
}
