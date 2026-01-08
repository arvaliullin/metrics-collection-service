package audit

import (
	"time"

	"github.com/arvaliullin/metrics-collection-service/internal/config"
	agenthttp "github.com/arvaliullin/metrics-collection-service/internal/http"
	httpretry "github.com/arvaliullin/metrics-collection-service/internal/http/retry"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

// InitializeReceivers создает и инициализирует приёмники аудита на основе конфигурации.
func InitializeReceivers(cfg *config.ServerConfig, notifier Notifier, logger zerolog.Logger) {
	if cfg.AuditFile != "" {
		fileReceiver := NewFileAuditReceiver(cfg.AuditFile)
		notifier.Subscribe(fileReceiver)
		logger.Info().
			Str("file", cfg.AuditFile).
			Msg("file audit receiver initialized")
	}

	if cfg.AuditURL != "" {
		restyClient := resty.New().SetTimeout(5 * time.Second)
		retryStrategy := retryutil.NewStrategy(
			retryutil.DefaultDelays,
			agenthttp.NetworkRetryPredicate,
		)
		httpClient := httpretry.NewHTTPRetryClient(restyClient, retryStrategy)
		urlReceiver := NewURLAuditReceiver(cfg.AuditURL, httpClient)
		notifier.Subscribe(urlReceiver)
		logger.Info().
			Str("url", cfg.AuditURL).
			Msg("URL audit receiver initialized")
	}
}
