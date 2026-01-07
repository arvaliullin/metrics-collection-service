package ports

import (
	"context"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

//go:generate mockgen -source=services.go -destination=mock/services_mock.go -package=portsmock

// MetricsCollector отвечает за сбор метрик системы.
type MetricsCollector interface {
	CollectRuntimeMetrics(ctx context.Context) error
	CollectGopsutilMetrics(ctx context.Context) error
}

// MetricsReporter отвечает за отправку метрик на сервер.
type MetricsReporter interface {
	ReportBatch(ctx context.Context, metrics []models.Metrics) error
	BuildBatch(ctx context.Context) ([]models.Metrics, error)
}

// MetricsSender отвечает за отправку метрик на сервер.
type MetricsSender interface {
	Send(ctx context.Context, metrics []models.Metrics) error
}

// MetricsService координирует работу сбора и отправки метрик.
type MetricsService interface {
	StartCollection(ctx context.Context) error
	StartReporting(ctx context.Context) error
}
