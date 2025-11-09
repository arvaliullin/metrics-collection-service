package agent

import (
	"context"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

//go:generate mockgen -source=deps.go -destination=mock/storage_mock.go -package=agentmock
type MetricStorage interface {
	UpdateGauge(ctx context.Context, id string, newGauge float64)
	AddCounter(ctx context.Context, id string, newCounter int64)
	AllCounters(ctx context.Context) []models.Metrics
	AllGauges(ctx context.Context) []models.Metrics
	ResetCounter(ctx context.Context, id string)
}
