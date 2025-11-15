package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// compressMetricsBatch сериализует и сжимает список метрик.
func compressMetricsBatch(metrics []models.Metrics) (io.Reader, error) {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return nil, fmt.Errorf("ошибка при сериализации метрик в JSON: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer gz.Close()

	if _, err := gz.Write(jsonData); err != nil {
		return nil, fmt.Errorf("ошибка при сжатии батча метрик: %w", err)
	}

	return &buf, nil
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

	resp, err := a.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Accept-Encoding", "gzip").
		SetBody(compressedBody).
		Post(requestURL)
	if err != nil {
		a.logger.Error().
			Err(err).
			Str("method", "sendBatch").
			Int("metrics_count", len(metrics)).
			Msg("failed to send metrics batch")
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
