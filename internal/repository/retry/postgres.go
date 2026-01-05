package retry

import (
	"context"
	"errors"

	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/arvaliullin/metrics-collection-service/internal/repository"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// ErrPostgresRepositoryUndefined сообщает о том, что репозиторий PostgreSQL не определён.
var ErrPostgresRepositoryUndefined = errors.New("postgres repository is undefined")

// PostgresAdapter добавляет стратегию повторов поверх репозитория PostgreSQL.
type PostgresAdapter struct {
	repo     repository.MetricStorage
	strategy *retryutil.Strategy
}

// NewPostgresAdapter создаёт адаптер репозитория PostgreSQL.
func NewPostgresAdapter(repo repository.MetricStorage, strategy *retryutil.Strategy) (*PostgresAdapter, error) {
	if repo == nil {
		return nil, ErrPostgresRepositoryUndefined
	}

	if strategy == nil {
		strategy = retryutil.NewStrategy(retryutil.DefaultDelays, IsConnectionRetryable)
	}

	return &PostgresAdapter{
		repo:     repo,
		strategy: strategy,
	}, nil
}

// UpdateGauge устанавливает значение gauge.
func (a *PostgresAdapter) UpdateGauge(ctx context.Context, id string, newGauge float64) {
	a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		a.repo.UpdateGauge(ctx, id, newGauge)
		return nil
	})
}

// GetGauge возвращает значение gauge.
func (a *PostgresAdapter) GetGauge(ctx context.Context, id string) (float64, error) {
	var value float64
	err := a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		var err error
		value, err = a.repo.GetGauge(ctx, id)
		return err
	})
	return value, err
}

// GetCounter возвращает значение counter.
func (a *PostgresAdapter) GetCounter(ctx context.Context, id string) (int64, error) {
	var value int64
	err := a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		var err error
		value, err = a.repo.GetCounter(ctx, id)
		return err
	})
	return value, err
}

// AddCounter увеличивает значение counter.
func (a *PostgresAdapter) AddCounter(ctx context.Context, id string, newCounter int64) {
	a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		a.repo.AddCounter(ctx, id, newCounter)
		return nil
	})
}

// AddCounterValue увеличивает значение counter на delta.
func (a *PostgresAdapter) AddCounterValue(ctx context.Context, id string, delta int64) {
	a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		a.repo.AddCounterValue(ctx, id, delta)
		return nil
	})
}

// ResetCounter сбрасывает значение счётчика.
func (a *PostgresAdapter) ResetCounter(ctx context.Context, id string) {
	a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		a.repo.ResetCounter(ctx, id)
		return nil
	})
}

// BatchUpdate выполняет пакетное обновление метрик.
func (a *PostgresAdapter) BatchUpdate(ctx context.Context, metrics []models.Metrics) error {
	return a.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		return a.repo.BatchUpdate(ctx, metrics)
	})
}

// AllCounters возвращает список counter метрик.
func (a *PostgresAdapter) AllCounters(ctx context.Context) []models.Metrics {
	return a.repo.AllCounters(ctx)
}

// AllGauges возвращает список gauge метрик.
func (a *PostgresAdapter) AllGauges(ctx context.Context) []models.Metrics {
	return a.repo.AllGauges(ctx)
}

// Ping проверяет соединение с базой данных.
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
