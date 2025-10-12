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
	baseUrl        string
}

func New() *Agent {
	return &Agent{
		client:         &http.Client{},
		metricsStorage: repository.NewEmptyMemStorage(),
		pollInterval:   2 * time.Second,
		reportInterval: 10 * time.Second,
		baseUrl:        `http://localhost:8080`,
	}
}
