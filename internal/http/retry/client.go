package retry

import (
	"context"

	agenthttp "github.com/arvaliullin/metrics-collection-service/internal/http"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/go-resty/resty/v2"
)

// HTTPRetryClient добавляет стратегию повторов для HTTP операций.
type HTTPRetryClient struct {
	client   *resty.Client
	strategy *retryutil.Strategy
}

// NewHTTPRetryClient создаёт HTTP клиент с поддержкой retry.
func NewHTTPRetryClient(client *resty.Client, strategy *retryutil.Strategy) *HTTPRetryClient {
	return &HTTPRetryClient{
		client:   client,
		strategy: strategy,
	}
}

// Post выполняет POST запрос с применением стратегии повторов.
func (c *HTTPRetryClient) Post(ctx context.Context, url string, body []byte, headers map[string]string) (agenthttp.Response, error) {
	var resp *resty.Response
	var err error
	retryErr := c.strategy.DoWithRetry(ctx, func(ctx context.Context) error {
		req := c.client.R().
			SetContext(ctx).
			SetBody(body)

		for key, value := range headers {
			req.SetHeader(key, value)
		}

		resp, err = req.Post(url)
		return err
	})
	if retryErr != nil {
		return nil, retryErr
	}
	return &restyResponse{resp: resp}, err
}

// restyResponse оборачивает resty.Response для реализации интерфейса Response.
type restyResponse struct {
	resp *resty.Response
}

// StatusCode возвращает HTTP статус код ответа.
func (r *restyResponse) StatusCode() int {
	return r.resp.StatusCode()
}

var _ agenthttp.HTTPClient = (*HTTPRetryClient)(nil)
