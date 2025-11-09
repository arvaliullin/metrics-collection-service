package filestorage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/memorystorage"
	"github.com/rs/zerolog"
)

// Config описывает параметры файлового хранилища.
type Config struct {
	FilePath             string
	StoreIntervalSeconds int
	Restore              bool
}

// Storage реализует MetricStorage с сохранением данных на диск.
type Storage struct {
	memStorage    *memorystorage.Storage
	filePath      string
	storeInterval time.Duration
	restore       bool
	logger        zerolog.Logger
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// New создает файловое хранилище с опциональным восстановлением и периодическим сохранением.
func New(ctx context.Context, cfg Config, logger zerolog.Logger) (*Storage, error) {
	fs := &Storage{
		memStorage:    memorystorage.New(),
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
func (s *Storage) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	s.memStorage.UpdateGauge(ctx, id, newGauge)
	s.triggerSave(ctx)
}

// GetGauge возвращает значение gauge-метрики.
func (s *Storage) GetGauge(ctx context.Context, id string) (float64, error) {
	return s.memStorage.GetGauge(ctx, id)
}

// GetCounter возвращает значение counter-метрики.
func (s *Storage) GetCounter(ctx context.Context, id string) (int64, error) {
	return s.memStorage.GetCounter(ctx, id)
}

// AddCounter устанавливает значение counter-метрики и инициирует сохранение.
func (s *Storage) AddCounter(ctx context.Context, id string, newCounter int64) {
	s.memStorage.AddCounterValue(ctx, id, newCounter)
	s.triggerSave(ctx)
}

// AddCounterValue увеличивает counter-метрику на delta и инициирует сохранение.
func (s *Storage) AddCounterValue(ctx context.Context, id string, delta int64) {
	s.memStorage.AddCounterValue(ctx, id, delta)
	s.triggerSave(ctx)
}

// AllCounters возвращает все counter-метрики.
func (s *Storage) AllCounters(ctx context.Context) []models.Metrics {
	return s.memStorage.AllCounters(ctx)
}

// AllGauges возвращает все gauge-метрики.
func (s *Storage) AllGauges(ctx context.Context) []models.Metrics {
	return s.memStorage.AllGauges(ctx)
}

// ResetCounter сбрасывает counter-метрику.
func (s *Storage) ResetCounter(ctx context.Context, id string) {
	s.memStorage.ResetCounter(ctx, id)
}

// Close останавливает периодическое сохранение и сохраняет данные на диск.
func (s *Storage) Close() error {
	close(s.stopChan)
	s.wg.Wait()
	return s.saveToFile(context.TODO())
}

func (s *Storage) triggerSave(ctx context.Context) {
	if s.storeInterval == 0 {
		if err := s.saveToFile(ctx); err != nil {
			s.logger.Error().Err(err).Msg("failed to save to file")
		}
		return
	}
}

func (s *Storage) periodicSave(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(s.storeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.saveToFile(ctx); err != nil {
				s.logger.Error().Err(err).Msg("periodic save failed")
			}
		case <-s.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (s *Storage) saveToFile(ctx context.Context) error {
	metrics := append(s.memStorage.AllGauges(ctx), s.memStorage.AllCounters(ctx)...)

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, s.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (s *Storage) loadFromFile(ctx context.Context) error {
	data, err := os.ReadFile(s.filePath)
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
			s.memStorage.UpdateGauge(ctx, m.ID, *m.Value)
		} else if m.MType == models.Counter {
			if m.Value != nil {
				s.memStorage.AddCounterValue(ctx, m.ID, int64(*m.Value))
			} else if m.Delta != nil {
				s.memStorage.AddCounterValue(ctx, m.ID, *m.Delta)
			}
		}
	}

	return nil
}
