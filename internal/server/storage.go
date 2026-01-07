package server

import (
	"context"

	"github.com/arvaliullin/metrics-collection-service/internal/config"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/file"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/memory"
	"github.com/arvaliullin/metrics-collection-service/internal/repository/postgres"
	retryrepo "github.com/arvaliullin/metrics-collection-service/internal/repository/retry"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/rs/zerolog"
)

// createStorage выбирает и инициализирует хранилище метрик
func createStorage(
	ctx context.Context,
	cfg *config.ServerConfig,
	logger zerolog.Logger,
) (repository.MetricStorage, error) {
	if cfg.DatabaseConfig.Dsn != "" {
		if err := postgres.RunMigrations(ctx, cfg.DatabaseConfig.Dsn); err != nil {
			return nil, err
		}

		pool, err := postgres.NewPool(ctx, cfg.DatabaseConfig.Dsn)
		if err != nil {
			return nil, err
		}

		retryStrategy := retryutil.NewStrategy(
			retryutil.DefaultDelays,
			postgres.IsConnectionRetryable,
		)

		retryClient := retryrepo.NewPostgresRetryClient(pool, retryStrategy)
		psqlRepo := postgres.NewRepository(retryClient)

		logger.Info().
			Str("backend", "postgres").
			Msg("storage initialized")
		return psqlRepo, nil
	}

	if cfg.FileStoragePath != "" {
		fs, err := file.NewRepository(
			ctx,
			file.Config{
				FilePath:             cfg.FileStoragePath,
				StoreIntervalSeconds: cfg.StoreInterval,
				Restore:              cfg.Restore,
			},
			logger,
		)
		if err != nil {
			return nil, err
		}
		logger.Info().
			Str("backend", "file").
			Str("file", cfg.FileStoragePath).
			Int("interval", cfg.StoreInterval).
			Bool("restore", cfg.Restore).
			Msg("storage initialized")
		return fs, nil
	}

	logger.Info().
		Str("backend", "memory").
		Msg("storage initialized")
	return memory.NewRepository(), nil
}
