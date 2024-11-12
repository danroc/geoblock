// Package rules contains all rules related logic.
package rules

import (
	"net"
	"strings"

	"github.com/danroc/geoblock/pkg/schema"
	"github.com/danroc/geoblock/pkg/utils"
)

// Engine is the access control egine that checks if a given query is allowed
// by the rules.
type Engine struct {
	config *schema.AccessControl
}

// NewEngine creates a new access control engine for the given access control
// configuration.
func NewEngine(config *schema.AccessControl) *Engine {
	return &Engine{
		config: config,
	}
}

// Query represents a query to be checked by the access control engine.
type Query struct {
	RequestedDomain string
	RequestedMethod string
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
// Domains, methods and countries are case-insensitive.
func ruleApplies(rule *schema.AccessControlRule, query *Query) bool {
	if len(rule.Domains) > 0 {
		if utils.None(rule.Domains, func(domain string) bool {
			return strings.EqualFold(domain, query.RequestedDomain)
		}) {
			return false
		}
	}

	if len(rule.Methods) > 0 {
		if utils.None(rule.Methods, func(method string) bool {
			return strings.EqualFold(method, query.RequestedMethod)
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

// Authorize checks if the given query is allowed by the engine's rules. The
// engine will return true if the query is allowed, false otherwise.
func (e *Engine) Authorize(query *Query) bool {
	for _, rule := range e.config.Rules {
		if ruleApplies(&rule, query) {
			return rule.Policy == schema.PolicyAllow
		}
	}

	return e.config.DefaultPolicy == schema.PolicyAllow
}
