package config_test

import (
	"errors"
	"net"
	"reflect"
	"strings"
	"testing"

	"github.com/danroc/geoblock/pkg/config"
)

const validConfig = `
access_control:
  default_policy: allow
  rules:
    - networks:
        - "10.0.0.0/8"
        - "127.0.0.0/8"
      domains:
        - "example.com"
        - "*.example.com"
      methods:
        - GET
        - POST
      countries:
        - US
        - FR
      autonomous_systems:
        - 1234
        - 5678
      policy: allow

    - policy: deny
`

const invalidLeadingDot = `
access_control:
  default_policy: allow
  rules:
    - domains:
      - ".example.com"
    policy: allow
`

const invalidWildcardLocation = `
access_control:
  default_policy: allow
  rules:
    - domains:
      - "*example.com"
    policy: allow
`

const invalidDomainChar = `
access_control:
  default_policy: allow
  rules:
    - domains:
      - "example?.com"
    policy: allow
`

const invalidLeadingDash = `
access_control:
  default_policy: allow
  rules:
    - domains:
      - "-example.com"
    policy: allow
`

const invalidTrailingDash = `
access_control:
  default_policy: allow
  rules:
    - domains:
      - "example-.com"
    policy: allow
`

const invalidDomainString = `
access_control:
  default_policy: allow
  rules:
    - domains:
      - false
    policy: allow
`

const invalidNetworkString = `
access_control:
  default_policy: allow
  rules:
    - networks:
        - "invalid-cidr"
      policy: allow
`

const invalidNetworkNumber = `
access_control:
  default_policy: allow
  rules:
    - networks:
        - 10
      policy: allow
`

const invalidNetworkRange = `
access_control:
  default_policy: allow
  rules:
    - networks:
        - 300.300.300.300/50
      policy: allow
`

func TestReadConfigValid(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected *config.Configuration
	}{
		{
			"valid configuration",
			validConfig,
			&config.Configuration{
				AccessControl: config.AccessControl{
					DefaultPolicy: "allow",
					Rules: []config.AccessControlRule{
						{
							Policy: "allow",
							Networks: []config.CIDR{
								{
									IPNet: &net.IPNet{
										IP:   net.IP{10, 0, 0, 0},
										Mask: net.CIDRMask(8, 32),
									},
								},
								{
									IPNet: &net.IPNet{
										IP:   net.IP{127, 0, 0, 0},
										Mask: net.CIDRMask(8, 32),
									},
								},
							},
							Domains: []string{
								"example.com",
								"*.example.com",
							},
							Methods:           []string{"GET", "POST"},
							Countries:         []string{"US", "FR"},
							AutonomousSystems: []uint32{1234, 5678},
						},
						{
							Policy:            "deny",
							Networks:          nil,
							Domains:           nil,
							Methods:           nil,
							Countries:         nil,
							AutonomousSystems: nil,
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.data)
			cfg, err := config.ReadConfig(reader)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(*cfg, *test.expected) {
				t.Errorf("expected %v, got %v", test.expected, cfg)
			}
		})
	}
}

func TestReadConfigErr(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"invalid leading dot", invalidLeadingDot},
		{"invalid wildcard location", invalidWildcardLocation},
		{"invalid domain character", invalidDomainChar},
		{"invalid leading dash", invalidLeadingDash},
		{"invalid trailing dash", invalidTrailingDash},
		{"invalid network string", invalidNetworkString},
		{"invalid network number", invalidNetworkNumber},
		{"invalid network range", invalidNetworkRange},
		{"invalid domain string", invalidDomainString},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reader := strings.NewReader(test.data)
			_, err := config.ReadConfig(reader)
			if err == nil {
				t.Error("expected an error but got nil")
			}
		})
	}
}

type errReader struct{}

func (r *errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestReadConfigErrReader(t *testing.T) {
	_, err := config.ReadConfig(&errReader{})
	if err == nil {
		t.Error("expected an error but got nil")
	}
}
