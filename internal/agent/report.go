package agent

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func (a *Agent) sendCounters() {
	for _, value := range a.metricsStorage.AllCounters() {
		path := fmt.Sprintf("%s/update/counter/%s/%d", a.baseURL, value.ID, int64(*value.Value))
		req, _ := http.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Content-Type", "text/plain")
		if response, err := a.client.Do(req); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else {
			response.Body.Close()
		}

	}
}

func (a *Agent) sendGauges() {
	for _, value := range a.metricsStorage.AllGauges() {
		path := fmt.Sprintf("%s/update/gauge/%s/%f", a.baseURL, value.ID, *value.Value)
		req, _ := http.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("Content-Type", "text/plain")
		if response, err := a.client.Do(req); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		} else {
			response.Body.Close()
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
