package configuration

const (
	PolicyAllow = "allow"
	PolicyDeny  = "deny"
)

type AccessControlRule struct {
	Policy            string   `yaml:"policy"                       validate:"required,oneof=allow deny"`
	Networks          []string `yaml:"networks,omitempty"           validate:"dive,cidr"`
	Domains           []string `yaml:"domains,omitempty"            validate:"dive,fqdn"`
	Countries         []string `yaml:"countries,omitempty"          validate:"dive,iso3166_1_alpha2"`
	AutonomousSystems []uint32 `yaml:"autonomous_systems,omitempty" validate:"dive,numeric"`
}

type AccessControl struct {
	DefaultPolicy string              `yaml:"default_policy" validate:"required,oneof=allow deny"`
	Rules         []AccessControlRule `yaml:"rules"          validate:"dive"`
}

type LocationDatabase struct {
	DatabaseType  string `yaml:"database_type"  validate:"required,oneof=country asn"`
	DatabaseURL   string `yaml:"database_url"   validate:"required,url"`
	IPVersion     int    `yaml:"ip_version"     validate:"required,oneof=4 6"`
	CacheDuration string `yaml:"cache_duration" validate:"required,duration"`
	CacheLocation string `yaml:"cache_location" validate:"required,dir"`
}

type Configuration struct {
	AccessControl     AccessControl      `yaml:"access_control"`
	LocationDatabases []LocationDatabase `yaml:"location_databases"`
}
