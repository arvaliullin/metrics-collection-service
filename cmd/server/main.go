package main

import (
	"github.com/arvaliullin/metrics-collection-service/internal/server"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
)

func main() {
	ctx, cancel := utils.CtxWithSyscallHandler()
	defer cancel()

	app := server.New(ctx)

	if err := app.Run(ctx); err != nil {
		app.Logger().Fatal().
			Err(err).
			Msg("failed to run metrics collection service")
	}
}
