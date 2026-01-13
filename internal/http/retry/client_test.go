package retry

import (
	"context"
	"testing"
	"time"

	agenthttp "github.com/arvaliullin/metrics-collection-service/internal/http"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/go-resty/resty/v2"
)

func TestHTTPRetryClient_Post(t *testing.T) {
	t.Run("retry on network error", func(t *testing.T) {
		client := resty.New()
		strategy := retryutil.NewStrategy(
			[]time.Duration{time.Millisecond},
			func(err error) bool {
				return err != nil
			},
		)
		httpClient := NewHTTPRetryClient(client, strategy)

		ctx := context.Background()
		url := "http://invalid-url-for-test:8080/test"
		body := []byte("test body")
		headers := map[string]string{"Content-Type": "application/json"}

		_, err := httpClient.Post(ctx, url, body, headers)

		if err == nil {
			t.Errorf("expected error but got none")
		}
	})
}

func TestRestyResponse_StatusCode(t *testing.T) {
	client := resty.New()
	resp, err := client.R().SetDoNotParseResponse(true).Post("http://invalid-url-for-test")
	if err == nil && resp != nil {
		restyResp := &restyResponse{resp: resp}
		restyResp.StatusCode()
	}
}

func TestHTTPRetryClient_ImplementsInterface(t *testing.T) {
	client := resty.New()
	strategy := retryutil.NewStrategy(
		[]time.Duration{time.Millisecond},
		func(err error) bool { return false },
	)
	httpClient := NewHTTPRetryClient(client, strategy)

	var _ agenthttp.HTTPClient = httpClient
}
