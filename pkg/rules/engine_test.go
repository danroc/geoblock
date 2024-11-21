package rules_test

import (
	"net"
	"testing"

	"github.com/danroc/geoblock/pkg/rules"
	"github.com/danroc/geoblock/pkg/schema"
)

func TestEngineAuthorize(t *testing.T) {
	tests := []struct {
		name   string
		config *schema.AccessControl
		query  *rules.Query
		want   bool
	}{
		{
			name: "allow by default policy",
			config: &schema.AccessControl{
				Rules:         []schema.AccessControlRule{},
				DefaultPolicy: schema.PolicyAllow,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
			},
			want: true,
		},
		{
			name: "deny by default policy",
			config: &schema.AccessControl{
				Rules:         []schema.AccessControlRule{},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
			},
			want: false,
		},
		{
			name: "allow by wildcard domain",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Domains: []string{"*.example.com"},
						Policy:  schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "sub.example.com",
			},
			want: true,
		},
		{
			name: "deny by wildcard domain",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Domains: []string{"*.example.com"},
						Policy:  schema.PolicyDeny,
					},
				},
				DefaultPolicy: schema.PolicyAllow,
			},
			query: &rules.Query{
				RequestedDomain: "sub.example.com",
			},
			want: false,
		},
		{
			name: "allow by domain",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Domains: []string{"example.org", "example.com"},
						Policy:  schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.org",
			},
			want: true,
		},
		{
			name: "deny by domain",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Domains: []string{"example.org", "example.com"},
						Policy:  schema.PolicyDeny,
					},
				},
				DefaultPolicy: schema.PolicyAllow,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
			},
			want: false,
		},
		{
			name: "deny unknown domain",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Domains: []string{"example.org"},
						Policy:  schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
			},
			want: false,
		},
		{
			name: "allow by method",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Methods: []string{"GET", "POST"},
						Policy:  schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				RequestedMethod: "POST",
			},
			want: true,
		},
		{
			name: "deny by method",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Methods: []string{"GET", "POST"},
						Policy:  schema.PolicyDeny,
					},
				},
				DefaultPolicy: schema.PolicyAllow,
			},
			query: &rules.Query{
				RequestedMethod: "POST",
			},
			want: false,
		},
		{
			name: "deny unknown method",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Methods: []string{"GET"},
						Policy:  schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				RequestedMethod: "POST",
			},
			want: false,
		},
		{
			name: "allow by network",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Networks: []schema.CIDR{
							{IPNet: &net.IPNet{
								IP:   net.IPv4(10, 0, 0, 0),
								Mask: net.CIDRMask(8, 32),
							}},
							{IPNet: &net.IPNet{
								IP:   net.IPv4(192, 168, 1, 0),
								Mask: net.CIDRMask(24, 32),
							}},
						},
						Policy: schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				SourceIP: net.IPv4(10, 1, 1, 1),
			},
			want: true,
		},
		{
			name: "deny by network",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Networks: []schema.CIDR{
							{IPNet: &net.IPNet{
								IP:   net.IPv4(10, 0, 0, 0),
								Mask: net.CIDRMask(8, 32),
							}},
							{IPNet: &net.IPNet{
								IP:   net.IPv4(192, 168, 1, 0),
								Mask: net.CIDRMask(24, 32),
							}},
						},
						Policy: schema.PolicyDeny,
					},
				},
				DefaultPolicy: schema.PolicyAllow,
			},
			query: &rules.Query{
				SourceIP: net.IPv4(192, 168, 1, 1),
			},
			want: false,
		},
		{
			name: "allow by country",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Countries: []string{"FR", "US"},
						Policy:    schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				SourceCountry: "FR",
			},
			want: true,
		},
		{
			name: "deny by country",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Countries: []string{"FR", "US"},
						Policy:    schema.PolicyDeny,
					},
				},
				DefaultPolicy: schema.PolicyAllow,
			},
			query: &rules.Query{
				SourceCountry: "US",
			},
			want: false,
		},
		{
			name: "deny unknown country",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Countries: []string{"FR", "US"},
						Policy:    schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				SourceCountry: "DE",
			},
			want: false,
		},
		{
			name: "allow by ASN",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						AutonomousSystems: []uint32{1111, 2222},
						Policy:            schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				SourceASN: 1111,
			},
			want: true,
		},
		{
			name: "deny by ASN",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						AutonomousSystems: []uint32{1111, 2222},
						Policy:            schema.PolicyDeny,
					},
				},
				DefaultPolicy: schema.PolicyAllow,
			},
			query: &rules.Query{
				SourceASN: 2222,
			},
			want: false,
		},
		{
			name: "deny unknown ASN",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						AutonomousSystems: []uint32{1111, 2222},
						Policy:            schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				SourceASN: 3333,
			},
			want: false,
		},
		{
			name: "allow by domain, network, country, and ASN",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Domains: []string{"example.com"},
						Networks: []schema.CIDR{
							{IPNet: &net.IPNet{
								IP:   net.IPv4(10, 0, 0, 0),
								Mask: net.CIDRMask(8, 32),
							}},
						},
						Countries:         []string{"FR"},
						AutonomousSystems: []uint32{1111},
						Policy:            schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
				SourceIP:        net.IPv4(10, 1, 1, 1),
				SourceCountry:   "FR",
				SourceASN:       1111,
			},
			want: true,
		},
		{
			name: "deny by default when query doesn't fully match rule",
			config: &schema.AccessControl{
				Rules: []schema.AccessControlRule{
					{
						Domains: []string{"example.com"},
						Networks: []schema.CIDR{
							{IPNet: &net.IPNet{
								IP:   net.IPv4(10, 0, 0, 0),
								Mask: net.CIDRMask(8, 32),
							}},
						},
						Policy: schema.PolicyAllow,
					},
				},
				DefaultPolicy: schema.PolicyDeny,
			},
			query: &rules.Query{
				RequestedDomain: "example.com",
				SourceIP:        net.IPv4(192, 168, 1, 1),
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
	e := rules.NewEngine(&schema.AccessControl{
		DefaultPolicy: schema.PolicyAllow,
	})

	if got := e.Authorize(&rules.Query{}); got != true {
		t.Errorf("Engine.Authorize() = %v, want %v", got, true)
	}

	e.UpdateConfig(&schema.AccessControl{
		DefaultPolicy: schema.PolicyDeny,
	})

	if got := e.Authorize(&rules.Query{}); got != false {
		t.Errorf("Engine.Authorize() = %v, want %v", got, false)
	}
}
