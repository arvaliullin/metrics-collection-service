package agent

import (
	"context"
	"os"

	agenthttp "github.com/arvaliullin/metrics-collection-service/internal/agent/http"
	httpretry "github.com/arvaliullin/metrics-collection-service/internal/agent/http/retry"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/memory"
	"github.com/arvaliullin/metrics-collection-service/internal/service"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

// Agent представляет агент для сбора и отправки метрик.
type Agent struct {
	metricsService ports.MetricsService
	logger         zerolog.Logger
}

// New создаёт новый экземпляр Agent с инициализированными зависимостями.
func New(ctx context.Context) *Agent {
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)

	cfg := loadConfig()

	logger.Info().
		Int("poll_interval", cfg.PollInterval).
		Int("report_interval", cfg.ReportInterval).
		Str("address", cfg.Address).
		Str("key", cfg.Key).
		Msg("agent configuration loaded")

	restyClient := resty.New()
	retryStrategy := retryutil.NewStrategy(
		retryutil.DefaultDelays,
		networkRetryPredicate,
	)
	httpClient := httpretry.NewHTTPRetryClient(restyClient, retryStrategy)

	metricsStorage := memory.NewRepository()
	collector := service.NewAgentCollector(metricsStorage)
	metricsSender := agenthttp.NewHTTPMetricsSender(httpClient, cfg.GetAddress(), cfg.Key)
	reporter := service.NewAgentReporter(metricsSender, metricsStorage, logger)
	metricsService := service.NewAgentService(
		collector,
		reporter,
		cfg.GetPollInterval(),
		cfg.GetReportInterval(),
		cfg.RateLimit,
		logger,
	)

	return &Agent{
		metricsService: metricsService,
		logger:         logger,
	}
}

func (a *Agent) Logger() *zerolog.Logger {
	return &a.logger
}
