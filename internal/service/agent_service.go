package service

import (
	"context"
	"sync"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/rs/zerolog"
)

var _ ports.MetricsService = (*AgentService)(nil)

// AgentService координирует работу сбора и отправки метрик.
type AgentService struct {
	collector      ports.MetricsCollector
	reporter       ports.MetricsReporter
	pollInterval   time.Duration
	reportInterval time.Duration
	rateLimit      int
	logger         zerolog.Logger
}

// NewAgentService создаёт новый экземпляр AgentService.
func NewAgentService(collector ports.MetricsCollector, reporter ports.MetricsReporter, pollInterval, reportInterval time.Duration, rateLimit int, logger zerolog.Logger) *AgentService {
	return &AgentService{
		collector:      collector,
		reporter:       reporter,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		rateLimit:      rateLimit,
		logger:         logger,
	}
}

// StartCollection запускает периодический сбор метрик.
func (s *AgentService) StartCollection(ctx context.Context) error {
	go s.collectRuntimeMetrics(ctx)
	go s.collectGopsutilMetrics(ctx)
	return nil
}

// StartReporting запускает периодическую отправку метрик через worker pool.
func (s *AgentService) StartReporting(ctx context.Context) error {
	jobs := make(chan models.MetricsBatch, 10)
	var wg sync.WaitGroup

	for range s.rateLimit {
		wg.Add(1)
		go s.worker(ctx, jobs, &wg)
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.reportingLoop(ctx, jobs)
	}()

	<-ctx.Done()
	close(jobs)
	wg.Wait()

	return nil
}

// reportingLoop периодически создает батчи метрик и отправляет их в канал jobs.
func (s *AgentService) reportingLoop(ctx context.Context, jobs chan<- models.MetricsBatch) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.reportInterval):
			metrics, err := s.reporter.BuildBatch(ctx)
			if err != nil {
				continue
			}
			if len(metrics) == 0 {
				continue
			}
			select {
			case jobs <- models.MetricsBatch{Metrics: metrics}:
			case <-ctx.Done():
				return
			}
		}
	}
}

// worker обрабатывает задачи из канала jobs.
func (s *AgentService) worker(ctx context.Context, jobs <-chan models.MetricsBatch, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case batch, ok := <-jobs:
			if !ok {
				return
			}
			if err := s.reporter.ReportBatch(ctx, batch.Metrics); err != nil {
				s.logger.Error().
					Err(err).
					Int("metrics_count", len(batch.Metrics)).
					Msg("failed to report metrics batch")
			} else {
				s.logger.Info().
					Int("metrics_count", len(batch.Metrics)).
					Msg("metrics batch sent successfully")
			}
		}
	}
}

// collectRuntimeMetrics собирает runtime метрики с заданной периодичностью.
func (s *AgentService) collectRuntimeMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.pollInterval):
			s.collector.CollectRuntimeMetrics(ctx)
		}
	}
}

// collectGopsutilMetrics собирает gopsutil метрики с заданной периодичностью.
func (s *AgentService) collectGopsutilMetrics(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(s.pollInterval):
			s.collector.CollectGopsutilMetrics(ctx)
		}
	}
}
