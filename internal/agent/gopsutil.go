package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// collectGopsutilMetrics собирает дополнительные метрики системы через gopsutil.
func (a *Agent) collectGopsutilMetrics(ctx context.Context) {
	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		a.logger.Error().
			Err(err).
			Str("method", "collectGopsutilMetrics").
			Msg("failed to get virtual memory stats")
	} else {
		a.metricsStorage.UpdateGauge(ctx, "TotalMemory", float64(vmStat.Total))
		a.metricsStorage.UpdateGauge(ctx, "FreeMemory", float64(vmStat.Free))
	}

	cpuPercents, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		a.logger.Error().
			Err(err).
			Str("method", "collectGopsutilMetrics").
			Msg("failed to get CPU utilization")
	} else {
		for i, percent := range cpuPercents {
			metricName := fmt.Sprintf("CPUutilization%d", i+1)
			a.metricsStorage.UpdateGauge(ctx, metricName, percent)
		}
	}
}

// PollGopsutil выполняет опрос дополнительных метрик через gopsutil с заданной периодичностью.
func (a *Agent) PollGopsutil(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(a.cfg.GetPollInterval()):
			a.collectGopsutilMetrics(ctx)
		}
	}
}
