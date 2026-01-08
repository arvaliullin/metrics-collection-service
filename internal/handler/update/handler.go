package update

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
)

var (
	ErrMethodNotSupported = fmt.Errorf("метод не поддерживается")
	ErrInvalidMetricValue = fmt.Errorf("некорретное значение метрики")
	ErrInvalidMetricType  = fmt.Errorf("некорректный тип метрики")
)

type UpdateHandler struct {
	metricsService ports.ServerMetricsService
	auditNotifier  ports.AuditNotifier
}

func NewUpdateHandler(metricsService ports.ServerMetricsService, auditNotifier ports.AuditNotifier) *UpdateHandler {
	return &UpdateHandler{
		metricsService: metricsService,
		auditNotifier:  auditNotifier,
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

	switch typeMetric {
	case models.Gauge:
		gauge, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			http.Error(w, ErrInvalidMetricValue.Error(), http.StatusBadRequest)
			return
		}
		if err := h.metricsService.UpdateGauge(r.Context(), idMetric, gauge); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	case models.Counter:
		counter, err := strconv.ParseInt(valueStr, 10, 64)
		if err != nil {
			http.Error(w, ErrInvalidMetricValue.Error(), http.StatusBadRequest)
			return
		}
		if err := h.metricsService.UpdateCounter(r.Context(), idMetric, counter); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, ErrInvalidMetricType.Error(), http.StatusBadRequest)
		return
	}

	if h.auditNotifier != nil {
		ipAddress := utils.ExtractIPAddress(r)
		event := models.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   []string{idMetric},
			IPAddress: ipAddress,
		}
		h.auditNotifier.NotifyAll(event)
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusOK)
}
