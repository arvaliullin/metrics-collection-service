package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
)

var (
	ErrInvalidJSON = fmt.Errorf("некорректный JSON")
)

type UpdateJSONHandler struct {
	metricsService ports.ServerMetricsService
	auditNotifier  ports.AuditNotifier
}

func NewUpdateJSONHandler(metricsService ports.ServerMetricsService, auditNotifier ports.AuditNotifier) *UpdateJSONHandler {
	return &UpdateJSONHandler{
		metricsService: metricsService,
		auditNotifier:  auditNotifier,
	}
}

// @Summary Обновление метрики
// @Description Обновляет значение метрики (gauge или counter) в формате JSON
// @Tags metrics
// @Accept json
// @Produce json
// @Param metric body models.Metrics true "Метрика для обновления"
// @Success 200 {object} object "Обновленная метрика"
// @Failure 400 {string} string "Некорректный запрос"
// @Failure 404 {string} string "Метрика не найдена"
// @Failure 405 {string} string "Метод не поддерживается"
// @Router /update [post]
func (h *UpdateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	if r.Body == nil || r.ContentLength == 0 {
		http.Error(w, ErrInvalidJSON.Error(), http.StatusBadRequest)
		return
	}

	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, ErrInvalidJSON.Error(), http.StatusBadRequest)
		return
	}

	response, err := h.metricsService.UpdateMetric(r.Context(), metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if h.auditNotifier != nil {
		ipAddress := utils.ExtractIPAddress(r)
		event := models.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   []string{metric.ID},
			IPAddress: ipAddress,
		}
		h.auditNotifier.NotifyAll(event)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
