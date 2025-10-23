package agent

import (
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/go-resty/resty/v2"
)

type Agent struct {
	client         *resty.Client
	metricsStorage repository.MemStorage
	cfg            *Config
}

func New() *Agent {
	return &Agent{
		client:         resty.New(),
		metricsStorage: repository.NewEmptyMemStorage(),
		cfg:            loadConfig(),
	}
}
