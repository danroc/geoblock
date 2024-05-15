package configuration

import (
	"io"
	"os"

	"github.com/go-playground/validator/v10"
	"gopkg.in/yaml.v3"
)

// ReadBytes reads the configuration from the giver bytes slice.
func ReadBytes(data []byte) (*Configuration, error) {
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
	return ReadBytes(data)
}

// ReadFile reads the configuration from the given file.
func ReadFile(filename string) (*Configuration, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ReadBytes(data)
}