package update_test

import (
	"context"
	"fmt"
	"net/url"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/go-resty/resty/v2"
)

func ExampleUpdateJSONHandler_ServeHTTP() {
	serverURL := "http://localhost:8080"
	requestURL, _ := url.JoinPath(serverURL, "/update")
	client := resty.New()

	// Отправка gauge метрики
	var updatedGauge models.Metrics
	allocValue := 1024000.0
	client.R().SetContext(context.Background()).SetHeader("Content-Type", "application/json").
		SetBody(models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &allocValue}).
		SetResult(&updatedGauge).Post(requestURL)
	fmt.Printf("Gauge метрика обновлена: %s = %.0f\n", updatedGauge.ID, *updatedGauge.Value)

	// Отправка counter метрики
	var updatedCounter models.Metrics
	pollCountDelta := int64(1)
	client.R().SetContext(context.Background()).SetHeader("Content-Type", "application/json").
		SetBody(models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &pollCountDelta}).
		SetResult(&updatedCounter).Post(requestURL)
	fmt.Printf("Counter метрика обновлена: %s = %d\n", updatedCounter.ID, *updatedCounter.Delta)

	// Output:
	// Gauge метрика обновлена: Alloc = 1024000
	// Counter метрика обновлена: PollCount = 1
}
