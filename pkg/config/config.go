package config

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
