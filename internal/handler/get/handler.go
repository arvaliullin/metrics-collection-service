package get

import (
	"fmt"
	"net/http"
	"strconv"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

var (
	ErrNotFound          = fmt.Errorf("метрика с указанным именем не найдена")
	ErrInvalidMetricType = fmt.Errorf("некорректный тип метрики")
)

type GetHandler struct {
	memStorage MetricStorage
}

func NewGetHandler(memStorage MetricStorage) *GetHandler {
	return &GetHandler{memStorage: memStorage}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	typeMetric := r.PathValue("type")
	idMetric := r.PathValue("id")

	if idMetric == "" {
		http.Error(w, ErrNotFound.Error(), http.StatusBadRequest)
		return
	}

	switch typeMetric {
	case models.Gauge:
		value, err := h.memStorage.GetGauge(r.Context(), idMetric)
		if err != nil {
			http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))

	case models.Counter:
		value, err := h.memStorage.GetCounter(r.Context(), idMetric)
		if err != nil {
			http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%d", value)

	default:
		http.Error(w, ErrInvalidMetricType.Error(), http.StatusBadRequest)
		return
	}
}
