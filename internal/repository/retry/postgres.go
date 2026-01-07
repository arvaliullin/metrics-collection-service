package retry

import (
	"context"

	"github.com/arvaliullin/metrics-collection-service/internal/repository/postgres"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresRetryClient добавляет стратегию повторов для операций с PostgreSQL.
type PostgresRetryClient struct {
	pool     *pgxpool.Pool
	strategy *retryutil.Strategy
}

// NewPostgresRetryClient создаёт клиент PostgreSQL с поддержкой retry.
func NewPostgresRetryClient(pool *pgxpool.Pool, strategy *retryutil.Strategy) *PostgresRetryClient {
	return &PostgresRetryClient{
		pool:     pool,
		strategy: strategy,
	}
}

// QueryRow выполняет запрос, возвращающий одну строку, с применением стратегии повторов.
func (c *PostgresRetryClient) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	var row pgx.Row
	c.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		row = c.pool.QueryRow(ctx, sql, args...)
		return nil
	})
	return row
}

// Exec выполняет команду SQL без возврата строк с применением стратегии повторов.
func (c *PostgresRetryClient) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	var result pgconn.CommandTag
	var err error
	retryErr := c.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		result, err = c.pool.Exec(ctx, sql, arguments...)
		return err
	})
	if retryErr != nil {
		return result, retryErr
	}
	return result, err
}

// Query выполняет запрос, возвращающий множество строк, с применением стратегии повторов.
func (c *PostgresRetryClient) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	var rows pgx.Rows
	var err error
	retryErr := c.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		rows, err = c.pool.Query(ctx, sql, args...)
		return err
	})
	if retryErr != nil {
		return nil, retryErr
	}
	return rows, err
}

// Begin начинает транзакцию с применением стратегии повторов.
func (c *PostgresRetryClient) Begin(ctx context.Context) (pgx.Tx, error) {
	var tx pgx.Tx
	var err error
	retryErr := c.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		tx, err = c.pool.Begin(ctx)
		return err
	})
	if retryErr != nil {
		return nil, retryErr
	}
	return tx, err
}

var _ postgres.PostgresClient = (*PostgresRetryClient)(nil)
