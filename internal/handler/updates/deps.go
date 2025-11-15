package updates

import (
	"context"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

//go:generate mockgen -source=deps.go -destination=mock/storage_mock.go -package=updatesmock
type MetricStorage interface {
	BatchUpdate(ctx context.Context, metrics []models.Metrics) error
}
