package agent

import (
	"context"
	"os"

	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type Agent struct {
	client         *resty.Client
	metricsStorage MetricStorage
	cfg            *Config
	logger         zerolog.Logger
}

func New(ctx context.Context) *Agent {
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)

	client := resty.New()

	return &Agent{
		client:         client,
		metricsStorage: repository.NewEmptyMemStorage(),
		cfg:            loadConfig(),
		logger:         logger,
	}
}

func (a *Agent) Logger() *zerolog.Logger {
	return &a.logger
}
