package update_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/update"
	repomock "github.com/arvaliullin/metrics-collection-service/internal/repository/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code        int
		body        string
		contentType string
	}

	tests := []struct {
		name   string
		method string
		target string
		setup  func(*repomock.MockMetricStorage)
		want   want
	}{
		{
			name:   "test #1: valid counter",
			method: http.MethodPost,
			target: "/update/counter/someMetric/527",
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().AddCounter(gomock.Any(), "someMetric", int64(527))
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name:   "test #2: invalid method (GET)",
			method: http.MethodGet,
			target: "/update/counter/someMetric/527",
			want: want{
				code:        http.StatusMethodNotAllowed,
				body:        update.ErrMethodNotSupported.Error(),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "test #3: valid gauge",
			method: http.MethodPost,
			target: "/update/gauge/gauage_1/3.14",
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().UpdateGauge(gomock.Any(), "gauage_1", 3.14)
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
			},
		},
		{
			name:   "test #4: invalid counter value",
			method: http.MethodPost,
			target: "/update/counter/someMetric/asdsdasd",
			want: want{
				code:        http.StatusBadRequest,
				body:        update.ErrInvalidMetricValue.Error(),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "test #5: invalid gauge value",
			method: http.MethodPost,
			target: "/update/gauge/someMetric/asdsdasd",
			want: want{
				code:        http.StatusBadRequest,
				body:        update.ErrInvalidMetricValue.Error(),
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "test #6: invalid metric type",
			method: http.MethodPost,
			target: "/update/undefined/someMetric/asdsdasd",
			want: want{
				code:        http.StatusBadRequest,
				body:        update.ErrInvalidMetricType.Error(),
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.target, nil)
			w := httptest.NewRecorder()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storage := repomock.NewMockMetricStorage(ctrl)
			if tt.setup != nil {
				tt.setup(storage)
			}

			handler := update.NewUpdateHandler(storage, nil)
			mux := http.NewServeMux()

			mux.Handle(`/update/{type}/{id}/{value}`, handler)
			mux.ServeHTTP(w, r)

			assert.Equal(t, tt.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
			if tt.want.body != "" {
				assert.Contains(t, w.Body.String(), tt.want.body)
			}
			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))
			}
		})
	}
}
