package config

import (
	"net"
)

// CIDR represents a CIDR network. It's used to support unmarshaling from YAML.
type CIDR struct {
	*net.IPNet
}

// UnmarshalYAML unmarshals a CIDR network from YAML.
func (n *CIDR) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var network string
	if err := unmarshal(&network); err != nil {
		return err
	}

	_, ipNet, err := net.ParseCIDR(network)
	if err != nil {
		return err
	}

	n.IPNet = ipNet
	return nil
}
