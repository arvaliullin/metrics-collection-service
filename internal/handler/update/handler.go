package update

import (
	"fmt"
	"net/http"
	"strconv"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
)

var (
	ErrMethodNotSupported = fmt.Errorf("метод не поддерживается")
	ErrNotFound           = fmt.Errorf("метрика с указанным именем не найдена")
	ErrInvalidMetricValue = fmt.Errorf("некорретное значение метрики")
	ErrInvalidMetricType  = fmt.Errorf("некорректный тип метрики")
)

type UpdateHandler struct {
	memStorage repository.MetricStorage
}

func NewUpdateHandler(memStorage repository.MetricStorage) *UpdateHandler {
	return &UpdateHandler{
		memStorage: memStorage,
	}
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, ErrMethodNotSupported.Error(), http.StatusMethodNotAllowed)
		return
	}

	typeMetric := r.PathValue("type")
	idMetric := r.PathValue("id")
	valueStr := r.PathValue("value")

	if idMetric == "" {
		http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
		return
	}

	switch typeMetric {
	case models.Gauge:
		gauge, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, ErrInvalidMetricValue.Error(), http.StatusBadRequest)
			return
		}
		h.memStorage.UpdateGauge(r.Context(), idMetric, gauge)

	case models.Counter:
		counter, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, ErrInvalidMetricValue.Error(), http.StatusBadRequest)
			return
		}
		h.memStorage.AddCounter(r.Context(), idMetric, counter)

	default:
		http.Error(w, ErrInvalidMetricType.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
