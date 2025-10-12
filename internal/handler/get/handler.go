package get

import (
	"fmt"
	"net/http"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
)

const (
	ErrNotFound          = "Метрика с указанным именем не найдена"
	ErrInvalidMetricType = "Некорректный тип метрики"
)

type GetHandler struct {
	memStorage repository.MemStorage
}

func NewGetHandler(memStorage repository.MemStorage) *GetHandler {
	return &GetHandler{memStorage: memStorage}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	typeMetric := r.PathValue("type")
	idMetric := r.PathValue("id")

	if idMetric == "" {
		http.Error(w, ErrNotFound, http.StatusNotFound)
		return
	}

	var result []byte
	switch typeMetric {
	case models.Gauge:
		value, _ := h.memStorage.GetGauge(idMetric)
		result = fmt.Appendf(result, "%f", value)

	case models.Counter:
		value, _ := h.memStorage.GetCounter(idMetric)
		result = fmt.Appendf(result, "%d", value)

	default:
		http.Error(w, ErrInvalidMetricType, http.StatusBadRequest)
		return
	}

	w.Write(result)
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
