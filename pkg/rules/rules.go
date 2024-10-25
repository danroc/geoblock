package rules

import (
	"net"

	"github.com/danroc/geoblock/pkg/configuration"
)

type Engine struct {
	config configuration.AccessControl
}

func NewEngine(config configuration.Configuration) *Engine {
	return &Engine{
		config: config.AccessControl,
	}
}

type Query struct {
	RequestedDomain string
	SourceIP        net.IP
	SourceCountry   string
	SourceASN       uint32
}

// ruleApplies checks if the given query is allowed or denied by the given
// rule. For a rule to be applicable, the query must match all of the rule's
// conditions.
func ruleApplies(query Query, rule configuration.AccessControlRule) bool {
	// Check country rules
	if len(rule.Countries) > 0 {
		found := false
		for _, country := range rule.Countries {
			if country == query.SourceCountry {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	// Check domain rules
	if len(rule.Domains) > 0 {
		found := false
		for _, domain := range rule.Domains {
			if domain == query.RequestedDomain {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	// Check ASN rules
	if len(rule.AutonomousSystems) > 0 {
		found := false
		for _, asn := range rule.AutonomousSystems {
			if asn == query.SourceASN {
				found = true
				break
			}
		}

		if !found {
			return false
		}
	}

	return true
}

func (e *Engine) Authorize(q Query) bool {
	for _, rule := range e.config.Rules {
		if ruleApplies(q, rule) {
			return rule.Policy == configuration.PolicyAllow
		}
	}

	return e.config.DefaultPolicy == configuration.PolicyAllow
}
