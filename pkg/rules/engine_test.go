package rules

import (
	"net"
	"testing"

	"github.com/danroc/geoblock/pkg/schema"
)

func TestEngine_Authorize(t *testing.T) {
	tests := []struct {
		name   string
		config *schema.AccessControl
		query  *Query
		want   bool
	}{
		{
			name: "allow by default policy",
			config: &schema.AccessControl{
				Rules:         []schema.AccessControlRule{},
				DefaultPolicy: schema.PolicyAllow,
			},
			query: &Query{
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
			query: &Query{
				RequestedDomain: "example.com",
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
			query: &Query{
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
			query: &Query{
				RequestedDomain: "example.com",
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
			query: &Query{
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
			query: &Query{
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
			query: &Query{
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
			query: &Query{
				SourceCountry: "US",
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
			query: &Query{
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
			query: &Query{
				SourceASN: 2222,
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
			query: &Query{
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
			query: &Query{
				RequestedDomain: "example.com",
				SourceIP:        net.IPv4(192, 168, 1, 1),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := NewEngine(tt.config)
			if got := e.Authorize(tt.query); got != tt.want {
				t.Errorf("Engine.Authorize() = %v, want %v", got, tt.want)
			}
		})
	}
}
