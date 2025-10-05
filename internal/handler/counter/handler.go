package counter

import (
	"net/http"
	"strconv"

	"github.com/arvaliullin/metrics-collection-service/internal/repository"
)

type CounterHandler struct {
	memStorage repository.MemStorage
}

func NewCounterHandler(memStorage repository.MemStorage) *CounterHandler {
	return &CounterHandler{
		memStorage: memStorage,
	}
}

func (h *CounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	idMetric := r.PathValue("id")
	valueStr := r.PathValue("value")
	if idMetric == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	counter, err := strconv.ParseInt(valueStr, 10, 0)
	if err != nil {
		http.Error(w, "Некорретное значение метрики", http.StatusBadRequest)
	}

	h.memStorage.AddCounter(idMetric, counter)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
