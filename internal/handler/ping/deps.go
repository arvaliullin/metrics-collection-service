package ping

import "context"

//go:generate mockgen -source=deps.go -destination=mock/repository_mock.go -package=pingmock
type Pinger interface {
	Ping(ctx context.Context) error
}
