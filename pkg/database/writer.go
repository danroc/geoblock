package database

import (
	"encoding/json"
	"io"
	"os"
)

// Write writes the database to the given writer.
func (db *Database) Write(writer io.Writer) error {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	return err
}

// WriteFile writes the database to the given file.
func (db *Database) WriteFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return db.Write(file)
}
