package postgres

import (
	"context"
	"fmt"
)

// Ping проверяет доступность соединения с базой данных.
func (repo *Repository) Ping(ctx context.Context) error {
	if repo.pool == nil {
		return fmt.Errorf("pool undefined")
	}

	if err := repo.pool.Ping(ctx); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
