package postgres

import (
	"context"
	"fmt"
)

// Ping проверяет доступность соединения с базой данных.
func (r *Repository) Ping(ctx context.Context) error {
	if r.pool == nil {
		return fmt.Errorf("pool undefined")
	}

	if err := r.pool.Ping(ctx); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
