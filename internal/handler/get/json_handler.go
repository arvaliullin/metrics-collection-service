package get

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/service"
)

var (
	ErrInvalidJSON      = fmt.Errorf("некорректный JSON")
	ErrMethodNotAllowed = fmt.Errorf("метод не поддерживается")
)

type GetJSONHandler struct {
	metricsService ports.ServerMetricsService
}

func NewGetJSONHandler(metricsService ports.ServerMetricsService) *GetJSONHandler {
	return &GetJSONHandler{metricsService: metricsService}
}

// @Summary Получение метрики
// @Description Получает значение метрики (gauge или counter) в формате JSON
// @Tags metrics
// @Accept json
// @Produce json
// @Param metric body models.Metrics true "Метрика для получения"
// @Success 200 {object} object "Метрика с текущим значением"
// @Failure 400 {string} string "Некорректный запрос"
// @Failure 404 {string} string "Метрика не найдена"
// @Failure 405 {string} string "Метод не поддерживается"
// @Router /value [post]
func (h *GetJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, ErrMethodNotAllowed.Error(), http.StatusMethodNotAllowed)
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

	response, err := h.metricsService.GetMetric(r.Context(), metric)
	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, service.ErrNotFound) {
			statusCode = http.StatusNotFound
		}
		http.Error(w, err.Error(), statusCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
