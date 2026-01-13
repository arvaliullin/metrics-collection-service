package service

import (
	"context"
	"fmt"
	"math/rand/v2"
	"runtime"

	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

// AgentCollector реализует MetricsCollector для сбора метрик.
type AgentCollector struct {
	storage repository.MetricStorage
}

// NewAgentCollector создаёт новый экземпляр AgentCollector.
func NewAgentCollector(storage repository.MetricStorage) *AgentCollector {
	return &AgentCollector{
		storage: storage,
	}
}

// CollectRuntimeMetrics собирает метрики runtime.
func (c *AgentCollector) CollectRuntimeMetrics(ctx context.Context) error {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	c.storage.UpdateGauge(ctx, "Alloc", float64(memStats.Alloc))
	c.storage.UpdateGauge(ctx, "BuckHashSys", float64(memStats.BuckHashSys))
	c.storage.UpdateGauge(ctx, "Frees", float64(memStats.Frees))
	c.storage.UpdateGauge(ctx, "GCCPUFraction", memStats.GCCPUFraction)
	c.storage.UpdateGauge(ctx, "GCSys", float64(memStats.GCSys))
	c.storage.UpdateGauge(ctx, "HeapAlloc", float64(memStats.HeapAlloc))
	c.storage.UpdateGauge(ctx, "HeapIdle", float64(memStats.HeapIdle))
	c.storage.UpdateGauge(ctx, "HeapInuse", float64(memStats.HeapInuse))
	c.storage.UpdateGauge(ctx, "HeapObjects", float64(memStats.HeapObjects))
	c.storage.UpdateGauge(ctx, "HeapReleased", float64(memStats.HeapReleased))
	c.storage.UpdateGauge(ctx, "HeapSys", float64(memStats.HeapSys))
	c.storage.UpdateGauge(ctx, "LastGC", float64(memStats.LastGC))
	c.storage.UpdateGauge(ctx, "Lookups", float64(memStats.Lookups))
	c.storage.UpdateGauge(ctx, "MCacheInuse", float64(memStats.MCacheInuse))
	c.storage.UpdateGauge(ctx, "MCacheSys", float64(memStats.MCacheSys))
	c.storage.UpdateGauge(ctx, "MSpanInuse", float64(memStats.MSpanInuse))
	c.storage.UpdateGauge(ctx, "MSpanSys", float64(memStats.MSpanSys))
	c.storage.UpdateGauge(ctx, "Mallocs", float64(memStats.Mallocs))
	c.storage.UpdateGauge(ctx, "NextGC", float64(memStats.NextGC))
	c.storage.UpdateGauge(ctx, "NumForcedGC", float64(memStats.NumForcedGC))
	c.storage.UpdateGauge(ctx, "NumGC", float64(memStats.NumGC))
	c.storage.UpdateGauge(ctx, "OtherSys", float64(memStats.OtherSys))
	c.storage.UpdateGauge(ctx, "PauseTotalNs", float64(memStats.PauseTotalNs))
	c.storage.UpdateGauge(ctx, "StackInuse", float64(memStats.StackInuse))
	c.storage.UpdateGauge(ctx, "StackSys", float64(memStats.StackSys))
	c.storage.UpdateGauge(ctx, "Sys", float64(memStats.Sys))
	c.storage.UpdateGauge(ctx, "TotalAlloc", float64(memStats.TotalAlloc))

	c.storage.AddCounter(ctx, "PollCount", 1)
	c.storage.UpdateGauge(ctx, "RandomValue", rand.Float64())

	return nil
}

// CollectGopsutilMetrics собирает дополнительные метрики системы через gopsutil.
func (c *AgentCollector) CollectGopsutilMetrics(ctx context.Context) error {
	vmStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get virtual memory stats: %w", err)
	}
	c.storage.UpdateGauge(ctx, "TotalMemory", float64(vmStat.Total))
	c.storage.UpdateGauge(ctx, "FreeMemory", float64(vmStat.Free))

	cpuPercents, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		return fmt.Errorf("failed to get CPU utilization: %w", err)
	}
	for i, percent := range cpuPercents {
		metricName := fmt.Sprintf("CPUutilization%d", i+1)
		c.storage.UpdateGauge(ctx, metricName, percent)
	}

	return nil
}
