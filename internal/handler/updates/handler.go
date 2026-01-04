package updates

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/arvaliullin/metrics-collection-service/internal/audit"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/arvaliullin/metrics-collection-service/internal/utils"
)

var (
	ErrMethodNotSupported  = errors.New("метод не поддерживается")
	ErrInvalidJSON         = errors.New("некорректный JSON")
	ErrEmptyPayload        = errors.New("пустой список метрик")
	ErrMissingID           = errors.New("не указано имя метрики")
	ErrMissingType         = errors.New("не указан тип метрики")
	ErrMissingGaugeValue   = errors.New("не указано значение для gauge")
	ErrMissingCounterDelta = errors.New(
		"не указана дельта для counter",
	)
	ErrInvalidMetricType = errors.New("некорректный тип метрики")
)

type UpdatesHandler struct {
	memStorage    repository.MetricStorage
	auditNotifier *audit.AuditNotifier
}

func NewUpdatesHandler(memStorage repository.MetricStorage, auditNotifier *audit.AuditNotifier) *UpdatesHandler {
	return &UpdatesHandler{
		memStorage:    memStorage,
		auditNotifier: auditNotifier,
	}
}

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

	if len(metrics) == 0 {
		http.Error(w, ErrEmptyPayload.Error(), http.StatusBadRequest)
		return
	}

	for i := range metrics {
		if err := validateMetric(&metrics[i]); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	if err := h.memStorage.BatchUpdate(r.Context(), metrics); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)

	if h.auditNotifier != nil {
		metricNames := make([]string, 0, len(metrics))
		for _, metric := range metrics {
			metricNames = append(metricNames, metric.ID)
		}
		event := models.AuditEvent{
			TS:        time.Now().Unix(),
			Metrics:   metricNames,
			IPAddress: utils.ExtractIPAddress(r),
		}
		h.auditNotifier.NotifyAll(event)
	}
}

func validateMetric(metric *models.Metrics) error {
	if metric.ID == "" {
		return ErrMissingID
	}

	if metric.MType == "" {
		return ErrMissingType
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return ErrMissingGaugeValue
		}
	case models.Counter:
		if metric.Delta == nil {
			if metric.Value != nil {
				delta := int64(*metric.Value)
				metric.Delta = &delta
				metric.Value = nil
			} else {
				return ErrMissingCounterDelta
			}
		}
	default:
		return ErrInvalidMetricType
	}

	return nil
}
