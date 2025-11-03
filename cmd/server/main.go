package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/arvaliullin/metrics-collection-service/internal/server"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := server.New(ctx)

	if err := app.Run(ctx); err != nil {
		app.Logger().Fatal().
			Err(err).
			Msg("failed to run metrics collection service")
	}
}
