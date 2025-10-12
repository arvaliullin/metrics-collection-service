package agent

import (
	"net/http"
	"time"

	"github.com/arvaliullin/metrics-collection-service/internal/repository"
)

type Agent struct {
	client         *http.Client
	metricsStorage repository.MemStorage
	pollInterval   time.Duration
	reportInterval time.Duration
	baseURL        string
}

func New(baseURL string, pollIntervalSec, reportIntervalSec int) *Agent {
	return &Agent{
		client:         &http.Client{},
		metricsStorage: repository.NewEmptyMemStorage(),
		pollInterval:   time.Duration(pollIntervalSec) * time.Second,
		reportInterval: time.Duration(reportIntervalSec) * time.Second,
		baseURL:        baseURL,
	}
}
