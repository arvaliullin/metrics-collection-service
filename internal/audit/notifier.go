package audit

import (
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/rs/zerolog"
)

// AuditNotifier представляет субъект в паттерне Observer.
type AuditNotifier struct {
	observers map[AuditObserver]struct{}
	logger    zerolog.Logger
}

// NewAuditNotifier создает новый экземпляр AuditNotifier.
func NewAuditNotifier(logger zerolog.Logger) *AuditNotifier {
	return &AuditNotifier{
		observers: make(map[AuditObserver]struct{}),
		logger:    logger,
	}
}

// Subscribe добавляет наблюдателя.
func (n *AuditNotifier) Subscribe(observer AuditObserver) {
	n.observers[observer] = struct{}{}
}

// Unsubscribe удаляет наблюдателя.
func (n *AuditNotifier) Unsubscribe(observer AuditObserver) {
	delete(n.observers, observer)
}

// NotifyAll уведомляет всех наблюдателей о событии аудита.
func (n *AuditNotifier) NotifyAll(event models.AuditEvent) {
	for observer := range n.observers {
		if err := observer.Notify(event); err != nil {
			n.logger.Error().
				Err(err).
				Msg("failed to notify audit observer")
		}
	}
}
