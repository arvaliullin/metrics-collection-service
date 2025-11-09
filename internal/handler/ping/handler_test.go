package ping_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arvaliullin/metrics-collection-service/internal/handler/ping"
	pingmock "github.com/arvaliullin/metrics-collection-service/internal/handler/ping/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPingHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		pingErr        error
		wantStatus     int
		wantBody       string
		wantHeaderType string
	}{
		{
			name:           "success",
			pingErr:        nil,
			wantStatus:     http.StatusOK,
			wantHeaderType: "text/plain",
		},
		{
			name:           "repository error",
			pingErr:        errors.New("database is unreachable"),
			wantStatus:     http.StatusInternalServerError,
			wantBody:       ping.ErrInternalServer.Error(),
			wantHeaderType: "text/plain; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := pingmock.NewMockPostgresRepository(ctrl)
			repo.EXPECT().Ping(gomock.Any()).Return(tt.pingErr)

			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			rr := httptest.NewRecorder()

			handler := ping.NewPingHandler(repo)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantHeaderType != "" {
				assert.Equal(t, tt.wantHeaderType, rr.Header().Get("Content-Type"))
			}
			if tt.wantBody != "" {
				assert.Contains(t, rr.Body.String(), tt.wantBody)
			} else {
				assert.Empty(t, rr.Body.String())
			}
		})
	}
}
