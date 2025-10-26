package main

import (
	"github.com/arvaliullin/metrics-collection-service/internal/agent"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
)

func main() {
	ctx, cancel := utils.CtxWithSyscallHandler()
	defer cancel()

	app := agent.New(ctx)

	if err := app.Run(ctx); err != nil {
		app.Logger().Fatal().
			Err(err).
			Msg("failed to run metrics agent")
	}
}
