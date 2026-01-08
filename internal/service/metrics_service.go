package service

import (
	"context"
	"errors"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/ports"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/rs/zerolog"
)

var (
	ErrNotFound            = errors.New("метрика с указанным именем не найдена")
	ErrInvalidMetricType   = errors.New("некорректный тип метрики")
	ErrMissingID           = errors.New("не указано имя метрики")
	ErrMissingType         = errors.New("не указан тип метрики")
	ErrMissingGaugeValue   = errors.New("не указано значение для gauge")
	ErrMissingCounterDelta = errors.New("не указана дельта для counter")
	ErrBothFieldsSet       = errors.New("указаны и delta и value")
	ErrNoFieldsSet         = errors.New("не указаны ни delta, ни value")
	ErrEmptyPayload        = errors.New("пустой список метрик")
)

var _ ports.ServerMetricsService = (*ServerMetricsService)(nil)

// ServerMetricsService реализует бизнес-логику работы с метриками на сервере.
type ServerMetricsService struct {
	storage repository.MetricStorage
	logger  zerolog.Logger
}

// NewServerMetricsService создает новый экземпляр ServerMetricsService.
func NewServerMetricsService(
	storage repository.MetricStorage,
	logger zerolog.Logger,
) *ServerMetricsService {
	return &ServerMetricsService{
		storage: storage,
		logger:  logger,
	}
}

// UpdateGauge обновляет значение gauge метрики.
func (s *ServerMetricsService) UpdateGauge(ctx context.Context, id string, value float64) error {
	s.storage.UpdateGauge(ctx, id, value)
	return nil
}

// UpdateCounter обновляет значение counter метрики.
func (s *ServerMetricsService) UpdateCounter(ctx context.Context, id string, delta int64) error {
	s.storage.AddCounter(ctx, id, delta)
	return nil
}

// GetGauge получает значение gauge метрики.
func (s *ServerMetricsService) GetGauge(ctx context.Context, id string) (float64, error) {
	return s.storage.GetGauge(ctx, id)
}

// GetCounter получает значение counter метрики.
func (s *ServerMetricsService) GetCounter(ctx context.Context, id string) (int64, error) {
	return s.storage.GetCounter(ctx, id)
}

// GetMetric получает метрику по её описанию.
func (s *ServerMetricsService) GetMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	if err := validateMetricForGet(&metric); err != nil {
		return models.Metrics{}, err
	}

	response := models.Metrics{
		ID:    metric.ID,
		MType: metric.MType,
	}

	switch metric.MType {
	case models.Gauge:
		value, err := s.storage.GetGauge(ctx, metric.ID)
		if err != nil {
			return models.Metrics{}, ErrNotFound
		}
		response.Value = &value

	case models.Counter:
		value, err := s.storage.GetCounter(ctx, metric.ID)
		if err != nil {
			return models.Metrics{}, ErrNotFound
		}
		delta := value
		response.Delta = &delta

	default:
		return models.Metrics{}, ErrInvalidMetricType
	}

	return response, nil
}

// UpdateMetric обновляет метрику по её описанию.
func (s *ServerMetricsService) UpdateMetric(ctx context.Context, metric models.Metrics) (models.Metrics, error) {
	if err := validateMetricForUpdate(&metric); err != nil {
		return models.Metrics{}, err
	}

	response := models.Metrics{
		ID:    metric.ID,
		MType: metric.MType,
	}

	switch metric.MType {
	case models.Gauge:
		if metric.Value == nil {
			return models.Metrics{}, ErrMissingGaugeValue
		}
		s.storage.UpdateGauge(ctx, metric.ID, *metric.Value)
		value, err := s.storage.GetGauge(ctx, metric.ID)
		if err != nil {
			return models.Metrics{}, ErrNotFound
		}
		response.Value = &value

	case models.Counter:
		if metric.Delta == nil {
			return models.Metrics{}, ErrMissingCounterDelta
		}
		s.storage.AddCounterValue(ctx, metric.ID, *metric.Delta)
		delta, err := s.storage.GetCounter(ctx, metric.ID)
		if err != nil {
			return models.Metrics{}, ErrNotFound
		}
		response.Delta = &delta

	default:
		return models.Metrics{}, ErrInvalidMetricType
	}

	return response, nil
}

// BatchUpdate выполняет пакетное обновление метрик.
func (s *ServerMetricsService) BatchUpdate(ctx context.Context, metrics []models.Metrics) ([]models.Metrics, error) {
	if len(metrics) == 0 {
		return nil, ErrEmptyPayload
	}

	for i := range metrics {
		if err := validateMetric(&metrics[i]); err != nil {
			return nil, err
		}
	}

	if err := s.storage.BatchUpdate(ctx, metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

// GetAllMetrics получает все метрики.
func (s *ServerMetricsService) GetAllMetrics(ctx context.Context) (gauges []models.Metrics, counters []models.Metrics, err error) {
	gauges = s.storage.AllGauges(ctx)
	counters = s.storage.AllCounters(ctx)
	return gauges, counters, nil
}

// validateMetric валидирует метрику для пакетного обновления.
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

// validateMetricForUpdate валидирует метрику для обновления.
func validateMetricForUpdate(metric *models.Metrics) error {
	if metric.ID == "" {
		return ErrMissingID
	}

	if metric.MType == "" {
		return ErrMissingType
	}

	if metric.Delta != nil && metric.Value != nil {
		return ErrBothFieldsSet
	}

	if metric.Delta == nil && metric.Value == nil {
		return ErrNoFieldsSet
	}

	return nil
}

// validateMetricForGet валидирует метрику для получения.
func validateMetricForGet(metric *models.Metrics) error {
	if metric.ID == "" {
		return ErrMissingID
	}

	if metric.MType == "" {
		return ErrMissingType
	}

	return nil
}
