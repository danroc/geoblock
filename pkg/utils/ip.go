package utils

import (
	"bytes"
	"fmt"
	"net"
)

// CompareIP compares two IP addresses. It returns 0 if a == b, -1 if a < b,
// and 1 if a > b.
func CompareIP(a net.IP, b net.IP) int {
	return bytes.Compare(a, b)
}

// ErrInvalidIP is used when a give IP address is invalid.
type ErrInvalidIP struct {
	Address string
}

// Error returns the error message.
func (e *ErrInvalidIP) Error() string {
	return fmt.Sprintf("invalid IP address: %s", e.Address)
}

// IsIPv4 returns true if the given IP address is an IPv4 address. Otherwise,
// it returns false.
func IsIPv4(ip net.IP) bool {
	return ip.To4() != nil
}
