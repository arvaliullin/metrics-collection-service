package html

import (
	"context"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

//go:generate mockgen -source=deps.go -destination=mock/storage_mock.go -package=htmlmock
type MetricStorage interface {
	AllGauges(ctx context.Context) []models.Metrics
	AllCounters(ctx context.Context) []models.Metrics
}
