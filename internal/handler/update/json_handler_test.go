package update_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	updatemock "github.com/arvaliullin/metrics-collection-service/internal/handler/update/mock"
	models "github.com/arvaliullin/metrics-collection-service/internal/model"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateJSONHandler_ServeHTTP(t *testing.T) {
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
		setup       func(*updatemock.MockMetricStorage)
		want        want
		invalidJSON bool
	}{
		{
			name:   "test #1: valid gauge",
			method: http.MethodPost,
			target: "/update",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
				Value: func() *float64 { v := 3.14; return &v }(),
			},
			setup: func(ms *updatemock.MockMetricStorage) {
				ms.EXPECT().UpdateGauge(gomock.Any(), "someMetric", 3.14)
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
			name:   "test #2: valid counter",
			method: http.MethodPost,
			target: "/update",
			body: models.Metrics{
				ID:    "counterMetric",
				MType: "counter",
				Delta: func() *int64 { v := int64(42); return &v }(),
			},
			setup: func(ms *updatemock.MockMetricStorage) {
				ms.EXPECT().AddCounterValue(gomock.Any(), "counterMetric", int64(42))
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
			target: "/update",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
				Value: func() *float64 { v := 3.14; return &v }(),
			},
			want: want{
				code: http.StatusMethodNotAllowed,
				body: update.ErrMethodNotSupported.Error(),
			},
		},
		{
			name:   "test #4: missing ID",
			method: http.MethodPost,
			target: "/update",
			body: models.Metrics{
				MType: "gauge",
				Value: func() *float64 { v := 3.14; return &v }(),
			},
			want: want{
				code: http.StatusBadRequest,
				body: update.ErrMissingID.Error(),
			},
		},
		{
			name:   "test #5: missing type",
			method: http.MethodPost,
			target: "/update",
			body: models.Metrics{
				ID:    "someMetric",
				Value: func() *float64 { v := 3.14; return &v }(),
			},
			want: want{
				code: http.StatusBadRequest,
				body: update.ErrMissingType.Error(),
			},
		},
		{
			name:   "test #6: both delta and value set",
			method: http.MethodPost,
			target: "/update",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
				Value: func() *float64 { v := 3.14; return &v }(),
				Delta: func() *int64 { v := int64(42); return &v }(),
			},
			want: want{
				code: http.StatusBadRequest,
				body: update.ErrBothFieldsSet.Error(),
			},
		},
		{
			name:   "test #7: neither delta nor value set",
			method: http.MethodPost,
			target: "/update",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
			},
			want: want{
				code: http.StatusBadRequest,
				body: update.ErrNoFieldsSet.Error(),
			},
		},
		{
			name:   "test #8: invalid metric type",
			method: http.MethodPost,
			target: "/update",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "invalid_type",
				Value: func() *float64 { v := 3.14; return &v }(),
			},
			want: want{
				code: http.StatusBadRequest,
				body: update.ErrInvalidMetricType.Error(),
			},
		},
		{
			name:   "test #9: gauge without value",
			method: http.MethodPost,
			target: "/update",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "gauge",
				Delta: func() *int64 { v := int64(1); return &v }(),
			},
			want: want{
				code: http.StatusBadRequest,
				body: update.ErrMissingGaugeValue.Error(),
			},
		},
		{
			name:   "test #10: counter without delta",
			method: http.MethodPost,
			target: "/update",
			body: models.Metrics{
				ID:    "someMetric",
				MType: "counter",
				Value: func() *float64 { v := 3.14; return &v }(),
			},
			want: want{
				code: http.StatusBadRequest,
				body: update.ErrMissingCounterDelta.Error(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reader *bytes.Reader
			if tt.invalidJSON {
				reader = bytes.NewReader([]byte("invalid json"))
			} else {
				bodyBytes, _ := json.Marshal(tt.body)
				reader = bytes.NewReader(bodyBytes)
			}

			r := httptest.NewRequest(tt.method, tt.target, reader)
			w := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := updatemock.NewMockMetricStorage(ctrl)
			if tt.setup != nil {
				tt.setup(storage)
			}

			handler := update.NewUpdateJSONHandler(storage)
			mux := http.NewServeMux()
			mux.Handle("/update", handler)
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
