package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/arvaliullin/metrics-collection-service/internal/agent"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := agent.New(ctx)

	if err := app.Run(ctx); err != nil {
		app.Logger().Fatal().
			Err(err).
			Msg("failed to run metrics agent")
	}
}
