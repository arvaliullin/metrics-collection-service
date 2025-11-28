package server

import (
	"bytes"
	"crypto/hmac"
	"fmt"
	"io"
	"net/http"

	"github.com/arvaliullin/metrics-collection-service/internal/utils"
	"github.com/rs/zerolog"
)

// HashValidationMiddleware проверяет хеш входящих запросов
func HashValidationMiddleware(key string, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if r.Body == nil || r.ContentLength == 0 {
				next.ServeHTTP(w, r)
				return
			}

			receivedHash := r.Header.Get("HashSHA256")
			if receivedHash == "" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Error().Err(err).Msg("failed to read request body")
				http.Error(w, "failed to read request body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()

			computedHash, err := utils.Hash(body, key)
			if err != nil {
				logger.Error().Err(err).Msg("failed to compute hash")
				http.Error(w, "failed to compute hash", http.StatusInternalServerError)
				return
			}

			computedHashHex := fmt.Sprintf("%x", computedHash)

			if !hmac.Equal([]byte(receivedHash), []byte(computedHashHex)) {
				logger.Warn().
					Str("received", receivedHash).
					Str("computed", computedHashHex).
					Msg("hash mismatch")
				http.Error(w, "hash mismatch", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(body))

			next.ServeHTTP(w, r)
		})
	}
}
