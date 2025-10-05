package counter

import (
	"fmt"
	"net/http"

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
	valueMetric := r.PathValue("value")

	w.Write(fmt.Appendf(nil, "idMetric = %s ", idMetric))
	w.Write(fmt.Appendf(nil, "valueMetric = %s ", valueMetric))
}
