package agent

import (
	"context"
	"fmt"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

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

		_, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(metric).
			Post(fmt.Sprintf("%s/update", a.cfg.GetAddress()))
		if err != nil {
			a.logger.Error().
				Err(err).
				Str("method", "sendCounters").
				Str("metric", value.ID).
				Msg("failed to send counter")
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

		_, err := a.client.R().
			SetHeader("Content-Type", "application/json").
			SetBody(metric).
			Post(fmt.Sprintf("%s/update", a.cfg.GetAddress()))
		if err != nil {
			a.logger.Error().
				Err(err).
				Str("method", "sendGauges").
				Str("metric", value.ID).
				Msg("failed to send gauge")
		}
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
