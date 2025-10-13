package agent

import (
	"time"

	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	client         *resty.Client
	metricsStorage repository.MemStorage
	pollInterval   time.Duration
	reportInterval time.Duration
	port           string
}

func New(port string, pollIntervalSec, reportIntervalSec int) *Agent {
	return &Agent{
		client:         resty.New(),
		metricsStorage: repository.NewEmptyMemStorage(),
		pollInterval:   time.Duration(pollIntervalSec) * time.Second,
		reportInterval: time.Duration(reportIntervalSec) * time.Second,
		port:           port,
	}
}
