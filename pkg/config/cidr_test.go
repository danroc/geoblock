// FILE: pkg/config/cidr_test.go
package config

import (
	"bytes"
	"net"
	"testing"

	"gopkg.in/yaml.v3"
)

func equalCIDR(a, b *net.IPNet) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

func TestUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *net.IPNet
		wantErr bool
	}{
		{
			name:  "valid CIDR",
			input: "192.168.1.0/24",
			want: &net.IPNet{
				IP:   net.IPv4(192, 168, 1, 0),
				Mask: net.CIDRMask(24, 32),
			},
			wantErr: false,
		},
		{
			name:    "invalid CIDR",
			input:   "invalid-cidr",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty CIDR",
			input:   "",
			want:    nil,
			wantErr: false, // The variable is left uninitialized
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cidr CIDR
			err := yaml.Unmarshal([]byte(tt.input), &cidr)
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"UnmarshalYAML() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
				return
			}
			if !equalCIDR(cidr.IPNet, tt.want) {
				t.Errorf(
					"UnmarshalYAML() got = %v, want %v",
					cidr.IPNet,
					tt.want,
				)
			}
		})
	}
}
