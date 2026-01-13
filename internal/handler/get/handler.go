package get

import (
	"fmt"
	"net/http"
	"strconv"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
)

var (
	ErrInvalidMetricType = fmt.Errorf("некорректный тип метрики")
)

type GetHandler struct {
	metricsService ports.ServerMetricsService
}

func NewGetHandler(metricsService ports.ServerMetricsService) *GetHandler {
	return &GetHandler{metricsService: metricsService}
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	typeMetric := r.PathValue("type")
	idMetric := r.PathValue("id")

	if idMetric == "" {
		http.Error(w, "метрика с указанным именем не найдена", http.StatusBadRequest)
		return
	}

	switch typeMetric {
	case models.Gauge:
		value, err := h.metricsService.GetGauge(r.Context(), idMetric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strconv.FormatFloat(value, 'f', -1, 64)))

	case models.Counter:
		value, err := h.metricsService.GetCounter(r.Context(), idMetric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
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
