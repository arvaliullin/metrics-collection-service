package middleware

import (
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestTrustedSubnetMiddleware(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		trustedSubnet  string
		method         string
		path           string
		xRealIP        string
		expectedStatus int
	}{
		{
			name:           "empty trusted_subnet allows all",
			trustedSubnet:  "",
			method:         http.MethodPost,
			path:           "/updates",
			xRealIP:        "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP in subnet allows",
			trustedSubnet:  "192.168.1.0/24",
			method:         http.MethodPost,
			path:           "/updates",
			xRealIP:        "192.168.1.10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "IP outside subnet returns 403",
			trustedSubnet:  "192.168.1.0/24",
			method:         http.MethodPost,
			path:           "/updates",
			xRealIP:        "10.0.0.1",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "missing X-Real-IP returns 403",
			trustedSubnet:  "192.168.1.0/24",
			method:         http.MethodPost,
			path:           "/updates",
			xRealIP:        "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid X-Real-IP returns 403",
			trustedSubnet:  "192.168.1.0/24",
			method:         http.MethodPost,
			path:           "/updates",
			xRealIP:        "not-an-ip",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "GET request not checked",
			trustedSubnet:  "192.168.1.0/24",
			method:         http.MethodGet,
			path:           "/",
			xRealIP:        "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST /value not checked",
			trustedSubnet:  "192.168.1.0/24",
			method:         http.MethodPost,
			path:           "/value",
			xRealIP:        "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST /update path style checked",
			trustedSubnet:  "127.0.0.0/8",
			method:         http.MethodPost,
			path:           "/update/gauge/foo/1",
			xRealIP:        "127.0.0.1",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var trustedNet *net.IPNet
			if tt.trustedSubnet != "" {
				_, parsed, err := net.ParseCIDR(tt.trustedSubnet)
				if err != nil {
					t.Fatalf("invalid test CIDR %q: %v", tt.trustedSubnet, err)
				}
				trustedNet = parsed
			}
			mw := TrustedSubnetMiddleware(trustedNet, logger)
			wrapped := mw(handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			rr := httptest.NewRecorder()
			wrapped.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}
