package updates_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/updates"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	repomock "github.com/arvaliullin/metrics-collection-service/internal/repository/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestUpdatesHandler_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := repomock.NewMockMetricStorage(ctrl)
	handler := updates.NewUpdatesHandler(storage, nil)

	req := httptest.NewRequest(http.MethodGet, "/updates", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestUpdatesHandler_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := repomock.NewMockMetricStorage(ctrl)
	handler := updates.NewUpdatesHandler(storage, nil)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatesHandler_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := repomock.NewMockMetricStorage(ctrl)
	handler := updates.NewUpdatesHandler(storage, nil)

	payload := []models.Metrics{
		{ID: "Alloc", MType: models.Gauge, Value: floatPtr(1.23)},
		{ID: "PollCount", MType: models.Counter, Delta: intPtr(5)},
	}

	storage.EXPECT().
		BatchUpdate(gomock.Any(), gomock.Len(len(payload))).
		DoAndReturn(func(ctx context.Context, metrics []models.Metrics) error {
			require.Len(t, metrics, len(payload))
			require.Equal(t, payload[0].ID, metrics[0].ID)
			require.Equal(t, payload[1].ID, metrics[1].ID)
			return nil
		})

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, string(body), w.Body.String())
}

func floatPtr(v float64) *float64 {
	return &v
}

func intPtr(v int64) *int64 {
	return &v
}
