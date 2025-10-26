package agent

import (
	"context"
	"fmt"
	"time"
)

func (a *Agent) sendCounters(ctx context.Context) {
	for _, value := range a.metricsStorage.AllCounters(ctx) {
		path := fmt.Sprintf("%s/update/counter/%s/%d", a.cfg.GetAddress(), value.ID, int64(*value.Value))
		_, err := a.client.R().
			SetHeader("Content-Type", "text/plain").
			Post(path)
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
		path := fmt.Sprintf("%s/update/gauge/%s/%f", a.cfg.GetAddress(), value.ID, *value.Value)
		_, err := a.client.R().
			SetHeader("Content-Type", "text/plain").
			Post(path)
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
