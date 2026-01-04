package audit

import (
	"github.com/arvaliullin/metrics-collection-service/internal/config"
	"github.com/rs/zerolog"
)

// InitializeReceivers создает и инициализирует приёмники аудита на основе конфигурации.
func InitializeReceivers(cfg *config.ServerConfig, notifier *AuditNotifier, logger zerolog.Logger) {
	if cfg.AuditFile != "" {
		fileReceiver := NewFileAuditReceiver(cfg.AuditFile)
		notifier.Subscribe(fileReceiver)
		logger.Info().
			Str("file", cfg.AuditFile).
			Msg("file audit receiver initialized")
	}

	if cfg.AuditURL != "" {
		urlReceiver := NewURLAuditReceiver(cfg.AuditURL)
		notifier.Subscribe(urlReceiver)
		logger.Info().
			Str("url", cfg.AuditURL).
			Msg("URL audit receiver initialized")
	}
}
