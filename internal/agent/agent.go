package agent

import (
	"context"
	"os"

	http "github.com/arvaliullin/metrics-collection-service/internal/http"
	httpretry "github.com/arvaliullin/metrics-collection-service/internal/http/retry"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/memory"
	"github.com/arvaliullin/metrics-collection-service/internal/service"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
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
		Str("crypto_key", cfg.CryptoKey).
		Msg("agent configuration loaded")

	restyClient := resty.New()
	retryStrategy := retryutil.NewStrategy(
		retryutil.DefaultDelays,
		http.NetworkRetryPredicate,
	)
	httpClient := httpretry.NewHTTPRetryClient(restyClient, retryStrategy)

	agentIP := ""
	if ip, err := utils.GetOutboundIP(cfg.GetAddress()); err != nil {
		logger.Warn().Err(err).Msg("could not determine outbound IP for X-Real-IP header")
	} else {
		agentIP = ip
	}

	metricsStorage := memory.NewRepository()
	collector := service.NewAgentCollector(metricsStorage)
	metricsSender := http.NewHTTPMetricsSender(httpClient, cfg.GetAddress(), cfg.Key, cfg.CryptoKey, agentIP)
	reporter := service.NewAgentReporter(metricsSender, metricsStorage)
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
