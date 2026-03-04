package agent

import (
	"context"
	"os"

	grpcclient "github.com/arvaliullin/metrics-collection-service/internal/grpc"
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
	closer         func() error
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
		Str("grpc_address", cfg.GRPCAddress).
		Str("key", cfg.Key).
		Str("crypto_key", cfg.CryptoKey).
		Msg("agent configuration loaded")

	agentIP := ""
	resolveAddr := cfg.GetAddress()
	if cfg.GRPCAddress != "" {
		resolveAddr = cfg.GRPCAddress
	}
	if ip, err := utils.GetOutboundIP(resolveAddr); err != nil {
		logger.Warn().Err(err).Msg("could not determine outbound IP")
	} else {
		agentIP = ip
	}

	metricsStorage := memory.NewRepository()
	collector := service.NewAgentCollector(metricsStorage)

	var metricsSender ports.MetricsSender
	var closer func() error
	if cfg.GRPCAddress != "" {
		grpcSender, err := grpcclient.NewGRPCMetricsSender(cfg.GRPCAddress, agentIP)
		if err != nil {
			logger.Fatal().Err(err).Msg("failed to create gRPC metrics sender")
		}
		metricsSender = grpcSender
		closer = grpcSender.Close
	} else {
		restyClient := resty.New()
		retryStrategy := retryutil.NewStrategy(
			retryutil.DefaultDelays,
			http.NetworkRetryPredicate,
		)
		httpClient := httpretry.NewHTTPRetryClient(restyClient, retryStrategy)
		metricsSender = http.NewHTTPMetricsSender(httpClient, cfg.GetAddress(),
			http.WithKey(cfg.Key),
			http.WithCryptoKey(cfg.CryptoKey),
			http.WithAgentIP(agentIP))
	}

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
		closer:         closer,
	}
}

func (a *Agent) Logger() *zerolog.Logger {
	return &a.logger
}
