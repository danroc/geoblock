package server

import (
	"net/netip"
	"testing"
)

func TestIsLocalIP(t *testing.T) {
	tests := []struct {
		ip      string
		isLocal bool
	}{
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"127.0.0.1", true},
		{"169.254.1.1", true},
		{"::1", true},
		{"fc00::1", true},
		{"fe80::1", true},
		{"8.8.8.8", false},
		{"2001:4860:4860::8888", false},
	}

	for _, test := range tests {
		ip, err := netip.ParseAddr(test.ip)
		if err != nil {
			t.Fatalf("Failed to parse IP %s: %v", test.ip, err)
		}
		if isLocalIP(ip) != test.isLocal {
			t.Errorf("isLocalIP(%s) = %v, want %v", test.ip, isLocalIP(ip), test.isLocal)
		}
	}
}
