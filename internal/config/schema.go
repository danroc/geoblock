package config

// Accepted policy values
const (
	PolicyAllow = "allow"
	PolicyDeny  = "deny"
)

// AccessControlRule represents an access control rule.
type AccessControlRule struct {
	Policy            string   `yaml:"policy"                       validate:"required,oneof=allow deny"`
	Networks          []CIDR   `yaml:"networks,omitempty"           validate:"dive,cidr"`
	Domains           []string `yaml:"domains,omitempty"            validate:"dive,domain"`
	Methods           []string `yaml:"methods,omitempty"            validate:"dive,oneof=GET HEAD POST PUT DELETE PATCH"`
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
