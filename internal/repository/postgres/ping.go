package postgres

import (
	"context"
	"errors"
	"fmt"
)

// ErrPoolUndefined сообщает о том, что пул соединений не определён.
var ErrPoolUndefined = errors.New("pool undefined")

// Ping проверяет доступность соединения с базой данных.
func (r *Repository) Ping(ctx context.Context) error {
	if r.pool == nil {
		return ErrPoolUndefined
	}

	if err := r.pool.Ping(ctx); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
