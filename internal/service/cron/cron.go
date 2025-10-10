package cron

import "time"

type Cron struct {
	pollInterval time.Duration
}

func New(pollInterval time.Duration) *Cron {
	return &Cron{pollInterval: pollInterval}
}

type CronFunc func()

func (c Cron) Do(f CronFunc) {
	for {
		time.Sleep(c.pollInterval)
		f()
	}
}
