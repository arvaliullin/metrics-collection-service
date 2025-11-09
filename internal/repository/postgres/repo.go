package postgres

import (
	"context"
	"fmt"

	"github.com/arvaliullin/metrics-collection-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository предоставляет реализацию хранилища метрик на PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий PostgreSQL и инициализирует пул соединений.
func NewRepository(ctx context.Context, cfg *config.PostgresConfig) (*Repository, error) {
	if pool, err := pgxpool.New(ctx, cfg.Dsn); err == nil {
		return &Repository{pool}, nil
	} else {
		return nil, fmt.Errorf("%w", err)
	}
}
