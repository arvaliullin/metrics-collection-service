package server

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/arvaliullin/metrics-collection-service/internal/utils"
	"github.com/rs/zerolog"
)

type responseWriterWithBody struct {
	http.ResponseWriter
	body          *bytes.Buffer
	statusCode    int
	headerWritten bool
	key           string
	logger        zerolog.Logger
}

func (rw *responseWriterWithBody) Write(b []byte) (int, error) {
	if !rw.headerWritten {
		rw.WriteHeader(http.StatusOK)
	}
	rw.body.Write(b)
	return len(b), nil
}

func (rw *responseWriterWithBody) WriteHeader(statusCode int) {
	if rw.headerWritten {
		return
	}
	rw.statusCode = statusCode
	rw.headerWritten = true
}

func (rw *responseWriterWithBody) flush() {
	if rw.body.Len() > 0 && rw.statusCode == http.StatusOK && rw.key != "" {
		responseHash, err := utils.Hash(rw.body.Bytes(), rw.key)
		if err != nil {
			rw.logger.Error().Err(err).Msg("failed to compute response hash")
		} else {
			rw.ResponseWriter.Header().Set("HashSHA256", fmt.Sprintf("%x", responseHash))
		}
	}

	rw.ResponseWriter.WriteHeader(rw.statusCode)
	rw.ResponseWriter.Write(rw.body.Bytes())
}

// HashResponseMiddleware добавляет хеш к ответам
func HashResponseMiddleware(key string, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			wrapped := &responseWriterWithBody{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
				headerWritten:  false,
				key:            key,
				logger:         logger,
			}

			next.ServeHTTP(wrapped, r)

			wrapped.flush()
		})
	}
}
