package service

import (
	"context"
	"encoding/json"
	"fmt"

	agenthttp "github.com/arvaliullin/metrics-collection-service/internal/http"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
)

var (
	ErrMarshalAuditEventURL = fmt.Errorf("не удалось сериализовать событие аудита")
	ErrSendAuditEvent       = fmt.Errorf("не удалось отправить событие аудита")
	ErrUnexpectedStatusCode = fmt.Errorf("неожиданный код статуса ответа")
)

var _ ports.AuditObserver = (*URLAuditReceiver)(nil)

// URLAuditReceiver реализует AuditObserver для отправки событий на удаленный сервер.
type URLAuditReceiver struct {
	url        string
	httpClient agenthttp.HTTPClient
}

// NewURLAuditReceiver создает новый экземпляр URLAuditReceiver.
func NewURLAuditReceiver(url string, httpClient agenthttp.HTTPClient) *URLAuditReceiver {
	return &URLAuditReceiver{
		url:        url,
		httpClient: httpClient,
	}
}

// Notify отправляет событие аудита на удаленный сервер.
func (r *URLAuditReceiver) Notify(event models.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMarshalAuditEventURL, err)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	resp, err := r.httpClient.Post(context.Background(), r.url, data, headers)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrSendAuditEvent, err)
	}

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return fmt.Errorf("%w: %d", ErrUnexpectedStatusCode, resp.StatusCode())
	}

	return nil
}
