// Package config contains the schema and helper functions to work with the
// configuration file.
package config

import (
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// isCIDRField checks if the value of the given field is a valid CIDR.
func isCIDRField(field validator.FieldLevel) bool {
	cidr, ok := field.Field().Interface().(CIDR)
	if !ok || cidr.IPNet == nil {
		return false
	}
	return true
}

// read reads the configuration from the giver bytes slice.
func read(data []byte) (*Configuration, error) {
	var config Configuration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	validate := validator.New()
	validate.RegisterValidation("cidr", isCIDRField) // #nosec G104

	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadConfig reads the configuration from the given file.
func LoadConfig(filename string) (*Configuration, error) {
	data, err := os.ReadFile(filename) // #nosec G304
	if err != nil {
		return nil, err
	}
	return read(data)
}
