package retry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	agenthttp "github.com/arvaliullin/metrics-collection-service/internal/http"
	retryutil "github.com/arvaliullin/metrics-collection-service/internal/utils/retry"
	"github.com/go-resty/resty/v2"
)

func TestHTTPRetryClient_Post(t *testing.T) {
	t.Run("retry on network error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		brokenURL := srv.URL
		srv.Close()

		client := resty.New()
		strategy := retryutil.NewStrategy(
			[]time.Duration{time.Millisecond},
			func(err error) bool {
				return err != nil
			},
		)
		httpClient := NewHTTPRetryClient(client, strategy)

		ctx := context.Background()
		body := []byte("test body")
		headers := map[string]string{"Content-Type": "application/json"}

		_, err := httpClient.Post(ctx, brokenURL, body, headers)

		if err == nil {
			t.Errorf("expected error but got none")
		}
	})
}

func TestRestyResponse_StatusCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()

	client := resty.New()
	resp, err := client.R().SetDoNotParseResponse(true).Post(srv.URL)
	if err != nil {
		t.Fatalf("unexpected post error: %v", err)
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	restyResp := &restyResponse{resp: resp}
	if code := restyResp.StatusCode(); code != http.StatusCreated {
		t.Errorf("StatusCode() = %d, want %d", code, http.StatusCreated)
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
