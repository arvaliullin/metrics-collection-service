package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/rs/zerolog"
)

type FileStorageConfig struct {
	FilePath             string
	StoreIntervalSeconds int
	Restore              bool
}

type FileStorage struct {
	memStorage    *memStorage
	filePath      string
	storeInterval time.Duration
	restore       bool
	logger        zerolog.Logger
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

func NewFileStorage(ctx context.Context, cfg FileStorageConfig, logger zerolog.Logger) (*FileStorage, error) {
	fs := &FileStorage{
		memStorage:    NewEmptyMemStorage(),
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

func (fs *FileStorage) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	fs.memStorage.UpdateGauge(ctx, id, newGauge)
	fs.triggerSave(ctx)
}

func (fs *FileStorage) GetGauge(ctx context.Context, id string) (float64, error) {
	return fs.memStorage.GetGauge(ctx, id)
}

func (fs *FileStorage) GetCounter(ctx context.Context, id string) (int64, error) {
	return fs.memStorage.GetCounter(ctx, id)
}

func (fs *FileStorage) AddCounter(ctx context.Context, id string, newCounter int64) {
	fs.memStorage.AddCounterValue(ctx, id, newCounter)
	fs.triggerSave(ctx)
}

func (fs *FileStorage) AddCounterValue(ctx context.Context, id string, delta int64) {
	fs.memStorage.AddCounterValue(ctx, id, delta)
	fs.triggerSave(ctx)
}

func (fs *FileStorage) AllCounters(ctx context.Context) []models.Metrics {
	return fs.memStorage.AllCounters(ctx)
}

func (fs *FileStorage) AllGauges(ctx context.Context) []models.Metrics {
	return fs.memStorage.AllGauges(ctx)
}

func (fs *FileStorage) ResetCounter(ctx context.Context, id string) {
	fs.memStorage.ResetCounter(ctx, id)
}

func (fs *FileStorage) Close() error {
	close(fs.stopChan)
	fs.wg.Wait()
	return fs.saveToFile(context.TODO())
}

func (fs *FileStorage) triggerSave(ctx context.Context) {
	if fs.storeInterval == 0 {
		if err := fs.saveToFile(ctx); err != nil {
			fs.logger.Error().Err(err).Msg("failed to save to file")
		}
		return
	}

}

func (fs *FileStorage) periodicSave(ctx context.Context) {
	defer fs.wg.Done()

	ticker := time.NewTicker(fs.storeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := fs.saveToFile(ctx); err != nil {
				fs.logger.Error().Err(err).Msg("periodic save failed")
			}
		case <-fs.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (fs *FileStorage) saveToFile(ctx context.Context) error {
	metrics := append(fs.memStorage.AllGauges(ctx), fs.memStorage.AllCounters(ctx)...)

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	dir := filepath.Dir(fs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	tmpFile := fs.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tmpFile, fs.filePath); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (fs *FileStorage) loadFromFile(ctx context.Context) error {
	data, err := os.ReadFile(fs.filePath)
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
			fs.memStorage.UpdateGauge(ctx, m.ID, *m.Value)
		} else if m.MType == models.Counter {
			if m.Value != nil {
				fs.memStorage.AddCounterValue(ctx, m.ID, int64(*m.Value))
			} else if m.Delta != nil {
				fs.memStorage.AddCounterValue(ctx, m.ID, *m.Delta)
			}
		}
	}

	return nil
}
