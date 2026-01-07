package postgres

import (
	"context"
	"fmt"
)

// Ping проверяет доступность соединения с базой данных.
func (r *Repository) Ping(ctx context.Context) error {
	var result int
	if err := r.client.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
