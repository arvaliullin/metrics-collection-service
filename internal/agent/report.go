package agent

import (
	"fmt"
	"net/http"
	"time"
)

func (a *Agent) sendCounters() {
	for _, value := range a.metricsStorage.AllCounters() {
		path := fmt.Sprintf("%s/counter/%s/%d", a.baseUrl, value.ID, int64(*value.Value))
		req, _ := http.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Content-Type", "text/plain")
		a.client.Do(req)
	}
}

func (a *Agent) sendGauges() {
	for _, value := range a.metricsStorage.AllGauges() {
		path := fmt.Sprintf("%s/gauge/%s/%f", a.baseUrl, value.ID, *value.Value)
		req, _ := http.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Content-Type", "text/plain")
		a.client.Do(req)
	}
}

func (a *Agent) Report() {
	for {
		time.Sleep(a.reportInterval)
		a.sendCounters()
		a.sendGauges()
	}
}
