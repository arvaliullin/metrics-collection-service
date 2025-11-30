package agent

import (
	"context"
	"sync"
)

func (a *Agent) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	var workersWg sync.WaitGroup

	jobs := make(chan MetricsBatch, 10)

	a.logger.Info().
		Int("rate_limit", a.cfg.RateLimit).
		Msg("starting agent with worker pool")

	a.startWorkerPool(ctx, jobs, &workersWg)

	wg.Add(3)

	go func() {
		defer wg.Done()
		a.Poll(ctx)
	}()

	go func() {
		defer wg.Done()
		a.PollGopsutil(ctx)
	}()

	go func() {
		defer wg.Done()
		a.Report(ctx, jobs)
	}()

	<-ctx.Done()
	a.logger.Info().Msg("shutting down agent")

	wg.Wait()

	close(jobs)

	workersWg.Wait()

	return nil
}
