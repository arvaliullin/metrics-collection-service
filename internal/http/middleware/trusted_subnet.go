package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

// TrustedSubnetMiddleware проверяет, что IP из заголовка X-Real-IP входит в доверенную подсеть.
func TrustedSubnetMiddleware(trustedSubnet string, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if trustedSubnet == "" {
				next.ServeHTTP(w, r)
				return
			}
			if r.Method != http.MethodPost {
				next.ServeHTTP(w, r)
				return
			}
			path := r.URL.Path
			if path != "/update" && path != "/update/" && path != "/updates" && path != "/updates/" && !strings.HasPrefix(path, "/update/") {
				next.ServeHTTP(w, r)
				return
			}
			clientIPStr := strings.TrimSpace(r.Header.Get("X-Real-IP"))
			ip := net.ParseIP(clientIPStr)
			if ip == nil {
				logger.Warn().Str("x_real_ip", clientIPStr).Msg("missing or invalid X-Real-IP")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			_, ipNet, err := net.ParseCIDR(trustedSubnet)
			if err != nil {
				logger.Error().Err(err).Str("trusted_subnet", trustedSubnet).Msg("invalid trusted subnet CIDR")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			if !ipNet.Contains(ip) {
				logger.Warn().Str("ip", clientIPStr).Str("trusted_subnet", trustedSubnet).Msg("IP not in trusted subnet")
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
