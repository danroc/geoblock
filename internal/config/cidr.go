package config

import (
	"net/netip"
)

// CIDR represents a CIDR network. It's used to support unmarshaling from YAML.
type CIDR struct {
	netip.Prefix
}

// UnmarshalYAML unmarshals a CIDR network from YAML.
func (c *CIDR) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var network string
	if err := unmarshal(&network); err != nil {
		return err
	}

	prefix, err := netip.ParsePrefix(network)
	if err != nil {
		return err
	}

	c.Prefix = prefix
	return nil
}
