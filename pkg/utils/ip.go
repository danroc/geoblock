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

// InvalidIPError is used when a give IP address is invalid.
type InvalidIPError struct {
	Address string
}

// Error returns the error message.
func (e *InvalidIPError) Error() string {
	return fmt.Sprintf("invalid IP address: %s", e.Address)
}

func IsIPv4(ip net.IP) bool {
	return ip.To4() != nil
}
