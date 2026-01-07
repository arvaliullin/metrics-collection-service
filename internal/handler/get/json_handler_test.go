package get_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/get"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	repomock "github.com/arvaliullin/metrics-collection-service/internal/repository/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetJSONHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code        int
		body        string
		contentType string
		metric      *models.Metrics
	}

	tests := []struct {
		name        string
		method      string
		target      string
		body        models.Metrics
		setup       func(*repomock.MockMetricStorage)
		want        want
		invalidJSON bool
	}{
		{
			name:   "test #1: valid gauge retrieval",
			method: http.MethodPost,
			target: "/value",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
			},
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetGauge(gomock.Any(), "someMetric").Return(3.14, nil)
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				metric: &models.Metrics{
					ID:    "someMetric",
					MType: "gauge",
					Value: func() *float64 { v := 3.14; return &v }(),
				},
			},
		},
		{
			name:   "test #2: valid counter retrieval",
			method: http.MethodPost,
			target: "/value",
			body: models.Metrics{
				ID:    "counterMetric",
				MType: "counter",
			},
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetCounter(gomock.Any(), "counterMetric").Return(int64(42), nil)
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				metric: &models.Metrics{
					ID:    "counterMetric",
					MType: "counter",
					Delta: func() *int64 { v := int64(42); return &v }(),
				},
			},
		},
		{
			name:   "test #3: invalid method (GET)",
			method: http.MethodGet,
			target: "/value",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
			},
			want: want{
				code: http.StatusMethodNotAllowed,
				body: get.ErrMethodNotAllowed.Error(),
			},
		},
		{
			name:   "test #4: missing ID",
			method: http.MethodPost,
			target: "/value",
			body: models.Metrics{
				MType: "gauge",
			},
			want: want{
				code: http.StatusBadRequest,
				body: get.ErrMissingID.Error(),
			},
		},
		{
			name:   "test #5: missing type",
			method: http.MethodPost,
			target: "/value",
			body: models.Metrics{
				ID: "someMetric",
			},
			want: want{
				code: http.StatusBadRequest,
				body: get.ErrMissingType.Error(),
			},
		},
		{
			name:   "test #6: gauge not found",
			method: http.MethodPost,
			target: "/value",
			body: models.Metrics{
				ID:    "nonexistent",
				MType: "gauge",
			},
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetGauge(gomock.Any(), "nonexistent").Return(0.0, get.ErrNotFound)
			},
			want: want{
				code: http.StatusNotFound,
				body: get.ErrNotFound.Error(),
			},
		},
		{
			name:   "test #7: counter not found",
			method: http.MethodPost,
			target: "/value",
			body: models.Metrics{
				ID:    "nonexistent",
				MType: "counter",
			},
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetCounter(gomock.Any(), "nonexistent").Return(int64(0), get.ErrNotFound)
			},
			want: want{
				code: http.StatusNotFound,
				body: get.ErrNotFound.Error(),
			},
		},
		{
			name:   "test #8: invalid metric type",
			method: http.MethodPost,
			target: "/value",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "invalid_type",
			},
			want: want{
				code: http.StatusBadRequest,
				body: get.ErrInvalidMetricType.Error(),
			},
		},
		{
			name:   "test #9: invalid JSON body",
			method: http.MethodPost,
			target: "/value",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
			},
			want: want{
				code: http.StatusBadRequest,
				body: get.ErrInvalidJSON.Error(),
			},
			invalidJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := repomock.NewMockMetricStorage(ctrl)
			if tt.setup != nil {
				tt.setup(storage)
			}

			var body *bytes.Reader
			if tt.invalidJSON {
				body = bytes.NewReader([]byte("invalid json }}}}}"))
			} else {
				bodyBytes, _ := json.Marshal(tt.body)
				body = bytes.NewReader(bodyBytes)
			}

			r := httptest.NewRequest(tt.method, tt.target, body)
			w := httptest.NewRecorder()

			handler := get.NewGetJSONHandler(storage)
			mux := http.NewServeMux()
			mux.Handle("/value", handler)
			mux.ServeHTTP(w, r)

			assert.Equal(t, tt.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
			if tt.want.body != "" {
				assert.Contains(t, w.Body.String(), tt.want.body)
			}
			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"), "Content-Type не совпадает с ожидаемым")
			}
			if tt.want.metric != nil {
				var response models.Metrics
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err, "Ошибка декодирования ответа")
				assert.Equal(t, tt.want.metric.ID, response.ID, "ID метрики не совпадает")
				assert.Equal(t, tt.want.metric.MType, response.MType, "Тип метрики не совпадает")
				if tt.want.metric.Value != nil {
					assert.NotNil(t, response.Value, "Значение Value не должно быть nil")
					assert.Equal(t, *tt.want.metric.Value, *response.Value, "Значение метрики не совпадает")
				}
				if tt.want.metric.Delta != nil {
					assert.NotNil(t, response.Delta, "Значение Delta не должно быть nil")
					assert.Equal(t, *tt.want.metric.Delta, *response.Delta, "Значение дельты не совпадает")
				}
			}
		})
	}
}
