package http

import (
	"context"
)

//go:generate mockgen -source=client.go -destination=mock/client_mock.go -package=httpmock

// Response представляет ответ HTTP запроса.
type Response interface {
	StatusCode() int
}

// HTTPClient определяет минимально необходимые операции для работы с HTTP клиентом.
type HTTPClient interface {
	Post(ctx context.Context, url string, body []byte, headers map[string]string) (Response, error)
}
