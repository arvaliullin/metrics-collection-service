package gauge

import (
	"fmt"
	"net/http"

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
	valueMetric := r.PathValue("value")

	w.Write(fmt.Appendf(nil, "idMetric = %s ", idMetric))
	w.Write(fmt.Appendf(nil, "valueMetric = %s ", valueMetric))
}
