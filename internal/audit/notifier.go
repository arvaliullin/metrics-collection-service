package audit

import (
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/rs/zerolog"
)

// AuditNotifier представляет субъект в паттерне Observer.
type AuditNotifier struct {
	observers []AuditObserver
	logger    zerolog.Logger
}

// NewAuditNotifier создает новый экземпляр AuditNotifier.
func NewAuditNotifier(logger zerolog.Logger) *AuditNotifier {
	return &AuditNotifier{
		observers: make([]AuditObserver, 0),
		logger:    logger,
	}
}

// Subscribe добавляет наблюдателя.
func (n *AuditNotifier) Subscribe(observer AuditObserver) {
	n.observers = append(n.observers, observer)
}

// Unsubscribe удаляет наблюдателя.
func (n *AuditNotifier) Unsubscribe(observer AuditObserver) {
	for i, obs := range n.observers {
		if obs == observer {
			n.observers = append(n.observers[:i], n.observers[i+1:]...)
			break
		}
	}
}

// NotifyAll уведомляет всех наблюдателей о событии аудита.
func (n *AuditNotifier) NotifyAll(event models.AuditEvent) {
	for _, observer := range n.observers {
		if err := observer.Notify(event); err != nil {
			n.logger.Error().
				Err(err).
				Msg("failed to notify audit observer")
		}
	}
}
