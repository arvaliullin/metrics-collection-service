package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/arvaliullin/metrics-collection-service/internal/crypto"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
)

const contentEncryptedHeader = "X-Content-Encrypted"

var _ ports.MetricsSender = (*HTTPMetricsSender)(nil)

// HTTPMetricsSender реализует MetricsSender для отправки метрик через HTTP.
type HTTPMetricsSender struct {
	httpClient    HTTPClient
	baseURL       string
	key           string
	cryptoKeyPath string
	publicKey     *rsa.PublicKey
	publicKeyErr  error
	publicKeyOnce sync.Once
}

// NewHTTPMetricsSender создаёт новый экземпляр HTTPMetricsSender.
func NewHTTPMetricsSender(httpClient HTTPClient, baseURL, key, cryptoKeyPath string) *HTTPMetricsSender {
	return &HTTPMetricsSender{
		httpClient:    httpClient,
		baseURL:       baseURL,
		key:           key,
		cryptoKeyPath: cryptoKeyPath,
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

	body, err := s.encryptBody(compressedBody)
	if err != nil {
		return err
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
	if s.cryptoKeyPath != "" {
		headers[contentEncryptedHeader] = "true"
	}
	if s.key != "" {
		hash256, hashErr := utils.Hash(body, s.key)
		if hashErr != nil {
			return fmt.Errorf("failed to compute hash: %w", hashErr)
		}
		headers["HashSHA256"] = fmt.Sprintf("%x", hash256)
	}

	resp, err := s.httpClient.Post(ctx, requestURL, body, headers)
	if err != nil {
		return fmt.Errorf("failed to send metrics batch: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %d", resp.StatusCode())
	}

	return nil
}

// encryptBody при наличии пути к публичному ключу шифрует данные; иначе возвращает их без изменений.
func (s *HTTPMetricsSender) encryptBody(data []byte) ([]byte, error) {
	if s.cryptoKeyPath == "" {
		return data, nil
	}
	s.publicKeyOnce.Do(func() {
		s.publicKey, s.publicKeyErr = crypto.LoadPublicKey(s.cryptoKeyPath)
	})
	if s.publicKeyErr != nil {
		return nil, fmt.Errorf("load public key: %w", s.publicKeyErr)
	}
	encrypted, err := crypto.Encrypt(s.publicKey, data)
	if err != nil {
		return nil, &EncryptBodyError{PayloadSize: len(data), Err: err}
	}
	return encrypted, nil
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
