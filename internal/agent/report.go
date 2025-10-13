package agent

import (
	"context"
	"fmt"
	"os"
	"time"
)

func (a *Agent) sendCounters() {
	ctx := context.Background()
	for _, value := range a.metricsStorage.AllCounters(ctx) {
		path := fmt.Sprintf("%s/update/counter/%s/%d", a.port, value.ID, int64(*value.Value))
		_, err := a.client.R().
			SetHeader("Content-Type", "text/plain").
			Post(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

func (a *Agent) sendGauges() {
	ctx := context.Background()
	for _, value := range a.metricsStorage.AllGauges(ctx) {
		path := fmt.Sprintf("%s/update/gauge/%s/%f", a.port, value.ID, *value.Value)
		_, err := a.client.R().
			SetHeader("Content-Type", "text/plain").
			Post(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}
}

func (a *Agent) Report() {
	for {
		time.Sleep(a.reportInterval)
		a.sendCounters()
		a.sendGauges()
	}
}
