package config_test

import (
	"errors"
	"net/netip"
	"reflect"
	"strings"
	"testing"

	"github.com/danroc/geoblock/internal/config"
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

const invalidPolicyValue = `
access_control:
  default_policy: invalid_policy
  rules: []
`

const missingDefaultPolicy = `
access_control:
  rules: []
`

const invalidMethodValue = `
access_control:
  default_policy: allow
  rules:
    - policy: allow
      methods:
        - INVALID_METHOD
`

const invalidCountryCode = `
access_control:
  default_policy: allow
  rules:
    - policy: allow
      countries:
        - INVALID
`

func TestReadConfig_Valid(t *testing.T) {
	reader := strings.NewReader(validConfig)

	cfg, err := config.ReadConfig(reader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := &config.Configuration{
		AccessControl: config.AccessControl{
			DefaultPolicy: "allow",
			Rules: []config.AccessControlRule{
				{
					Policy: "allow",
					Networks: []config.CIDR{
						{
							Prefix: netip.MustParsePrefix(
								"10.0.0.0/8",
							),
						},
						{
							Prefix: netip.MustParsePrefix(
								"127.0.0.0/8",
							),
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
	}

	if !reflect.DeepEqual(*cfg, *expected) {
		t.Errorf("expected %v, got %v", expected, cfg)
	}
}

func TestReadConfig_Err(t *testing.T) {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			_, err := config.ReadConfig(reader)
			if err == nil {
				t.Error("expected an error but got nil")
			}
		})
	}
}

func TestReadConfig_ValidationErrors(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{"invalid policy value", invalidPolicyValue},
		{"missing default policy", missingDefaultPolicy},
		{"invalid method value", invalidMethodValue},
		{"invalid country code", invalidCountryCode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			_, err := config.ReadConfig(reader)
			if err == nil {
				t.Error("expected validation error but got nil")
			}
		})
	}
}

type errReader struct{}

func (r *errReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestReadConfig_ErrReader(t *testing.T) {
	_, err := config.ReadConfig(&errReader{})
	if err == nil {
		t.Error("expected an error but got nil")
	}
}
