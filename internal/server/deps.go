package server

import "context"

type PostgresRepository interface {
	Ping(ctx context.Context) error
}
