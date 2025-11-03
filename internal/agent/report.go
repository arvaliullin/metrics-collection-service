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

// compressMetric сжимает метрики
func compressMetric(metric models.Metrics) (io.Reader, error) {
	jsonData, err := json.Marshal(metric)

	if err != nil {
		return nil, fmt.Errorf("ошибка при сериализации метрики в JSON: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	defer gz.Close()

	if _, err := gz.Write(jsonData); err != nil {
		return nil, fmt.Errorf("ошибка при сжатии метрики %w", err)
	}

	return &buf, nil
}

// sendCounters передает метрики типа Counter средствами клиента http
func (a *Agent) sendCounters(ctx context.Context) {
	for _, value := range a.metricsStorage.AllCounters(ctx) {
		var delta int64
		if value.Delta != nil {
			delta = *value.Delta
		} else if value.Value != nil {
			delta = int64(*value.Value)
		}

		metric := models.Metrics{
			ID:    value.ID,
			MType: models.Counter,
			Delta: &delta,
		}

		compressedBody, err := compressMetric(metric)
		if err != nil {
			a.logger.Error().
				Err(err).
				Str("method", "sendCounters").
				Str("metric", value.ID).
				Msg("failed to compress counter")
			continue
		}

		requestURL, err := url.JoinPath(a.cfg.GetAddress(), "/update")
		if err != nil {
			a.logger.Error().
				Err(err).
				Str("method", "sendCounters").
				Str("metric", value.ID).
				Str("address", a.cfg.GetAddress()).
				Msg("failed to join URL path")
			continue
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
				Str("method", "sendCounters").
				Str("metric", value.ID).
				Msg("failed to send counter")
			continue
		}

		a.logger.Info().
			Str("method", "sendCounters").
			Str("metric", value.ID).
			Int("status", resp.StatusCode()).
			Msg("counter sent successfully")

		if resp.StatusCode() == http.StatusOK {
			a.metricsStorage.ResetCounter(ctx, value.ID)
		}
	}
}

func (a *Agent) sendGauges(ctx context.Context) {
	for _, value := range a.metricsStorage.AllGauges(ctx) {
		if value.Value == nil {
			continue
		}

		metric := models.Metrics{
			ID:    value.ID,
			MType: models.Gauge,
			Value: value.Value,
		}

		compressedMetric, err := compressMetric(metric)
		if err != nil {
			a.logger.Error().
				Err(err).
				Str("method", "sendGauges").
				Str("metric", value.ID).
				Msg("failed to compress gauge")
			continue
		}

		requestURL, err := url.JoinPath(a.cfg.GetAddress(), "/update")
		if err != nil {
			a.logger.Error().
				Err(err).
				Str("method", "sendGauges").
				Str("metric", value.ID).
				Str("address", a.cfg.GetAddress()).
				Msg("failed to join URL path")
			continue
		}

		resp, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(compressedMetric).
			Post(requestURL)
		if err != nil {
			a.logger.Error().
				Err(err).
				Str("method", "sendGauges").
				Str("metric", value.ID).
				Msg("failed to send gauge")
			continue
		}

		a.logger.Info().
			Str("method", "sendGauges").
			Str("metric", value.ID).
			Int("status", resp.StatusCode()).
			Msg("gauge sent successfully")
	}
}

func (a *Agent) Report(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cfg.GetReportInterval()):
			a.sendCounters(ctx)
			a.sendGauges(ctx)
		}
	}
}
