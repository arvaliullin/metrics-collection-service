package agent

import (
	"math/rand/v2"
	"runtime"
	"time"
)

func (a *Agent) collectMetrics() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	a.metricsStorage.UpdateGauge("Alloc", float64(memStats.Alloc))
	a.metricsStorage.UpdateGauge("BuckHashSys", float64(memStats.BuckHashSys))
	a.metricsStorage.UpdateGauge("Frees", float64(memStats.Frees))
	a.metricsStorage.UpdateGauge("GCCPUFraction", memStats.GCCPUFraction)
	a.metricsStorage.UpdateGauge("GCSys", float64(memStats.GCSys))
	a.metricsStorage.UpdateGauge("HeapAlloc", float64(memStats.HeapAlloc))
	a.metricsStorage.UpdateGauge("HeapIdle", float64(memStats.HeapIdle))
	a.metricsStorage.UpdateGauge("HeapInuse", float64(memStats.HeapInuse))
	a.metricsStorage.UpdateGauge("HeapObjects", float64(memStats.HeapObjects))
	a.metricsStorage.UpdateGauge("HeapReleased", float64(memStats.HeapReleased))
	a.metricsStorage.UpdateGauge("HeapSys", float64(memStats.HeapSys))
	a.metricsStorage.UpdateGauge("LastGC", float64(memStats.LastGC))
	a.metricsStorage.UpdateGauge("Lookups", float64(memStats.Lookups))
	a.metricsStorage.UpdateGauge("MCacheInuse", float64(memStats.MCacheInuse))
	a.metricsStorage.UpdateGauge("MCacheSys", float64(memStats.MCacheSys))
	a.metricsStorage.UpdateGauge("MSpanInuse", float64(memStats.MSpanInuse))
	a.metricsStorage.UpdateGauge("MSpanSys", float64(memStats.MSpanSys))
	a.metricsStorage.UpdateGauge("Mallocs", float64(memStats.Mallocs))
	a.metricsStorage.UpdateGauge("NextGC", float64(memStats.NextGC))
	a.metricsStorage.UpdateGauge("NumForcedGC", float64(memStats.NumForcedGC))
	a.metricsStorage.UpdateGauge("NumGC", float64(memStats.NumGC))
	a.metricsStorage.UpdateGauge("OtherSys", float64(memStats.OtherSys))
	a.metricsStorage.UpdateGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	a.metricsStorage.UpdateGauge("StackInuse", float64(memStats.StackInuse))
	a.metricsStorage.UpdateGauge("StackSys", float64(memStats.StackSys))
	a.metricsStorage.UpdateGauge("Sys", float64(memStats.Sys))
	a.metricsStorage.UpdateGauge("TotalAlloc", float64(memStats.TotalAlloc))

	a.metricsStorage.AddCounter("PollCount", 1)
	a.metricsStorage.UpdateGauge("RandomValue", rand.Float64())
}

func (a *Agent) Poll() {
	for {
		time.Sleep(a.pollInterval)
		a.collectMetrics()
	}
}
