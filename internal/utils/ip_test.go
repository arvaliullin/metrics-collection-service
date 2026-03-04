package utils

import (
	"net"
	"testing"
)

func TestGetOutboundIP(t *testing.T) {

	const (
		listenAddr = "127.0.0.1:0"
		expectIP   = "127.0.0.1"
	)

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		t.Skipf("cannot listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()

	ip, err := GetOutboundIP(addr)
	if err != nil {
		t.Fatalf("GetOutboundIP(%q): %v", addr, err)
	}
	if ip != expectIP {
		t.Errorf("GetOutboundIP(%q) = %q, want %s", addr, ip, expectIP)
	}

	ip, err = GetOutboundIP("http://" + addr)
	if err != nil {
		t.Fatalf("GetOutboundIP with URL: %v", err)
	}
	if ip != expectIP {
		t.Errorf("GetOutboundIP(URL) = %q, want %s", ip, expectIP)
	}
}

func TestGetOutboundIP_InvalidAddress(t *testing.T) {
	_, err := GetOutboundIP("http://nonexistent.invalid:99999")
	if err == nil {
		t.Error("expected error for invalid address")
	}
}
