package get

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
)

var (
	ErrInvalidJSON      = fmt.Errorf("некорректный JSON")
	ErrMissingID        = fmt.Errorf("не указано имя метрики")
	ErrMissingType      = fmt.Errorf("не указан тип метрики")
	ErrMethodNotAllowed = fmt.Errorf("метод не поддерживается")
)

type GetJSONHandler struct {
	memStorage repository.MetricStorage
}

func NewGetJSONHandler(memStorage repository.MetricStorage) *GetJSONHandler {
	return &GetJSONHandler{memStorage: memStorage}
}

// @Summary Получение метрики
// @Description Получает значение метрики (gauge или counter) в формате JSON
// @Tags metrics
// @Accept json
// @Produce json
// @Param metric body models.Metrics true "Метрика для получения" example({"id":"Alloc","type":"gauge"})
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

	if metric.ID == "" {
		http.Error(w, ErrMissingID.Error(), http.StatusBadRequest)
		return
	}

	if metric.MType == "" {
		http.Error(w, ErrMissingType.Error(), http.StatusBadRequest)
		return
	}

	var response models.Metrics
	response.ID = metric.ID
	response.MType = metric.MType

	switch metric.MType {
	case models.Gauge:
		value, err := h.memStorage.GetGauge(r.Context(), metric.ID)
		if err != nil {
			http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		response.Value = &value

	case models.Counter:
		value, err := h.memStorage.GetCounter(r.Context(), metric.ID)
		if err != nil {
			http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		delta := value
		response.Delta = &delta

	default:
		http.Error(w, ErrInvalidMetricType.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
