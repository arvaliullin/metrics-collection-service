package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/arvaliullin/metrics-collection-service/internal/config"
	_ "github.com/arvaliullin/metrics-collection-service/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Repository предоставляет реализацию хранилища метрик на PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий PostgreSQL и инициализирует пул соединений.
func NewRepository(ctx context.Context, cfg *config.PostgresConfig) (*Repository, error) {
	if err := runMigrations(ctx, cfg.Dsn); err != nil {
		return nil, fmt.Errorf("apply migrations: %w", err)
	}

	if pool, err := pgxpool.New(ctx, cfg.Dsn); err == nil {
		return &Repository{pool}, nil
	} else {
		return nil, fmt.Errorf("%w", err)
	}
}

func runMigrations(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}
