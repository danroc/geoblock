package rules_test

import (
	"net/netip"
	"testing"

	"github.com/danroc/geoblock/internal/config"
	"github.com/danroc/geoblock/internal/rules"
)

func TestEngineAuthorize(t *testing.T) {
	tests := []struct {
		name   string
		config *config.AccessControl
		query  *rules.Query
		want   bool
	}{
		{
			name: "allow by default policy",
			config: &config.AccessControl{
				Rules:         []config.AccessControlRule{},
				DefaultPolicy: config.PolicyAllow,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
			},
			want: true,
		},
		{
			name: "deny by default policy",
			config: &config.AccessControl{
				Rules:         []config.AccessControlRule{},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
			},
			want: false,
		},
		{
			name: "allow by wildcard domain",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Domains: []string{"*.example.com"},
						Policy:  config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "sub.example.com",
			},
			want: true,
		},
		{
			name: "deny by wildcard domain",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Domains: []string{"*.example.com"},
						Policy:  config.PolicyDeny,
					},
				},
				DefaultPolicy: config.PolicyAllow,
			},
			query: &rules.Query{
				RequestedDomain: "sub.example.com",
			},
			want: false,
		},
		{
			name: "allow by domain",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Domains: []string{"example.org", "example.com"},
						Policy:  config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.org",
			},
			want: true,
		},
		{
			name: "deny by domain",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Domains: []string{"example.org", "example.com"},
						Policy:  config.PolicyDeny,
					},
				},
				DefaultPolicy: config.PolicyAllow,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
			},
			want: false,
		},
		{
			name: "deny unknown domain",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Domains: []string{"example.org"},
						Policy:  config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
			},
			want: false,
		},
		{
			name: "domains are case-insensitive",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Domains: []string{"example.org", "example.com"},
						Policy:  config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "EXAMPLE.ORG",
			},
			want: true,
		},
		{
			name: "allow by method",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Methods: []string{"GET", "POST"},
						Policy:  config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedMethod: "POST",
			},
			want: true,
		},
		{
			name: "deny by method",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Methods: []string{"GET", "POST"},
						Policy:  config.PolicyDeny,
					},
				},
				DefaultPolicy: config.PolicyAllow,
			},
			query: &rules.Query{
				RequestedMethod: "POST",
			},
			want: false,
		},
		{
			name: "deny unknown method",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Methods: []string{"GET"},
						Policy:  config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedMethod: "POST",
			},
			want: false,
		},
		{
			name: "methods are case-insensitive",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Methods: []string{"GET", "POST"},
						Policy:  config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedMethod: "get",
			},
			want: true,
		},
		{
			name: "allow by network",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Networks: []config.CIDR{
							{Prefix: netip.MustParsePrefix("10.0.0.0/8")},
							{Prefix: netip.MustParsePrefix("192.168.1.0/24")},
						},
						Policy: config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				SourceIP: netip.MustParseAddr("10.1.1.1"),
			},
			want: true,
		},
		{
			name: "deny by network",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Networks: []config.CIDR{
							{Prefix: netip.MustParsePrefix("10.0.0.0/8")},
							{Prefix: netip.MustParsePrefix("192.168.1.0/24")},
						},
						Policy: config.PolicyDeny,
					},
				},
				DefaultPolicy: config.PolicyAllow,
			},
			query: &rules.Query{
				SourceIP: netip.MustParseAddr("192.168.1.1"),
			},
			want: false,
		},
		{
			name: "allow by country",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Countries: []string{"FR", "US"},
						Policy:    config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				SourceCountry: "FR",
			},
			want: true,
		},
		{
			name: "deny by country",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Countries: []string{"FR", "US"},
						Policy:    config.PolicyDeny,
					},
				},
				DefaultPolicy: config.PolicyAllow,
			},
			query: &rules.Query{
				SourceCountry: "US",
			},
			want: false,
		},
		{
			name: "deny unknown country",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Countries: []string{"FR", "US"},
						Policy:    config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				SourceCountry: "DE",
			},
			want: false,
		},
		{
			name: "countries are case-insensitive",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Countries: []string{"FR", "US"},
						Policy:    config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				SourceCountry: "fr",
			},
			want: true,
		},
		{
			name: "allow by ASN",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						AutonomousSystems: []uint32{1111, 2222},
						Policy:            config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				SourceASN: 1111,
			},
			want: true,
		},
		{
			name: "deny by ASN",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						AutonomousSystems: []uint32{1111, 2222},
						Policy:            config.PolicyDeny,
					},
				},
				DefaultPolicy: config.PolicyAllow,
			},
			query: &rules.Query{
				SourceASN: 2222,
			},
			want: false,
		},
		{
			name: "deny unknown ASN",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						AutonomousSystems: []uint32{1111, 2222},
						Policy:            config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				SourceASN: 3333,
			},
			want: false,
		},
		{
			name: "allow by domain, network, country, and ASN",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Domains: []string{"example.com"},
						Networks: []config.CIDR{
							{Prefix: netip.MustParsePrefix("10.0.0.0/8")},
						},
						Countries:         []string{"FR"},
						AutonomousSystems: []uint32{1111},
						Policy:            config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
				SourceIP:        netip.MustParseAddr("10.1.1.1"),
				SourceCountry:   "FR",
				SourceASN:       1111,
			},
			want: true,
		},
		{
			name: "deny by default when query doesn't fully match rule",
			config: &config.AccessControl{
				Rules: []config.AccessControlRule{
					{
						Domains: []string{"example.com"},
						Networks: []config.CIDR{
							{Prefix: netip.MustParsePrefix("10.0.0.0/8")},
						},
						Policy: config.PolicyAllow,
					},
				},
				DefaultPolicy: config.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
				SourceIP:        netip.MustParseAddr("192.168.1.1"),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := rules.NewEngine(tt.config)
			if got := e.Authorize(tt.query); got != tt.want {
				t.Errorf("Engine.Authorize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEngineUpdateConfig(t *testing.T) {
	e := rules.NewEngine(&config.AccessControl{
		DefaultPolicy: config.PolicyAllow,
	})

	if got := e.Authorize(&rules.Query{}); got != true {
		t.Errorf("Engine.Authorize() = %v, want %v", got, true)
	}

	e.UpdateConfig(&config.AccessControl{
		DefaultPolicy: config.PolicyDeny,
	})

	if got := e.Authorize(&rules.Query{}); got != false {
		t.Errorf("Engine.Authorize() = %v, want %v", got, false)
	}
}
