// Package rules contains all rules related logic.
package rules

import (
	"net/netip"
	"strings"
	"sync/atomic"

	"github.com/danroc/geoblock/internal/config"
	"github.com/danroc/geoblock/internal/utils/glob"
)

// Engine is the access control engine that checks if a given query is allowed
// by the rules.
type Engine struct {
	config atomic.Pointer[config.AccessControl]
}

// NewEngine creates a new access control engine for the given access control
// configuration.
func NewEngine(config *config.AccessControl) *Engine {
	e := &Engine{}
	e.config.Store(config)
	return e
}

// Query represents a query to be checked by the access control engine.
type Query struct {
	RequestedDomain string
	RequestedMethod string
	SourceIP        netip.Addr
	SourceCountry   string
	SourceASN       uint32
}

// match checks if any of the conditions match the given matchFunc.
func match[T any](conditions []T, matchFunc func(T) bool) bool {
	for _, condition := range conditions {
		if matchFunc(condition) {
			return true
		}
	}
	return len(conditions) == 0
}

// ruleApplies checks if the given query is allowed or denied by the given
// rule. For a rule to be applicable, the query must match all of the rule's
// conditions.
//
// Empty conditions are considered as "match all". For example, if a rule has
// no domains, it will match all domains.
//
// Domains, methods and countries are case-insensitive.
func ruleApplies(rule *config.AccessControlRule, query *Query) bool {
	matchDomain := match(rule.Domains, func(domain string) bool {
		return glob.Star(
			strings.ToLower(domain),
			strings.ToLower(query.RequestedDomain),
		)
	})

	matchMethod := match(rule.Methods, func(method string) bool {
		return strings.EqualFold(method, query.RequestedMethod)
	})

	matchIP := match(rule.Networks, func(network config.CIDR) bool {
		return network.Contains(query.SourceIP)
	})

	matchCountry := match(rule.Countries, func(country string) bool {
		return strings.EqualFold(country, query.SourceCountry)
	})

	matchASN := match(rule.AutonomousSystems, func(asn uint32) bool {
		return asn == query.SourceASN
	})

	return matchDomain && matchMethod && matchIP && matchCountry && matchASN
}

// UpdateConfig updates the engine's configuration with the given access
// control configuration.
func (e *Engine) UpdateConfig(config *config.AccessControl) {
	e.config.Store(config)
}

// Authorize checks if the given query is allowed by the engine's rules. The
// engine will return true if the query is allowed, false otherwise.
func (e *Engine) Authorize(query *Query) bool {
	cfg := e.config.Load()
	for _, rule := range cfg.Rules {
		if ruleApplies(&rule, query) {
			return rule.Policy == config.PolicyAllow
		}
	}
	return cfg.DefaultPolicy == config.PolicyAllow
}
