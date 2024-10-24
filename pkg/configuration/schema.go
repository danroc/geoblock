package configuration

import "net"

const (
	PolicyAllow = "allow"
	PolicyDeny  = "deny"
)

type CIDR struct {
	*net.IPNet
}

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

type AccessControlRule struct {
	Policy            string   `yaml:"policy"                       validate:"required,oneof=allow deny"`
	Networks          []CIDR   `yaml:"networks,omitempty"           validate:"dive,cidr"`
	Domains           []string `yaml:"domains,omitempty"            validate:"dive,fqdn"`
	Countries         []string `yaml:"countries,omitempty"          validate:"dive,iso3166_1_alpha2"`
	AutonomousSystems []uint32 `yaml:"autonomous_systems,omitempty" validate:"dive,numeric"`
}

type AccessControl struct {
	DefaultPolicy string              `yaml:"default_policy" validate:"required,oneof=allow deny"`
	Rules         []AccessControlRule `yaml:"rules"          validate:"dive"`
}

type Configuration struct {
	AccessControl AccessControl `yaml:"access_control"`
}
