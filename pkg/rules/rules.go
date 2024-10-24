package rules

import (
	"net"

	"github.com/danroc/geoblock/pkg/configuration"
)

type rule struct {
	policy    string
	networks  []net.IPNet
	domains   []string
	countries []string
	asn       []uint32
}

type Engine struct {
	defaultPolicy string
	rules         []rule
}

func NewEngine(config configuration.Configuration) (*Engine, error) {
	controller := &Engine{
		defaultPolicy: config.AccessControl.DefaultPolicy,
		rules:         make([]rule, len(config.AccessControl.Rules)),
	}

	for i, r := range config.AccessControl.Rules {
		controller.rules[i] = rule{
			policy:    r.Policy,
			networks:  make([]net.IPNet, len(r.Networks)),
			domains:   r.Domains,
			countries: r.Countries,
			asn:       r.AutonomousSystems,
		}

		for j, network := range r.Networks {
			_, ipNet, err := net.ParseCIDR(network)
			if err != nil {
				return nil, err
			}
			controller.rules[i].networks[j] = *ipNet
		}
	}
	return controller, nil
}
