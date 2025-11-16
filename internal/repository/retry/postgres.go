package retry

import (
	"context"
	"errors"
	"fmt"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresRepository interface {
	UpdateGauge(ctx context.Context, id string, newGauge float64)
	GetGauge(ctx context.Context, id string) (float64, error)
	GetCounter(ctx context.Context, id string) (int64, error)
	AddCounter(ctx context.Context, id string, newCounter int64)
	AddCounterValue(ctx context.Context, id string, delta int64)
	BatchUpdate(ctx context.Context, metrics []models.Metrics) error
	AllCounters(ctx context.Context) []models.Metrics
	AllGauges(ctx context.Context) []models.Metrics
	Ping(ctx context.Context) error
}

// PostgresAdapter добавляет стратегию повторов поверх репозитория PostgreSQL.
type PostgresAdapter struct {
	repo     PostgresRepository
	strategy *retryutil.Strategy
}

// NewPostgresAdapter создаёт адаптер и настраивает стратегию повторов по умолчанию при необходимости.
func NewPostgresAdapter(repo PostgresRepository, strategy *retryutil.Strategy) (*PostgresAdapter, error) {
	if repo == nil {
		return nil, fmt.Errorf("postgres repository is undefined")
	}

	if strategy == nil {
		strategy = retryutil.NewStrategy(retryutil.DefaultDelays, IsConnectionRetryable)
	}

	return &PostgresAdapter{
		repo:     repo,
		strategy: strategy,
	}, nil
}

// UpdateGauge устанавливает значение gauge без повторов на уровне адаптера.
func (a *PostgresAdapter) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	a.repo.UpdateGauge(ctx, id, newGauge)
}

// GetGauge возвращает значение gauge с учётом повторов при сбоях соединения.
func (a *PostgresAdapter) GetGauge(ctx context.Context, id string) (float64, error) {
	var value float64
	err := a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		var err error
		value, err = a.repo.GetGauge(ctx, id)
		return err
	})
	return value, err
}

// GetCounter возвращает значение counter с повторными попытками при ошибках соединения.
func (a *PostgresAdapter) GetCounter(ctx context.Context, id string) (int64, error) {
	var value int64
	err := a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		var err error
		value, err = a.repo.GetCounter(ctx, id)
		return err
	})
	return value, err
}

// AddCounter увеличивает значение counter без повторов на уровне адаптера.
func (a *PostgresAdapter) AddCounter(ctx context.Context, id string, newCounter int64) {
	a.repo.AddCounter(ctx, id, newCounter)
}

// AddCounterValue добавляет значение delta без повторов на уровне адаптера.
func (a *PostgresAdapter) AddCounterValue(ctx context.Context, id string, delta int64) {
	a.repo.AddCounterValue(ctx, id, delta)
}

// BatchUpdate выполняет пакет обновлений с повторными попытками для Class 08 ошибок.
func (a *PostgresAdapter) BatchUpdate(ctx context.Context, metrics []models.Metrics) error {
	return a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		return a.repo.BatchUpdate(ctx, metrics)
	})
}

// AllCounters возвращает список counter метрик из исходного репозитория.
func (a *PostgresAdapter) AllCounters(ctx context.Context) []models.Metrics {
	return a.repo.AllCounters(ctx)
}

// AllGauges возвращает список gauge метрик из исходного репозитория.
func (a *PostgresAdapter) AllGauges(ctx context.Context) []models.Metrics {
	return a.repo.AllGauges(ctx)
}

// Ping выполняет проверку соединения с базой данных с повторными попытками.
func (a *PostgresAdapter) Ping(ctx context.Context) error {
	return a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		return a.repo.Ping(ctx)
	})
}

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
