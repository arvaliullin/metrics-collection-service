package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

func compressBody(data any) (io.Reader, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}

	return &buf, nil
}

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

		compressedBody, err := compressBody(metric)
		if err != nil {
			a.logger.Error().
				Err(err).
				Str("method", "sendCounters").
				Str("metric", value.ID).
				Msg("failed to compress counter")
			continue
		}

		resp, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(compressedBody).
			Post(fmt.Sprintf("%s/update", a.cfg.GetAddress()))
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

		compressedBody, err := compressBody(metric)
		if err != nil {
			a.logger.Error().
				Err(err).
				Str("method", "sendGauges").
				Str("metric", value.ID).
				Msg("failed to compress gauge")
			continue
		}

		resp, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Accept-Encoding", "gzip").
			SetBody(compressedBody).
			Post(fmt.Sprintf("%s/update", a.cfg.GetAddress()))
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
