package update_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/memory"
)

func ExampleUpdateJSONHandler_ServeHTTP() {
	// Создаём изолированный репозиторий для теста
	repo := memory.NewRepository()
	handler := update.NewUpdateJSONHandler(repo, nil)

	// Создаём тестовый HTTP сервер
	mux := http.NewServeMux()
	mux.Handle("/update", handler)
	server := httptest.NewServer(mux)
	defer server.Close()

	// Отправка gauge метрики
	allocValue := 1024000.0
	gaugeBody, _ := json.Marshal(models.Metrics{ID: "Alloc", MType: models.Gauge, Value: &allocValue})
	gaugeReq, _ := http.NewRequest(http.MethodPost, server.URL+"/update", bytes.NewReader(gaugeBody))
	gaugeReq.Header.Set("Content-Type", "application/json")
	gaugeResp, _ := http.DefaultClient.Do(gaugeReq)
	var updatedGauge models.Metrics
	json.NewDecoder(gaugeResp.Body).Decode(&updatedGauge)
	gaugeResp.Body.Close()
	fmt.Printf("Gauge метрика обновлена: %s = %.0f\n", updatedGauge.ID, *updatedGauge.Value)

	// Отправка counter метрики
	pollCountDelta := int64(1)
	counterBody, _ := json.Marshal(models.Metrics{ID: "PollCount", MType: models.Counter, Delta: &pollCountDelta})
	counterReq, _ := http.NewRequest(http.MethodPost, server.URL+"/update", bytes.NewReader(counterBody))
	counterReq.Header.Set("Content-Type", "application/json")
	counterResp, _ := http.DefaultClient.Do(counterReq)
	var updatedCounter models.Metrics
	json.NewDecoder(counterResp.Body).Decode(&updatedCounter)
	counterResp.Body.Close()
	fmt.Printf("Counter метрика обновлена: %s = %d\n", updatedCounter.ID, *updatedCounter.Delta)

	// Output:
	// Gauge метрика обновлена: Alloc = 1024000
	// Counter метрика обновлена: PollCount = 1
}
