package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
	"github.com/go-resty/resty/v2"
)

// compressMetricsBatch сериализует и сжимает список метрик.
func compressMetricsBatch(metrics []models.Metrics) ([]byte, error) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("ошибка при сериализации метрик в JSON: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(jsonData); err != nil {
		gz.Close()
		return nil, fmt.Errorf("ошибка при сжатии батча метрик: %w", err)
	}

	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("ошибка при закрытии gzip писателя: %w", err)
	}

	return buf.Bytes(), nil
}

func (a *Agent) buildBatch(ctx context.Context) []models.Metrics {
	gauges := a.metricsStorage.AllGauges(ctx)
	counters := a.metricsStorage.AllCounters(ctx)

	if len(gauges) == 0 && len(counters) == 0 {
		return nil
	}

	batch := make([]models.Metrics, 0, len(gauges)+len(counters))
	batch = append(batch, gauges...)
	batch = append(batch, counters...)

	return batch
}

func (a *Agent) resetCounters(ctx context.Context, metrics []models.Metrics) {
	for _, metric := range metrics {
		if metric.MType == models.Counter {
			a.metricsStorage.ResetCounter(ctx, metric.ID)
		}
	}
}

func (a *Agent) sendBatch(ctx context.Context) {
	metrics := a.buildBatch(ctx)
	if len(metrics) == 0 {
		return
	}

	compressedBody, err := compressMetricsBatch(metrics)
	if err != nil {
		a.logger.Error().
			Err(err).
			Str("method", "sendBatch").
			Int("metrics_count", len(metrics)).
			Msg("failed to compress metrics batch")
		return
	}

	requestURL, err := url.JoinPath(a.cfg.GetAddress(), "/updates")
	if err != nil {
		a.logger.Error().
			Err(err).
			Str("method", "sendBatch").
			Str("address", a.cfg.GetAddress()).
			Msg("failed to join URL path")
		return
	}

	var resp *resty.Response
	err = a.retryStrategy.DoWithRetry(ctx, func(ctx context.Context) error {
		var reqErr error
		req := a.client.R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(compressedBody)

		if a.cfg.Key != "" {
			hash256, hashErr := utils.Hash(compressedBody, a.cfg.Key)
			if hashErr != nil {
				return hashErr
			}
			req.SetHeader("HashSHA256", fmt.Sprintf("%x", hash256))
		}

		resp, reqErr = req.Post(requestURL)
		return reqErr
	})
	if err != nil {
		a.logger.Error().
			Err(err).
			Str("method", "sendBatch").
			Int("metrics_count", len(metrics)).
			Msg("failed to send metrics batch after retries")
		return
	}

	if resp.StatusCode() == http.StatusOK {
		a.resetCounters(ctx, metrics)
		a.logger.Info().
			Str("method", "sendBatch").
			Int("metrics_count", len(metrics)).
			Msg("metrics batch sent successfully")
		return
	}

	a.logger.Warn().
		Str("method", "sendBatch").
		Int("metrics_count", len(metrics)).
		Int("status", resp.StatusCode()).
		Msg("server returned non-OK status for metrics batch")
}

func (a *Agent) Report(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cfg.GetReportInterval()):
			a.sendBatch(ctx)
		}
	}
}
