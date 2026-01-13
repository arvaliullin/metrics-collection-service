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

// AuditNotifier определяет интерфейс для уведомления о событиях аудита.
type AuditNotifier interface {
	// Subscribe добавляет наблюдателя.
	Subscribe(observer AuditObserver)
	// Unsubscribe удаляет наблюдателя.
	Unsubscribe(observer AuditObserver)
	// NotifyAll уведомляет всех наблюдателей о событии аудита.
	NotifyAll(event models.AuditEvent)
}

// AuditObserver определяет интерфейс для наблюдателей аудита.
type AuditObserver interface {
	// Notify уведомляет наблюдателя о событии аудита.
	Notify(event models.AuditEvent) error
}

// ServerMetricsService определяет интерфейс для работы с метриками на сервере.
type ServerMetricsService interface {
	// UpdateGauge обновляет значение gauge метрики.
	UpdateGauge(ctx context.Context, id string, value float64) error
	// UpdateCounter обновляет значение counter метрики.
	UpdateCounter(ctx context.Context, id string, delta int64) error
	// GetGauge получает значение gauge метрики.
	GetGauge(ctx context.Context, id string) (float64, error)
	// GetCounter получает значение counter метрики.
	GetCounter(ctx context.Context, id string) (int64, error)
	// GetMetric получает метрику по её описанию.
	GetMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error)
	// UpdateMetric обновляет метрику по её описанию.
	UpdateMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error)
	// BatchUpdate выполняет пакетное обновление метрик.
	BatchUpdate(ctx context.Context, metrics []models.Metrics) ([]models.Metrics, error)
	// GetAllMetrics получает все метрики (gauges и counters).
	GetAllMetrics(ctx context.Context) (gauges []models.Metrics, counters []models.Metrics, err error)
}
