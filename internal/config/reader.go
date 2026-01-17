// Package config contains the schema and helper functions to work with the configuration file.
package config

import (
	"io"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

// DomainNameRegex matches a valid domain name as per RFC 1035. It also allows labels to be a
// single `*` wildcard.
var domainNameRegex = regexp.MustCompile(
	`^(\*|[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)(\.(\*|[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?))*$`,
)

// isDomainNameField checks if the value of the given field is a valid domain name. It also allows
// labels to be a single `*` wildcard.
func isDomainNameField(field validator.FieldLevel) bool {
	domain, ok := field.Field().Interface().(string)
	return ok && domainNameRegex.MatchString(domain)
}

// isCIDRField checks if the value of the given field is a valid CIDR.
func isCIDRField(field validator.FieldLevel) bool {
	_, ok := field.Field().Interface().(CIDR)
	return ok
}

// read reads the configuration from the giver bytes slice.
func read(data []byte) (*Configuration, error) {
	var config Configuration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	validate := validator.New()
	validate.RegisterValidation("cidr", isCIDRField)         // #nosec G104
	validate.RegisterValidation("domain", isDomainNameField) // #nosec G104

	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ReadConfig reads the configuration from the given reader and returns it.
func ReadConfig(reader io.Reader) (*Configuration, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return read(data)
}
