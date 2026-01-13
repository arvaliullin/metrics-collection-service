package http_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	agenthttp "github.com/arvaliullin/metrics-collection-service/internal/http"
	agenthttpmock "github.com/arvaliullin/metrics-collection-service/internal/http/mock"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"go.uber.org/mock/gomock"
)

func TestHTTPMetricsSender_Send_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := agenthttpmock.NewMockHTTPClient(ctrl)
	mockResponse := agenthttpmock.NewMockResponse(ctrl)

	metrics := []models.Metrics{
		{ID: "gauge1", MType: models.Gauge, Value: floatPtr(1.0)},
		{ID: "counter1", MType: models.Counter, Delta: intPtr(1)},
	}

	mockHTTPClient.EXPECT().
		Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, url string, body []byte, headers map[string]string) (agenthttp.Response, error) {
			if url != "http://localhost:8080/updates" {
				t.Errorf("expected URL 'http://localhost:8080/updates' but got '%s'", url)
			}
			if headers["Content-Type"] != "application/json" {
				t.Errorf("expected Content-Type 'application/json' but got '%s'", headers["Content-Type"])
			}
			if headers["Content-Encoding"] != "gzip" {
				t.Errorf("expected Content-Encoding 'gzip' but got '%s'", headers["Content-Encoding"])
			}
			return mockResponse, nil
		})

	mockResponse.EXPECT().
		StatusCode().
		Return(http.StatusOK)

	sender := agenthttp.NewHTTPMetricsSender(mockHTTPClient, "http://localhost:8080", "")

	err := sender.Send(context.Background(), metrics)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHTTPMetricsSender_Send_WithHash(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := agenthttpmock.NewMockHTTPClient(ctrl)
	mockResponse := agenthttpmock.NewMockResponse(ctrl)

	metrics := []models.Metrics{
		{ID: "gauge1", MType: models.Gauge, Value: floatPtr(1.0)},
	}

	mockHTTPClient.EXPECT().
		Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, url string, body []byte, headers map[string]string) (agenthttp.Response, error) {
			if _, ok := headers["HashSHA256"]; !ok {
				t.Error("expected HashSHA256 header but it's missing")
			}
			return mockResponse, nil
		})

	mockResponse.EXPECT().
		StatusCode().
		Return(http.StatusOK)

	sender := agenthttp.NewHTTPMetricsSender(mockHTTPClient, "http://localhost:8080", "test-key")

	err := sender.Send(context.Background(), metrics)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHTTPMetricsSender_Send_NonOKStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := agenthttpmock.NewMockHTTPClient(ctrl)
	mockResponse := agenthttpmock.NewMockResponse(ctrl)

	metrics := []models.Metrics{
		{ID: "gauge1", MType: models.Gauge, Value: floatPtr(1.0)},
	}

	mockHTTPClient.EXPECT().
		Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(mockResponse, nil)

	mockResponse.EXPECT().
		StatusCode().
		Return(http.StatusInternalServerError).
		AnyTimes()

	sender := agenthttp.NewHTTPMetricsSender(mockHTTPClient, "http://localhost:8080", "")

	err := sender.Send(context.Background(), metrics)
	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestHTTPMetricsSender_Send_HTTPError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := agenthttpmock.NewMockHTTPClient(ctrl)

	metrics := []models.Metrics{
		{ID: "gauge1", MType: models.Gauge, Value: floatPtr(1.0)},
	}

	mockHTTPClient.EXPECT().
		Post(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("network error"))

	sender := agenthttp.NewHTTPMetricsSender(mockHTTPClient, "http://localhost:8080", "")

	err := sender.Send(context.Background(), metrics)
	if err == nil {
		t.Error("expected error but got none")
	}
}

func TestHTTPMetricsSender_Send_EmptyMetrics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHTTPClient := agenthttpmock.NewMockHTTPClient(ctrl)

	sender := agenthttp.NewHTTPMetricsSender(mockHTTPClient, "http://localhost:8080", "")

	err := sender.Send(context.Background(), nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int64) *int64 {
	return &i
}
