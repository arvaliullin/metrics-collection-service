package gauge

import (
	"net/http"
	"strconv"

	"github.com/arvaliullin/metrics-collection-service/internal/repository"
)

type GaugeHandler struct {
	memStorage repository.MemStorage
}

func NewGaugeHandler(memStorage repository.MemStorage) *GaugeHandler {
	return &GaugeHandler{
		memStorage: memStorage,
	}
}

func (h *GaugeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idMetric := r.PathValue("id")
	valueStr := r.PathValue("value")
	if idMetric == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	gauge, err := strconv.ParseFloat(valueStr, 0)
	if err != nil {
		http.Error(w, "Некорретное значение метрики", http.StatusBadRequest)
	}

	h.memStorage.UpdateGauge(idMetric, gauge)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
