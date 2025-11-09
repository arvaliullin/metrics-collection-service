package ping

import "context"

//go:generate mockgen -source=deps.go -destination=mock/repository_mock.go -package=pingmock
type PostgresRepository interface {
	Ping(ctx context.Context) error
}
