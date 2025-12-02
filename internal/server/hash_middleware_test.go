package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/arvaliullin/metrics-collection-service/internal/utils"
	"github.com/rs/zerolog"
)

func TestHashValidationMiddleware(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	key := "test-key"
	testData := []byte(`{"test": "data"}`)

	validHash, _ := utils.Hash(testData, key)
	validHashHex := fmt.Sprintf("%x", validHash)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		key            string
		body           []byte
		hash           string
		expectedStatus int
	}{
		{
			name:           "valid hash",
			key:            key,
			body:           testData,
			hash:           validHashHex,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid hash",
			key:            key,
			body:           testData,
			hash:           "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "no hash header",
			key:            key,
			body:           testData,
			hash:           "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "no key configured",
			key:            "",
			body:           testData,
			hash:           "anything",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := HashValidationMiddleware(tt.key, logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader(tt.body))
			if tt.hash != "" {
				req.Header.Set("HashSHA256", tt.hash)
			}

			rr := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

func TestHashResponseMiddleware(t *testing.T) {
	logger := zerolog.New(os.Stdout)
	key := "test-key"
	responseData := []byte(`{"response": "data"}`)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(responseData)
	})

	tests := []struct {
		name       string
		key        string
		expectHash bool
		statusCode int
	}{
		{
			name:       "adds hash with key",
			key:        key,
			expectHash: true,
			statusCode: http.StatusOK,
		},
		{
			name:       "no hash without key",
			key:        "",
			expectHash: false,
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := HashResponseMiddleware(tt.key, logger)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.statusCode {
				t.Errorf("expected status %d, got %d", tt.statusCode, rr.Code)
			}

			hashHeader := rr.Header().Get("HashSHA256")
			if tt.expectHash && hashHeader == "" {
				t.Error("expected HashSHA256 header, but it was not set")
			}
			if !tt.expectHash && hashHeader != "" {
				t.Error("did not expect HashSHA256 header, but it was set")
			}

			if tt.expectHash {
				expectedHash, _ := utils.Hash(responseData, tt.key)
				expectedHashHex := fmt.Sprintf("%x", expectedHash)
				if hashHeader != expectedHashHex {
					t.Errorf("hash mismatch: expected %s, got %s", expectedHashHex, hashHeader)
				}
			}

			body, _ := io.ReadAll(rr.Body)
			if !bytes.Equal(body, responseData) {
				t.Errorf("response body mismatch: expected %s, got %s", responseData, body)
			}
		})
	}
}
