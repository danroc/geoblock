package database

import (
	"encoding/json"
	"io"
	"os"

	"github.com/go-playground/validator/v10"
)

// ReadBytes reads the database from the given bytes slice.
func ReadBytes(data []byte) (*Database, error) {
	var db Database
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, err
	}

	validate := validator.New()
	if err := validate.Struct(db); err != nil {
		return nil, err
	}

	return &db, nil
}

// Read reads the database from the given reader.
func Read(reader io.Reader) (*Database, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return ReadBytes(data)
}

// ReadFile reads the database from the given file.
func ReadFile(filename string) (*Database, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ReadBytes(data)
}
