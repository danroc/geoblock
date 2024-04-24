package config

import (
	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

const (
	PolicyAllow = "allow"
	PolicyDeny  = "deny"
)

type Rule struct {
	Policy    string   `yaml:"policy"                  validate:"required,oneof=allow deny"`
	Networks  []string `yaml:"networks,omitempty"      validate:"dive,cidr"`
	Domains   []string `yaml:"domains,omitempty"       validate:"dive,fqdn"`
	Countries []string `yaml:"country_codes,omitempty" validate:"dive,iso3166_1_alpha2"`
}

type Config struct {
	DefaultPolicy string `yaml:"default_policy" validate:"required,oneof=allow deny"`
	Rules         []Rule `yaml:"rules"          validate:"dive"`
}

// ParseConfig validates and parses the given YAML data into a Config struct.
func ParseConfig(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	return &config, nil
}
