package get

import "context"

//go:generate mockgen -source=deps.go -destination=mock/storage_mock.go -package=getmock
type MetricStorage interface {
	GetGauge(ctx context.Context, id string) (float64, error)
	GetCounter(ctx context.Context, id string) (int64, error)
}
