// FILE: pkg/config/cidr_test.go
package config

import (
	"net/netip"
	"testing"

	"gopkg.in/yaml.v3"
)

func equalCIDR(a, b netip.Prefix) bool {
	return a.String() == b.String()
}

func TestUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    netip.Prefix
		wantErr bool
	}{
		{
			name:    "valid CIDR",
			input:   "192.168.1.0/24",
			want:    netip.MustParsePrefix("192.168.1.0/24"),
			wantErr: false,
		},
		{
			name:    "invalid CIDR",
			input:   "invalid-cidr",
			want:    netip.Prefix{},
			wantErr: true,
		},
		{
			name:    "empty CIDR",
			input:   "",
			want:    netip.Prefix{},
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
			if !equalCIDR(cidr.Prefix, tt.want) {
				t.Errorf(
					"UnmarshalYAML() got = %v, want %v",
					cidr.Prefix,
					tt.want,
				)
			}
		})
	}
}
