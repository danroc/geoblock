package iputils_test

import (
	"net"
	"testing"

	"github.com/danroc/geoblock/internal/utils/iputils"
)

func TestCompareIP(t *testing.T) {
	tests := []struct {
		a, b   string
		result int
	}{
		{"192.168.1.1", "192.168.1.1", 0},
		{"192.168.1.1", "192.168.1.2", -1},
		{"192.168.1.2", "192.168.1.1", 1},
		{"::1", "::1", 0},
		{"::1", "::2", -1},
		{"::2", "::1", 1},
	}

	for _, test := range tests {
		ipA := net.ParseIP(test.a)
		ipB := net.ParseIP(test.b)
		if ipA == nil || ipB == nil {
			t.Fatalf("Invalid IP address in test case: %s, %s", test.a, test.b)
		}

		result := iputils.CompareIP(ipA, ipB)
		if result != test.result {
			t.Errorf(
				"CompareIP(%s, %s) = %d; want %d",
				test.a,
				test.b,
				result,
				test.result,
			)
		}
	}
}

func TestIsIPv4(t *testing.T) {
	tests := []struct {
		ip     string
		result bool
	}{
		{"192.168.1.1", true},
		{"255.255.255.255", true},
		{"0.0.0.0", true},
		{"::1", false},
		{"2001:db8::68", false},
		{"", false},
	}

	for _, test := range tests {
		ip := net.ParseIP(test.ip)
		if ip == nil && test.ip != "" {
			t.Fatalf("Invalid IP address in test case: %s", test.ip)
		}

		result := iputils.IsIPv4(ip)
		if result != test.result {
			t.Errorf(
				"IsIPv4(%s) = %t; want %t",
				test.ip,
				result,
				test.result,
			)
		}
	}
}

func TestErrInvalidIP(t *testing.T) {
	tests := []struct {
		address string
		message string
	}{
		{"256.256.256.256", "invalid IP address: 256.256.256.256"},
		{"invalid-ip", "invalid IP address: invalid-ip"},
		{"", "invalid IP address: "},
	}

	for _, test := range tests {
		err := &iputils.ErrInvalidIP{Address: test.address}
		if err.Error() != test.message {
			t.Errorf(
				"ErrInvalidIP(%s).Error() = %s; want %s",
				test.address,
				err.Error(),
				test.message,
			)
		}
	}
}
