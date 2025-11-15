package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/memory"
	"github.com/rs/zerolog"
)

// Config описывает параметры файлового хранилища.
type Config struct {
	FilePath             string
	StoreIntervalSeconds int
	Restore              bool
}

// Repository реализует MetricStorage с сохранением данных на диск.
type Repository struct {
	memStorage    *memory.Repository
	filePath      string
	storeInterval time.Duration
	restore       bool
	logger        zerolog.Logger
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewRepository создает файловое хранилище с опциональным восстановлением и периодическим сохранением.
func NewRepository(ctx context.Context, cfg Config, logger zerolog.Logger) (*Repository, error) {
	fs := &Repository{
		memStorage:    memory.NewRepository(),
		filePath:      cfg.FilePath,
		storeInterval: time.Duration(cfg.StoreIntervalSeconds) * time.Second,
		restore:       cfg.Restore,
		logger:        logger,
		stopChan:      make(chan struct{}),
	}

	if cfg.Restore {
		if err := fs.loadFromFile(ctx); err != nil {
			logger.Warn().Err(err).Msg("failed to restore from file")
		}
	}

	if cfg.StoreIntervalSeconds > 0 {
		fs.wg.Add(1)
		go fs.periodicSave(ctx)
	}

	return fs, nil
}

// UpdateGauge обновляет значение gauge-метрики и инициирует сохранение.
func (r *Repository) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	r.memStorage.UpdateGauge(ctx, id, newGauge)
	r.triggerSave(ctx)
}

// GetGauge возвращает значение gauge-метрики.
func (r *Repository) GetGauge(ctx context.Context, id string) (float64, error) {
	return r.memStorage.GetGauge(ctx, id)
}

// GetCounter возвращает значение counter-метрики.
func (r *Repository) GetCounter(ctx context.Context, id string) (int64, error) {
	return r.memStorage.GetCounter(ctx, id)
}

// AddCounter устанавливает значение counter-метрики и инициирует сохранение.
func (r *Repository) AddCounter(ctx context.Context, id string, newCounter int64) {
	r.memStorage.AddCounterValue(ctx, id, newCounter)
	r.triggerSave(ctx)
}

// AddCounterValue увеличивает counter-метрику на delta и инициирует сохранение.
func (r *Repository) AddCounterValue(ctx context.Context, id string, delta int64) {
	r.memStorage.AddCounterValue(ctx, id, delta)
	r.triggerSave(ctx)
}

// BatchUpdate применяет пакет метрик и инициирует сохранение при необходимости.
func (r *Repository) BatchUpdate(ctx context.Context, metrics []models.Metrics) error {
	if err := r.memStorage.BatchUpdate(ctx, metrics); err != nil {
		return err
	}
	r.triggerSave(ctx)
	return nil
}

// AllCounters возвращает все counter-метрики.
func (r *Repository) AllCounters(ctx context.Context) []models.Metrics {
	return r.memStorage.AllCounters(ctx)
}

// AllGauges возвращает все gauge-метрики.
func (r *Repository) AllGauges(ctx context.Context) []models.Metrics {
	return r.memStorage.AllGauges(ctx)
}

// ResetCounter сбрасывает counter-метрику.
func (r *Repository) ResetCounter(ctx context.Context, id string) {
	r.memStorage.ResetCounter(ctx, id)
}

// Close останавливает периодическое сохранение и сохраняет данные на диск.
func (r *Repository) Close() error {
	close(r.stopChan)
	r.wg.Wait()
	return r.saveToFile(context.TODO())
}

func (r *Repository) triggerSave(ctx context.Context) {
	if r.storeInterval == 0 {
		if err := r.saveToFile(ctx); err != nil {
			r.logger.Error().Err(err).Msg("failed to save to file")
		}
		return
	}
}

func (r *Repository) periodicSave(ctx context.Context) {
	defer r.wg.Done()

	ticker := time.NewTicker(r.storeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.saveToFile(ctx); err != nil {
				r.logger.Error().Err(err).Msg("periodic save failed")
			}
		case <-r.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (r *Repository) saveToFile(ctx context.Context) error {
	metrics := append(r.memStorage.AllGauges(ctx), r.memStorage.AllCounters(ctx)...)

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	dir := filepath.Dir(r.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpFile := r.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, r.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (r *Repository) loadFromFile(ctx context.Context) error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read file: %w", err)
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return fmt.Errorf("failed to unmarshal metrics: %w", err)
	}

	for _, m := range metrics {
		if m.MType == models.Gauge && m.Value != nil {
			r.memStorage.UpdateGauge(ctx, m.ID, *m.Value)
		} else if m.MType == models.Counter {
			if m.Value != nil {
				r.memStorage.AddCounterValue(ctx, m.ID, int64(*m.Value))
			} else if m.Delta != nil {
				r.memStorage.AddCounterValue(ctx, m.ID, *m.Delta)
			}
		}
	}

	return nil
}

// Ping возвращает ошибку, так как подключение к БД не используется.
func (r *Repository) Ping(ctx context.Context) error {
	return fmt.Errorf("проверка соединения с БД недоступна: используется файловое хранилище")
}
