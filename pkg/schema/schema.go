package schema

import "net"

const (
	PolicyAllow = "allow"
	PolicyDeny  = "deny"
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

// AccessControlRule represents an access control rule.
type AccessControlRule struct {
	Policy            string   `yaml:"policy"                       validate:"required,oneof=allow deny"`
	Networks          []CIDR   `yaml:"networks,omitempty"           validate:"dive,cidr"`
	Domains           []string `yaml:"domains,omitempty"            validate:"dive,fqdn"`
	Countries         []string `yaml:"countries,omitempty"          validate:"dive,iso3166_1_alpha2"`
	AutonomousSystems []uint32 `yaml:"autonomous_systems,omitempty" validate:"dive,numeric"`
}

// AccessControl represents the access control configuration.
type AccessControl struct {
	DefaultPolicy string              `yaml:"default_policy" validate:"required,oneof=allow deny"`
	Rules         []AccessControlRule `yaml:"rules"          validate:"dive"`
}

// Configuration represents the configuration of the application.
type Configuration struct {
	AccessControl AccessControl `yaml:"access_control"`
}
