package agent

import (
	"context"
	"sync"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// MetricsBatch представляет собой батч метрик для отправки.
type MetricsBatch struct {
	Metrics []models.Metrics
}

// startWorkerPool запускает пул воркеров для обработки батчей метрик.
func (a *Agent) startWorkerPool(ctx context.Context, jobs <-chan MetricsBatch, wg *sync.WaitGroup) {
	for i := 0; i < a.cfg.RateLimit; i++ {
		wg.Add(1)
		go a.worker(ctx, jobs, wg)
	}
}

// worker обрабатывает задачи из канала jobs.
func (a *Agent) worker(ctx context.Context, jobs <-chan MetricsBatch, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case batch, ok := <-jobs:
			if !ok {
				return
			}
			a.processBatch(ctx, batch.Metrics)
		}
	}
}
