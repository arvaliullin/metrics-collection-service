package update

import (
	"encoding/json"
	"fmt"
	"net/http"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
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
	memStorage repository.MetricStorage
}

func NewUpdateJSONHandler(memStorage repository.MetricStorage) *UpdateJSONHandler {
	return &UpdateJSONHandler{
		memStorage: memStorage,
	}
}

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
}
