package database_test

import (
	"errors"
	"net"
	"strings"
	"testing"

	"github.com/danroc/geoblock/pkg/database"
)

const (
	// Valid CSV data
	csvData1 = "192.168.1.1,192.168.1.255,data1,data2\n" + "10.0.0.1,10.0.0.255,data3,data4\n"

	// Missing start IP
	csvData2 = ",192.168.1.2,data1,data2\n"

	// Missing end IP
	csvData3 = "192.168.1.1,,data1,data2\n"

	// Missing data
	csvData4 = "192.168.1.1,192.168.1.2\n"

	// No CSV data
	csvData5 = "\n"

	// Invalid start IP
	csvData6 = "invalid-ip,192.168.1.2,data1,data2\n"

	// Invalid end IP
	csvData7 = "192.168.1.1,invalid-ip,data1,data2\n"
)

func TestNewDatabase(t *testing.T) {
	tests := []struct {
		name string
		data string
		err  bool
	}{
		{"Valid CSV data", csvData1, false},
		{"Missing start IP", csvData2, true},
		{"Missing end IP", csvData3, true},
		{"Missing data", csvData4, false},
		{"No CSV data", csvData5, false},
		{"Invalid start IP", csvData6, true},
		{"Invalid end IP", csvData7, true},
	}

	for _, test := range tests {
		reader := strings.NewReader(test.data)
		err := database.NewDatabase().Update(reader)
		if test.err && err == nil {
			t.Errorf("%s: expected an error but got nil", test.name)
		}
		if !test.err && err != nil {
			t.Errorf("%s: expected no error but got %v", test.name, err)
		}
	}
}

type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestNewDatabaseReadErr(t *testing.T) {
	reader := &errorReader{}
	err := database.NewDatabase().Update(reader)
	if err == nil {
		t.Fatalf("Expected an error but got nil")
	}
}

func TestFind(t *testing.T) {
	reader := strings.NewReader(csvData1)
	db := database.NewDatabase()
	if err := db.Update(reader); err != nil {
		t.Fatalf("Expected no error but got %v", err)
	}

	tests := []struct {
		ip       string
		expected []string
	}{
		{"192.168.1.50", []string{"data1", "data2"}},
		{"10.0.0.50", []string{"data3", "data4"}},
		{"172.16.0.1", nil},
		{"1.1.1.1", nil},
		{"invalid", nil},
	}

	for _, test := range tests {
		ip := net.ParseIP(test.ip)
		result := db.Find(ip)
		if len(result) != len(test.expected) {
			t.Errorf("Expected %v but got %v", test.expected, result)
		}
		for i, v := range result {
			if v != test.expected[i] {
				t.Errorf("Expected %s but got %s", test.expected[i], v)
			}
		}
	}
}
