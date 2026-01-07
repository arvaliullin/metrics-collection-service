package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/arvaliullin/metrics-collection-service/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// Repository предоставляет реализацию хранилища метрик на PostgreSQL.
type Repository struct {
	client PostgresClient
}

// NewRepository создаёт репозиторий PostgreSQL.
func NewRepository(client PostgresClient) *Repository {
	return &Repository{client: client}
}

// RunMigrations применяет миграции к базе данных.
func RunMigrations(ctx context.Context, dsn string) error {
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
