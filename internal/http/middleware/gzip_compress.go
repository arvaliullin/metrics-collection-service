package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gz           *gzip.Writer
	needCompress bool
	contentType  string
}

func newGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		ResponseWriter: w,
		gz:             nil,
		needCompress:   false,
		contentType:    "",
	}
}

func (gw *gzipResponseWriter) WriteHeader(code int) {
	gw.contentType = gw.Header().Get("Content-Type")

	gw.needCompress = strings.Contains(gw.contentType, "application/json") ||
		strings.Contains(gw.contentType, "text/html")

	if gw.needCompress {
		gw.Header().Set("Content-Encoding", "gzip")
		gw.gz = gzip.NewWriter(gw.ResponseWriter)
	}

	gw.ResponseWriter.WriteHeader(code)
}

func (gw *gzipResponseWriter) Write(b []byte) (int, error) {
	if gw.needCompress && gw.gz != nil {
		return gw.gz.Write(b)
	}
	return gw.ResponseWriter.Write(b)
}

func (gw *gzipResponseWriter) Close() error {
	if gw.gz != nil {
		return gw.gz.Close()
	}
	return nil
}

func GzipCompressMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
				next.ServeHTTP(w, r)
				return
			}

			gz := newGzipResponseWriter(w)
			defer gz.Close()

			next.ServeHTTP(gz, r)
		})
	}
}
