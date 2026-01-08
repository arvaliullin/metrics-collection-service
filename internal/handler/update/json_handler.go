package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
)

var (
	ErrInvalidJSON         = fmt.Errorf("некорректный JSON")
	ErrMissingID           = fmt.Errorf("не указано имя метрики")
	ErrMissingType         = fmt.Errorf("не указан тип метрики")
	ErrBothFieldsSet       = fmt.Errorf("указаны и delta и value")
	ErrNoFieldsSet         = fmt.Errorf("не указаны ни delta, ни value")
	ErrMissingGaugeValue   = fmt.Errorf("не указано значение для gauge")
	ErrMissingCounterDelta = fmt.Errorf("не указана дельта для counter")
)

type UpdateJSONHandler struct {
	memStorage    repository.MetricStorage
	auditNotifier ports.AuditNotifier
}

func NewUpdateJSONHandler(memStorage repository.MetricStorage, auditNotifier ports.AuditNotifier) *UpdateJSONHandler {
	return &UpdateJSONHandler{
		memStorage:    memStorage,
		auditNotifier: auditNotifier,
	}
}

// @Summary Обновление метрики
// @Description Обновляет значение метрики (gauge или counter) в формате JSON
// @Tags metrics
// @Accept json
// @Produce json
// @Param metric body models.Metrics true "Метрика для обновления" example({"id":"Alloc","type":"gauge","value":1024000.0})
// @Success 200 {object} object "Обновленная метрика"
// @Failure 400 {string} string "Некорректный запрос"
// @Failure 404 {string} string "Метрика не найдена"
// @Failure 405 {string} string "Метод не поддерживается"
// @Router /update [post]
func (h *UpdateJSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, ErrMethodNotSupported.Error(), http.StatusMethodNotAllowed)
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

	if metric.Delta != nil && metric.Value != nil {
		http.Error(w, ErrBothFieldsSet.Error(), http.StatusBadRequest)
		return
	}

	if metric.Delta == nil && metric.Value == nil {
		http.Error(w, ErrNoFieldsSet.Error(), http.StatusBadRequest)
		return
	}

	var response models.Metrics
	response.ID = metric.ID
	response.MType = metric.MType

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			http.Error(w, ErrMissingGaugeValue.Error(), http.StatusBadRequest)
			return
		}
		h.memStorage.UpdateGauge(r.Context(), metric.ID, *metric.Value)
		value, err := h.memStorage.GetGauge(r.Context(), metric.ID)
		if err != nil {
			http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		response.Value = &value

	case models.Counter:
		if metric.Delta == nil {
			http.Error(w, ErrMissingCounterDelta.Error(), http.StatusBadRequest)
			return
		}
		h.memStorage.AddCounterValue(r.Context(), metric.ID, *metric.Delta)
		delta, err := h.memStorage.GetCounter(r.Context(), metric.ID)
		if err != nil {
			http.Error(w, ErrNotFound.Error(), http.StatusNotFound)
			return
		}
		response.Delta = &delta

	default:
		http.Error(w, ErrInvalidMetricType.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	if h.auditNotifier != nil {
		event := models.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   []string{metric.ID},
			IPAddress: utils.ExtractIPAddress(r),
		}
		h.auditNotifier.NotifyAll(event)
	}
}
