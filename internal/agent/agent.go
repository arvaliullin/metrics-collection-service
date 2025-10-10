package agent

import (
	"time"

	"github.com/arvaliullin/metrics-collection-service/internal/service/cron"
)

type Agent struct {
	cron *cron.Cron
}

func New() *Agent {
	return &Agent{
		cron: cron.New(time.Second * 2),
	}
}
