package get_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/get"
	repomock "github.com/arvaliullin/metrics-collection-service/internal/repository/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetHandler_ServeHTTP(t *testing.T) {
	type want struct {
		code        int
		body        string
		contentType string
	}

	tests := []struct {
		name  string
		path  string
		setup func(*repomock.MockMetricStorage)
		want  want
	}{
		{
			name: "test #1: valid gauge retrieval",
			path: "/gauge/someMetric",
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetGauge(gomock.Any(), "someMetric").Return(3.14, nil)
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
				body:        "3.14",
			},
		},
		{
			name: "test #2: valid counter retrieval",
			path: "/counter/counterMetric",
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetCounter(gomock.Any(), "counterMetric").Return(int64(42), nil)
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
				body:        "42",
			},
		},
		{
			name: "test #3: gauge not found",
			path: "/gauge/nonexistent",
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetGauge(gomock.Any(), "nonexistent").Return(0.0, get.ErrNotFound)
			},
			want: want{
				code: http.StatusNotFound,
				body: get.ErrNotFound.Error(),
			},
		},
		{
			name: "test #4: counter not found",
			path: "/counter/nonexistent",
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetCounter(gomock.Any(), "nonexistent").Return(int64(0), get.ErrNotFound)
			},
			want: want{
				code: http.StatusNotFound,
				body: get.ErrNotFound.Error(),
			},
		},
		{
			name: "test #5: invalid metric type",
			path: "/invalid_type/someMetric",
			want: want{
				code: http.StatusBadRequest,
				body: get.ErrInvalidMetricType.Error(),
			},
		},
		{
			name: "test #6: gauge with zero value",
			path: "/gauge/zeroMetric",
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetGauge(gomock.Any(), "zeroMetric").Return(0.0, nil)
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
				body:        "0",
			},
		},
		{
			name: "test #7: counter with zero value",
			path: "/counter/zeroCounter",
			setup: func(ms *repomock.MockMetricStorage) {
				ms.EXPECT().GetCounter(gomock.Any(), "zeroCounter").Return(int64(0), nil)
			},
			want: want{
				code:        http.StatusOK,
				contentType: "text/plain",
				body:        "0",
			},
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

			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler := get.NewGetHandler(storage)
			mux := http.NewServeMux()
			mux.Handle("/{type}/{id}", handler)
			mux.ServeHTTP(w, r)

			assert.Equal(t, tt.want.code, w.Code, "Код ответа не совпадает с ожидаемым")
			if tt.want.body != "" {
				assert.Contains(t, w.Body.String(), tt.want.body, "Тело ответа не совпадает с ожидаемым")
			}
			if tt.want.contentType != "" {
				assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"), "Content-Type не совпадает с ожидаемым")
			}
		})
	}
}

func TestGetHandler_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := repomock.NewMockMetricStorage(ctrl)

	r := httptest.NewRequest(http.MethodGet, "/gauge", nil)
	w := httptest.NewRecorder()

	handler := get.NewGetHandler(storage)

	mux := http.NewServeMux()
	mux.HandleFunc("/gauge", func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("type", "gauge")
		r.SetPathValue("id", "")
		handler.ServeHTTP(w, r)
	})

	mux.ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), get.ErrNotFound.Error())
}
