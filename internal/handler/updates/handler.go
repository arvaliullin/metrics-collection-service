package updates

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
)

var (
	ErrMethodNotSupported = errors.New("метод не поддерживается")
	ErrInvalidJSON        = errors.New("некорректный JSON")
)

type UpdatesHandler struct {
	metricsService ports.ServerMetricsService
	auditNotifier  ports.AuditNotifier
}

func NewUpdatesHandler(metricsService ports.ServerMetricsService, auditNotifier ports.AuditNotifier) *UpdatesHandler {
	return &UpdatesHandler{
		metricsService: metricsService,
		auditNotifier:  auditNotifier,
	}
}

// @Summary Пакетное обновление метрик
// @Description Обновляет несколько метрик за один запрос в формате JSON
// @Tags metrics
// @Accept json
// @Produce json
// @Param metrics body []models.Metrics true "Массив метрик для обновления"
// @Success 200 {array} object "Обновленные метрики"
// @Failure 400 {string} string "Некорректный запрос"
// @Failure 405 {string} string "Метод не поддерживается"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /updates [post]
func (h *UpdatesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, ErrMethodNotSupported.Error(), http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil || r.ContentLength == 0 {
		http.Error(w, ErrInvalidJSON.Error(), http.StatusBadRequest)
		return
	}

	var metrics []models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, ErrInvalidJSON.Error(), http.StatusBadRequest)
		return
	}

	updatedMetrics, err := h.metricsService.BatchUpdate(r.Context(), metrics)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.auditNotifier != nil {
		ipAddress := utils.ExtractIPAddress(r)
		metricNames := make([]string, 0, len(updatedMetrics))
		for _, metric := range updatedMetrics {
			metricNames = append(metricNames, metric.ID)
		}
		event := models.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   metricNames,
			IPAddress: ipAddress,
		}
		h.auditNotifier.NotifyAll(event)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedMetrics)
}
