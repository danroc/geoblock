package database

import "time"

type IPRange struct {
	StartIP     string `json:"start_ip"     validate:"required,ip"`
	EndIP       string `json:"end_ip"       validate:"required,ip"`
	CountryCode string `json:"country_code" validate:"required,iso3166_1_alpha2"`
}

type Database struct {
	UpdatedAt string    `json:"updated_at" validate:"datetime"`
	Ranges    []IPRange `json:"ranges"     validate:"dive"`
}

func NewDatabase() *Database {
	return &Database{
		UpdatedAt: time.Now().Format(time.RFC3339),
		Ranges:    make([]IPRange, 0),
	}
}
