package agent

import (
	"context"
)

// Run запускает агент.
func (a *Agent) Run(ctx context.Context) error {
	a.logger.Info().Msg("starting agent")

	if err := a.metricsService.StartCollection(ctx); err != nil {
		return err
	}

	if err := a.metricsService.StartReporting(ctx); err != nil {
		return err
	}

	<-ctx.Done()
	a.logger.Info().Msg("shutting down agent")

	return nil
}
