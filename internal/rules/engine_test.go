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
			if got := e.Authorize(tt.query).Allowed; got != tt.want {
				t.Errorf("Engine.Authorize().Allowed = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAuthorizationResult(t *testing.T) {
	tests := []struct {
		name          string
		ruleIndex     int
		action        string
		wantAllowed   bool
		wantRuleIndex int
		wantAction    string
		wantIsDefault bool
	}{
		{
			name:          "allow action with rule match",
			ruleIndex:     0,
			action:        config.PolicyAllow,
			wantAllowed:   true,
			wantRuleIndex: 0,
			wantAction:    config.PolicyAllow,
			wantIsDefault: false,
		},
		{
			name:          "deny action with rule match",
			ruleIndex:     0,
			action:        config.PolicyDeny,
			wantAllowed:   false,
			wantRuleIndex: 0,
			wantAction:    config.PolicyDeny,
			wantIsDefault: false,
		},
		{
			name:          "allow action with default policy",
			ruleIndex:     rules.NoMatchingRuleIndex,
			action:        config.PolicyAllow,
			wantAllowed:   true,
			wantRuleIndex: rules.NoMatchingRuleIndex,
			wantAction:    config.PolicyAllow,
			wantIsDefault: true,
		},
		{
			name:          "deny action with default policy",
			ruleIndex:     rules.NoMatchingRuleIndex,
			action:        config.PolicyDeny,
			wantAllowed:   false,
			wantRuleIndex: rules.NoMatchingRuleIndex,
			wantAction:    config.PolicyDeny,
			wantIsDefault: true,
		},
		{
			name:          "allow action with higher rule index",
			ruleIndex:     5,
			action:        config.PolicyAllow,
			wantAllowed:   true,
			wantRuleIndex: 5,
			wantAction:    config.PolicyAllow,
			wantIsDefault: false,
		},
		{
			name:          "unknown action treated as deny",
			ruleIndex:     0,
			action:        "unknown",
			wantAllowed:   false,
			wantRuleIndex: 0,
			wantAction:    "unknown",
			wantIsDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rules.NewAuthorizationResult(tt.ruleIndex, tt.action)

			if got.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", got.Allowed, tt.wantAllowed)
			}
			if got.RuleIndex != tt.wantRuleIndex {
				t.Errorf("RuleIndex = %v, want %v", got.RuleIndex, tt.wantRuleIndex)
			}
			if got.Action != tt.wantAction {
				t.Errorf("Action = %v, want %v", got.Action, tt.wantAction)
			}
			if got.IsDefaultPolicy != tt.wantIsDefault {
				t.Errorf(
					"IsDefaultPolicy = %v, want %v",
					got.IsDefaultPolicy,
					tt.wantIsDefault,
				)
			}
		})
	}
}

func TestEngineUpdateConfig(t *testing.T) {
	e := rules.NewEngine(&config.AccessControl{
		DefaultPolicy: config.PolicyAllow,
	})

	if got := e.Authorize(&rules.Query{}).Allowed; got != true {
		t.Errorf("Engine.Authorize().Allowed = %v, want %v", got, true)
	}

	e.UpdateConfig(&config.AccessControl{
		DefaultPolicy: config.PolicyDeny,
	})

	if got := e.Authorize(&rules.Query{}).Allowed; got != false {
		t.Errorf("Engine.Authorize().Allowed = %v, want %v", got, false)
	}
}
