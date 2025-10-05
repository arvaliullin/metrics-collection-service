package update

import (
	"net/http"
	"strconv"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
)

type UpdateHandler struct {
	memStorage repository.MemStorage
}

func NewUpdateHandler(memStorage repository.MemStorage) *UpdateHandler {
	return &UpdateHandler{
		memStorage: memStorage,
	}
}

func (h *UpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	typeMetric := r.PathValue("type")
	idMetric := r.PathValue("id")
	valueStr := r.PathValue("value")

	if idMetric == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	switch typeMetric {
	case models.Gauge:
		gauge, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, "Некорретное значение метрики", http.StatusBadRequest)
			return
		}
		h.memStorage.UpdateGauge(idMetric, gauge)

	case models.Counter:
		counter, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, "Некорретное значение метрики", http.StatusBadRequest)
			return
		}
		h.memStorage.AddCounter(idMetric, counter)

	default:
		http.Error(w, "Некорректный тип метрики", http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
