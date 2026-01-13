package service

import (
	"time"

	"github.com/arvaliullin/metrics-collection-service/internal/config"
	agenthttp "github.com/arvaliullin/metrics-collection-service/internal/http"
	httpretry "github.com/arvaliullin/metrics-collection-service/internal/http/retry"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

var _ ports.AuditNotifier = (*AuditService)(nil)

// AuditService координирует работу аудита.
type AuditService struct {
	notifier ports.AuditNotifier
	logger   zerolog.Logger
}

// NewAuditService создает новый экземпляр AuditService.
func NewAuditService(notifier ports.AuditNotifier, logger zerolog.Logger) *AuditService {
	return &AuditService{
		notifier: notifier,
		logger:   logger,
	}
}

// Subscribe добавляет наблюдателя.
func (s *AuditService) Subscribe(observer ports.AuditObserver) {
	s.notifier.Subscribe(observer)
}

// Unsubscribe удаляет наблюдателя.
func (s *AuditService) Unsubscribe(observer ports.AuditObserver) {
	s.notifier.Unsubscribe(observer)
}

// NotifyAll уведомляет всех наблюдателей о событии аудита.
func (s *AuditService) NotifyAll(event models.AuditEvent) {
	s.notifier.NotifyAll(event)
}

// InitializeReceivers создает и инициализирует приёмники аудита на основе конфигурации.
func (s *AuditService) InitializeReceivers(cfg *config.ServerConfig) {
	if cfg.AuditFile != "" {
		fileReceiver := NewFileAuditReceiver(cfg.AuditFile)
		s.Subscribe(fileReceiver)
		s.logger.Info().
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
		s.Subscribe(urlReceiver)
		s.logger.Info().
			Str("url", cfg.AuditURL).
			Msg("URL audit receiver initialized")
	}
}
