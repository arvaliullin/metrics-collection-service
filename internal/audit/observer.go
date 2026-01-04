package audit

import (
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// AuditObserver определяет интерфейс для наблюдателей аудита.
type AuditObserver interface {
	// Notify уведомляет наблюдателя о событии аудита.
	Notify(event models.AuditEvent) error
}
