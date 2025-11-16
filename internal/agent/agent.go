package agent

import (
	"context"
	"os"

	"github.com/arvaliullin/metrics-collection-service/internal/repository/memory"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

type Agent struct {
	client         *resty.Client
	metricsStorage MetricStorage
	cfg            *Config
	logger         zerolog.Logger
	retryStrategy  *retryutil.Strategy
}

func New(ctx context.Context) *Agent {
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)

	client := resty.New()

	cfg := loadConfig()

	logger.Info().
		Int("poll_interval", cfg.PollInterval).
		Int("report_interval", cfg.ReportInterval).
		Str("address", cfg.Address).
		Msg("agent configuration loaded")

	return &Agent{
		client:         client,
		metricsStorage: memory.NewRepository(),
		cfg:            cfg,
		logger:         logger,
		retryStrategy: retryutil.NewStrategy(
			retryutil.DefaultDelays,
			networkRetryPredicate,
		),
	}
}

func (a *Agent) Logger() *zerolog.Logger {
	return &a.logger
}
