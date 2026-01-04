package audit

import (
	"encoding/json"
	"fmt"
	"os"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

var (
	ErrOpenAuditFile     = fmt.Errorf("не удалось открыть файл аудита")
	ErrMarshalAuditEvent = fmt.Errorf("не удалось сериализовать событие аудита")
	ErrWriteAuditEvent   = fmt.Errorf("не удалось записать событие аудита")
)

// FileAuditReceiver реализует AuditObserver для записи событий в файл.
type FileAuditReceiver struct {
	filePath string
}

// NewFileAuditReceiver создает новый экземпляр FileAuditReceiver.
func NewFileAuditReceiver(filePath string) *FileAuditReceiver {
	return &FileAuditReceiver{
		filePath: filePath,
	}
}

// Notify записывает событие аудита в файл.
func (r *FileAuditReceiver) Notify(event models.AuditEvent) error {
	file, err := os.OpenFile(r.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrOpenAuditFile, err)
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshalAuditEvent, err)
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("%w: %w", ErrWriteAuditEvent, err)
	}

	return nil
}
