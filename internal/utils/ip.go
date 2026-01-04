package utils

import (
	"net"
	"net/http"
	"strings"
)

// ExtractIPAddress извлекает IP-адрес из HTTP-запроса.
func ExtractIPAddress(r *http.Request) string {
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	ip = strings.Trim(ip, "[]")
	if net.ParseIP(ip) != nil {
		return ip
	}
	return "unknown"
}
