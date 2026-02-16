package config

import (
	"net/netip"
	"testing"

	"github.com/goccy/go-yaml"
)

func TestCIDR_UnmarshalYAML(t *testing.T) {
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
			name:    "non-string value",
			input:   "[1, 2]",
			want:    netip.Prefix{},
			wantErr: true,
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
			if cidr.Prefix != tt.want {
				t.Errorf(
					"UnmarshalYAML() got = %v, want %v",
					cidr.Prefix,
					tt.want,
				)
			}
		})
	}
}
