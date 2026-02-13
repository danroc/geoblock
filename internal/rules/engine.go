// Package rules contains all rules related logic.
package rules

import (
	"net/netip"
	"strings"
	"sync/atomic"

	"github.com/danroc/geoblock/internal/config"
	"github.com/danroc/geoblock/internal/utils/glob"
)

// NoMatchingRuleIndex is the rule index used when no rule matched and the default
// policy was applied.
const NoMatchingRuleIndex = -1

// Engine is the access control engine that checks if a given query is allowed by the
// rules.
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

// match checks if any of the conditions match the given matchFunc. If the conditions
// slice is empty, it returns true (match all).
func match[T any](conditions []T, matchFunc func(T) bool) bool {
	for _, condition := range conditions {
		if matchFunc(condition) {
			return true
		}
	}
	return len(conditions) == 0
}

// ruleApplies checks if the given query is allowed or denied by the given rule. For a
// rule to be applicable, the query must match all of the rule's conditions.
//
// Empty conditions are considered as "match all". For example, if a rule has no
// domains, it will match all domains.
//
// Domains, methods and countries are case-insensitive.
func ruleApplies(
	rule *config.AccessControlRule,
	query *Query,
) bool {
	if !match(rule.Domains, func(domain string) bool {
		return glob.MatchFold(domain, query.RequestedDomain)
	}) {
		return false
	}

	if !match(rule.Methods, func(method string) bool {
		return strings.EqualFold(method, query.RequestedMethod)
	}) {
		return false
	}

	if !match(rule.Networks, func(network config.CIDR) bool {
		return network.Contains(query.SourceIP)
	}) {
		return false
	}

	if !match(rule.Countries, func(country string) bool {
		return strings.EqualFold(country, query.SourceCountry)
	}) {
		return false
	}

	if !match(rule.AutonomousSystems, func(asn uint32) bool {
		return asn == query.SourceASN
	}) {
		return false
	}

	return true
}

// UpdateConfig updates the engine's configuration with the given access control
// configuration.
func (e *Engine) UpdateConfig(config *config.AccessControl) {
	e.config.Store(config)
}

// AuthorizationResult contains the result of an authorization check with metadata.
// RuleIndex is NoMatchingRuleIndex if the default policy was used.
type AuthorizationResult struct {
	RuleIndex int
	Action    string
}

// Allowed reports whether the result permits access.
func (r AuthorizationResult) Allowed() bool {
	return r.Action == config.PolicyAllow
}

// IsDefaultPolicy reports whether the default policy was applied.
func (r AuthorizationResult) IsDefaultPolicy() bool {
	return r.RuleIndex == NoMatchingRuleIndex
}

// Authorize checks if the given query is allowed by the engine's rules and returns
// detailed result including which rule matched.
func (e *Engine) Authorize(query *Query) AuthorizationResult {
	// Loop through all rules and apply the first matching rule.
	cfg := e.config.Load()
	for i, rule := range cfg.Rules {
		if ruleApplies(&rule, query) {
			return AuthorizationResult{
				RuleIndex: i,
				Action:    rule.Policy,
			}
		}
	}

	// No rule matched, apply default policy.
	return AuthorizationResult{
		RuleIndex: NoMatchingRuleIndex,
		Action:    cfg.DefaultPolicy,
	}
}
