package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var connectionErrorCodes = map[string]struct{}{
	pgerrcode.ConnectionException:                           {},
	pgerrcode.ConnectionDoesNotExist:                        {},
	pgerrcode.ConnectionFailure:                             {},
	pgerrcode.SQLClientUnableToEstablishSQLConnection:       {},
	pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection: {},
	pgerrcode.TransactionResolutionUnknown:                  {},
	pgerrcode.ProtocolViolation:                             {},
}

// IsConnectionRetryable определяет, относится ли ошибка к классу сбоев соединения PostgreSQL.
func IsConnectionRetryable(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		_, ok := connectionErrorCodes[pgErr.Code]
		return ok
	}

	var connectErr *pgconn.ConnectError
	return errors.As(err, &connectErr)
}

// NewPool создаёт пул соединений PostgreSQL.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, dsn)
}
