package utils

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

// GetOutboundIP возвращает исходящий IP хоста при подключении к serverAddr.
func GetOutboundIP(serverAddr string) (string, error) {
	hostPort := serverAddr
	if u, err := url.Parse(serverAddr); err == nil && u.Host != "" {
		hostPort = u.Host
	}
	conn, err := net.Dial("tcp", hostPort)
	if err != nil {
		return "", err
	}
	defer conn.Close()
	addr := conn.LocalAddr()
	if tcpAddr, ok := addr.(*net.TCPAddr); ok {
		return tcpAddr.IP.String(), nil
	}
	return strings.Split(addr.String(), ":")[0], nil
}

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
