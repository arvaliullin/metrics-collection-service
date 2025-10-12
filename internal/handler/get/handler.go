package get

import (
	"fmt"
	"net/http"
	"strconv"

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

	switch typeMetric {
	case models.Gauge:
		value, err := h.memStorage.GetGauge(idMetric)
		if err != nil {
			http.Error(w, ErrNotFound, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))

	case models.Counter:
		value, err := h.memStorage.GetCounter(idMetric)
		if err != nil {
			http.Error(w, ErrNotFound, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%d", value)

	default:
		http.Error(w, ErrInvalidMetricType, http.StatusBadRequest)
		return
	}
}
