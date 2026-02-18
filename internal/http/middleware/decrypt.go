package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"
	"sync"

	"github.com/arvaliullin/metrics-collection-service/internal/crypto"
	"github.com/rs/zerolog"
)

const contentEncryptedHeader = "X-Content-Encrypted"

// DecryptMiddleware расшифровывает тело запроса при наличии заголовка X-Content-Encrypted и пути к приватному ключу.
func DecryptMiddleware(cryptoKeyPath string, logger zerolog.Logger) func(http.Handler) http.Handler {
	var (
		privateKey    *rsa.PrivateKey
		privateKeyErr error
		keyOnce       sync.Once
	)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cryptoKeyPath == "" || r.Header.Get(contentEncryptedHeader) != "true" {
				next.ServeHTTP(w, r)
				return
			}
			if r.Body == nil {
				next.ServeHTTP(w, r)
				return
			}
			keyOnce.Do(func() {
				privateKey, privateKeyErr = crypto.LoadPrivateKey(cryptoKeyPath)
			})
			if privateKeyErr != nil {
				logger.Error().Err(privateKeyErr).Msg("decrypt: load private key")
				http.Error(w, "decrypt failed", http.StatusInternalServerError)
				return
			}
			body, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				logger.Error().Err(err).Msg("decrypt: read body")
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			decrypted, err := crypto.Decrypt(privateKey, body)
			if err != nil {
				logger.Error().Err(err).Msg("decrypt: decrypt body")
				http.Error(w, "decrypt failed", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			next.ServeHTTP(w, r)
		})
	}
}
