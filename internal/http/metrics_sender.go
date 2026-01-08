package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
)

var _ ports.MetricsSender = (*HTTPMetricsSender)(nil)

// HTTPMetricsSender реализует MetricsSender для отправки метрик через HTTP.
type HTTPMetricsSender struct {
	httpClient HTTPClient
	baseURL    string
	key        string
}

// NewHTTPMetricsSender создаёт новый экземпляр HTTPMetricsSender.
func NewHTTPMetricsSender(httpClient HTTPClient, baseURL, key string) *HTTPMetricsSender {
	return &HTTPMetricsSender{
		httpClient: httpClient,
		baseURL:    baseURL,
		key:        key,
	}
}

// Send отправляет метрики на сервер через HTTP.
func (s *HTTPMetricsSender) Send(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return nil
	}

	compressedBody, err := s.compressMetricsBatch(metrics)
	if err != nil {
		return fmt.Errorf("failed to compress metrics batch: %w", err)
	}

	requestURL, err := url.JoinPath(s.baseURL, "/updates")
	if err != nil {
		return fmt.Errorf("failed to join URL path: %w", err)
	}

	headers := map[string]string{
		"Content-Type":     "application/json",
		"Content-Encoding": "gzip",
		"Accept-Encoding":  "gzip",
	}

	if s.key != "" {
		hash256, hashErr := utils.Hash(compressedBody, s.key)
		if hashErr != nil {
			return fmt.Errorf("failed to compute hash: %w", hashErr)
		}
		headers["HashSHA256"] = fmt.Sprintf("%x", hash256)
	}

	resp, err := s.httpClient.Post(ctx, requestURL, compressedBody, headers)
	if err != nil {
		return fmt.Errorf("failed to send metrics batch: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %d", resp.StatusCode())
	}

	return nil
}

// compressMetricsBatch сериализует и сжимает список метрик.
func (s *HTTPMetricsSender) compressMetricsBatch(metrics []models.Metrics) ([]byte, error) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metrics: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(jsonData); err != nil {
		gz.Close()
		return nil, fmt.Errorf("failed to write to gzip: %w", err)
	}

	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}
