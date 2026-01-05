package memory

import (
	"context"
	"sync"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
)

// Repository предоставляет in-memory реализацию MetricStorage.
type Repository struct {
	mu      sync.RWMutex
	gauge   map[string]float64
	counter map[string]int64
}

// NewRepository создает пустое in-memory хранилище метрик.
func NewRepository() *Repository {
	return &Repository{
		gauge:   make(map[string]float64),
		counter: make(map[string]int64),
	}
}

// UpdateGauge замещает значение метрики типа gauge значением newGauge.
func (r *Repository) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.gauge[id] = newGauge
}

// GetGauge возвращает значение метрики gauge для id.
func (r *Repository) GetGauge(ctx context.Context, id string) (float64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if value, exists := r.gauge[id]; exists {
		return value, nil
	}
	return 0, &GaugeNotFoundError{ID: id}
}

// GetCounter возвращает значение метрики counter для id.
func (r *Repository) GetCounter(ctx context.Context, id string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if value, exists := r.counter[id]; exists {
		return value, nil
	}
	return 0, &CounterNotFoundError{ID: id}
}

// AddCounter добавляет к предыдущему значению newCounter.
func (r *Repository) AddCounter(ctx context.Context, id string, newCounter int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter[id] += newCounter
}

// ResetCounter сбрасывает значение счетчика до нуля.
func (r *Repository) ResetCounter(ctx context.Context, id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter[id] = 0
}

// AddCounterValue добавляет значение delta к счетчику.
func (r *Repository) AddCounterValue(ctx context.Context, id string, delta int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter[id] += delta
}

// AllCounters возвращает все счетчики.
func (r *Repository) AllCounters(ctx context.Context) []models.Metrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(r.counter))

	for id, value := range r.counter {
		delta := value
		metrics = append(metrics, models.Metrics{
			ID:    id,
			MType: models.Counter,
			Delta: &delta,
		})
	}

	return metrics
}

// AllGauges возвращает все gauge-метрики.
func (r *Repository) AllGauges(ctx context.Context) []models.Metrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metrics := make([]models.Metrics, 0, len(r.gauge))

	for id, value := range r.gauge {
		val := value
		metrics = append(metrics, models.Metrics{
			ID:    id,
			MType: models.Gauge,
			Value: &val,
		})
	}

	return metrics
}

// Ping возвращает ошибку, так как подключение к БД не используется.
func (r *Repository) Ping(ctx context.Context) error {
	return ErrPingNotAvailable
}
