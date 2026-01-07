package service

import (
	"context"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/rs/zerolog"
)

// AgentReporter реализует MetricsReporter для отправки метрик.
type AgentReporter struct {
	sender  ports.MetricsSender
	storage repository.MetricStorage
	logger  zerolog.Logger
}

// NewAgentReporter создаёт новый экземпляр AgentReporter.
func NewAgentReporter(sender ports.MetricsSender, storage repository.MetricStorage, logger zerolog.Logger) *AgentReporter {
	return &AgentReporter{
		sender:  sender,
		storage: storage,
		logger:  logger,
	}
}

// BuildBatch строит батч метрик из хранилища.
func (r *AgentReporter) BuildBatch(ctx context.Context) ([]models.Metrics, error) {
	gauges := r.storage.AllGauges(ctx)
	counters := r.storage.AllCounters(ctx)

	if len(gauges) == 0 && len(counters) == 0 {
		return nil, nil
	}

	batch := make([]models.Metrics, 0, len(gauges)+len(counters))
	batch = append(batch, gauges...)
	batch = append(batch, counters...)

	return batch, nil
}

// ReportBatch отправляет батч метрик на сервер.
func (r *AgentReporter) ReportBatch(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	if err := r.sender.Send(ctx, metrics); err != nil {
		return err
	}

	r.resetCounters(ctx, metrics)
	return nil
}

// resetCounters сбрасывает counters после успешной отправки.
func (r *AgentReporter) resetCounters(ctx context.Context, metrics []models.Metrics) {
	for _, metric := range metrics {
		if metric.MType == models.Counter {
			r.storage.ResetCounter(ctx, metric.ID)
		}
	}
}

