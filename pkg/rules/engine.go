package rules

import (
	"net"
	"strings"

	"github.com/danroc/geoblock/pkg/schema"
	"github.com/danroc/geoblock/pkg/utils"
)

type Engine struct {
	config schema.AccessControl
}

func NewEngine(config schema.AccessControl) *Engine {
	return &Engine{
		config: config,
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
//
// Empty conditions are considered as "match all". For example, if a rule has
// no domains, it will match all domains.
//
// Domains and countries are case-insensitive.
func ruleApplies(query Query, rule schema.AccessControlRule) bool {
	if len(rule.Domains) > 0 {
		if utils.None(rule.Domains, func(domain string) bool {
			return strings.EqualFold(domain, query.RequestedDomain)
		}) {
			return false
		}
	}

	if len(rule.Networks) > 0 {
		if utils.None(rule.Networks, func(network schema.CIDR) bool {
			return network.Contains(query.SourceIP)
		}) {
			return false
		}
	}

	if len(rule.Countries) > 0 {
		if utils.None(rule.Countries, func(country string) bool {
			return strings.EqualFold(country, query.SourceCountry)
		}) {
			return false
		}
	}

	if len(rule.AutonomousSystems) > 0 {
		if utils.None(rule.AutonomousSystems, func(asn uint32) bool {
			return asn == query.SourceASN
		}) {
			return false
		}
	}

	return true
}

func (e *Engine) Authorize(q Query) bool {
	for _, rule := range e.config.Rules {
		if ruleApplies(q, rule) {
			return rule.Policy == schema.PolicyAllow
		}
	}

	return e.config.DefaultPolicy == schema.PolicyAllow
}
