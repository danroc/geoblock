package database

type DatabaseEntryData struct {
	CountryCode string `json:"country_code" validate:"required,iso3166_1_alpha2"`
}

type DatabaseEntry struct {
	StartIP string `json:"start_ip" validate:"required,ip"`
	EndIP   string `json:"end_ip"   validate:"required,ip"`
	Data    DatabaseEntryData
}
