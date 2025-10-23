package agent

import (
	"context"
	"math/rand/v2"
	"runtime"
	"time"
)

func (a *Agent) collectMetrics() {
	ctx := context.Background()
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	a.metricsStorage.UpdateGauge(ctx, "Alloc", float64(memStats.Alloc))
	a.metricsStorage.UpdateGauge(ctx, "BuckHashSys", float64(memStats.BuckHashSys))
	a.metricsStorage.UpdateGauge(ctx, "Frees", float64(memStats.Frees))
	a.metricsStorage.UpdateGauge(ctx, "GCCPUFraction", memStats.GCCPUFraction)
	a.metricsStorage.UpdateGauge(ctx, "GCSys", float64(memStats.GCSys))
	a.metricsStorage.UpdateGauge(ctx, "HeapAlloc", float64(memStats.HeapAlloc))
	a.metricsStorage.UpdateGauge(ctx, "HeapIdle", float64(memStats.HeapIdle))
	a.metricsStorage.UpdateGauge(ctx, "HeapInuse", float64(memStats.HeapInuse))
	a.metricsStorage.UpdateGauge(ctx, "HeapObjects", float64(memStats.HeapObjects))
	a.metricsStorage.UpdateGauge(ctx, "HeapReleased", float64(memStats.HeapReleased))
	a.metricsStorage.UpdateGauge(ctx, "HeapSys", float64(memStats.HeapSys))
	a.metricsStorage.UpdateGauge(ctx, "LastGC", float64(memStats.LastGC))
	a.metricsStorage.UpdateGauge(ctx, "Lookups", float64(memStats.Lookups))
	a.metricsStorage.UpdateGauge(ctx, "MCacheInuse", float64(memStats.MCacheInuse))
	a.metricsStorage.UpdateGauge(ctx, "MCacheSys", float64(memStats.MCacheSys))
	a.metricsStorage.UpdateGauge(ctx, "MSpanInuse", float64(memStats.MSpanInuse))
	a.metricsStorage.UpdateGauge(ctx, "MSpanSys", float64(memStats.MSpanSys))
	a.metricsStorage.UpdateGauge(ctx, "Mallocs", float64(memStats.Mallocs))
	a.metricsStorage.UpdateGauge(ctx, "NextGC", float64(memStats.NextGC))
	a.metricsStorage.UpdateGauge(ctx, "NumForcedGC", float64(memStats.NumForcedGC))
	a.metricsStorage.UpdateGauge(ctx, "NumGC", float64(memStats.NumGC))
	a.metricsStorage.UpdateGauge(ctx, "OtherSys", float64(memStats.OtherSys))
	a.metricsStorage.UpdateGauge(ctx, "PauseTotalNs", float64(memStats.PauseTotalNs))
	a.metricsStorage.UpdateGauge(ctx, "StackInuse", float64(memStats.StackInuse))
	a.metricsStorage.UpdateGauge(ctx, "StackSys", float64(memStats.StackSys))
	a.metricsStorage.UpdateGauge(ctx, "Sys", float64(memStats.Sys))
	a.metricsStorage.UpdateGauge(ctx, "TotalAlloc", float64(memStats.TotalAlloc))

	a.metricsStorage.AddCounter(ctx, "PollCount", 1)
	a.metricsStorage.UpdateGauge(ctx, "RandomValue", rand.Float64())
}

func (a *Agent) Poll() {
	for {
		time.Sleep(a.cfg.pollInterval)
		a.collectMetrics()
	}
}
