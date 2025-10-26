package agent

import (
	"context"
	"sync"
)

func (a *Agent) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(2)

	a.logger.Info().Msg("starting agent")

	go func() {
		defer wg.Done()
		a.Poll(ctx)
	}()

	go func() {
		defer wg.Done()
		a.Report(ctx)
	}()

	<-ctx.Done()
	a.logger.Info().Msg("shutting down agent")

	wg.Wait()
	return nil
}
