package configuration

import (
	"io"
	"os"

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

// Load reads the configuration from the giver bytes slice.
func Load(data []byte) (*Configuration, error) {
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

// Read reads then configuration from the given reader.
func Read(reader io.Reader) (*Configuration, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return Load(data)
}

// ReadFile reads the configuration from the given file.
func ReadFile(filename string) (*Configuration, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Load(data)
}
