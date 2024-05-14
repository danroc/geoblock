package configuration

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

type FileDatabase struct {
	SourceURL       string `yaml:"source_url"       validate:"http_url"`
	CacheFile       string `yaml:"cache_file"       validate:"filepath"`
	RefreshInterval string `yaml:"refresh_interval" validate:"duration"`
}

type Databases struct {
	IPv4 FileDatabase `yaml:"ipv4"`
	IPv6 FileDatabase `yaml:"ipv6"`
}

type Configuration struct {
	DecisionsFile string        `yaml:"decisions_file" validate:"filepath"`
	AccessControl AccessControl `yaml:"access_control"`
	Databases     Databases     `yaml:"databases"`
}
