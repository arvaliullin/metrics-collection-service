package update

import "context"

//go:generate mockgen -source=deps.go -destination=mock/storage_mock.go -package=updatemock
type MetricStorage interface {
	UpdateGauge(ctx context.Context, id string, newGauge float64)
	GetGauge(ctx context.Context, id string) (float64, error)
	GetCounter(ctx context.Context, id string) (int64, error)
	AddCounter(ctx context.Context, id string, newCounter int64)
	AddCounterValue(ctx context.Context, id string, delta int64)
}
