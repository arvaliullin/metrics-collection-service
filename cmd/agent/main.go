package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arvaliullin/metrics-collection-service/internal/agent"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func orNA(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}

func main() {
	fmt.Fprintf(os.Stdout, "Build version: %s\n", orNA(buildVersion))
	fmt.Fprintf(os.Stdout, "Build date: %s\n", orNA(buildDate))
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", orNA(buildCommit))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	app := agent.New(ctx)

	if err := app.Run(ctx); err != nil {
		app.Logger().Fatal().
			Err(err).
			Msg("failed to run metrics agent")
	}
}
