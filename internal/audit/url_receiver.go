package audit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

var (
	ErrMarshalAuditEventURL = fmt.Errorf("не удалось сериализовать событие аудита")
	ErrCreateRequest        = fmt.Errorf("не удалось создать HTTP-запрос")
	ErrSendAuditEvent       = fmt.Errorf("не удалось отправить событие аудита")
	ErrUnexpectedStatusCode = fmt.Errorf("неожиданный код статуса ответа")
)

// URLAuditReceiver реализует AuditObserver для отправки событий на удаленный сервер.
type URLAuditReceiver struct {
	url     string
	client  *http.Client
	timeout time.Duration
}

// NewURLAuditReceiver создает новый экземпляр URLAuditReceiver.
func NewURLAuditReceiver(url string) *URLAuditReceiver {
	timeout := 5 * time.Second
	return &URLAuditReceiver{
		url:     url,
		client:  &http.Client{Timeout: timeout},
		timeout: timeout,
	}
}

// Notify отправляет событие аудита на удаленный сервер.
func (r *URLAuditReceiver) Notify(event models.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshalAuditEventURL, err)
	}

	req, err := http.NewRequest(http.MethodPost, r.url, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("%w: %w", ErrCreateRequest, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendAuditEvent, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, resp.StatusCode)
	}

	return nil
}
