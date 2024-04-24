package configuration

import (
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	PolicyAllow = "allow"
	PolicyDeny  = "deny"
)

type AccessControlRule struct {
	Policy    string   `yaml:"policy"                  validate:"required,oneof=allow deny"`
	Networks  []string `yaml:"networks,omitempty"      validate:"dive,cidr"`
	Domains   []string `yaml:"domains,omitempty"       validate:"dive,fqdn"`
	Countries []string `yaml:"country_codes,omitempty" validate:"dive,iso3166_1_alpha2"`
}

type AccessControl struct {
	DefaultPolicy string              `yaml:"default_policy" validate:"required,oneof=allow deny"`
	Rules         []AccessControlRule `yaml:"rules"          validate:"dive"`
}

type Configuration struct {
	AccessControl AccessControl `yaml:"access_control"`
}

// ParseConfiguration validates and parses the given YAML data into a
// Configuration struct.
func ParseConfiguration(data []byte) (*Configuration, error) {
	var config Configuration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	return &config, nil
}
