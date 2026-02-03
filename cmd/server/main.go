// @title Metrics Collection Service API
// @version 1.0
// @description API для сбора и хранения метрик
// @host localhost:8080
// @BasePath /
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/arvaliullin/metrics-collection-service/internal/server"
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

	app := server.New(ctx)

	if err := app.Run(ctx); err != nil {
		app.Logger().Fatal().
			Err(err).
			Msg("failed to run metrics collection service")
	}
}
