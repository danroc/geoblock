package main

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		setValue string
		fallback string
		want     string
	}{
		{
			name:     "env set",
			key:      "TEST_ENV_KEY",
			setValue: "test_value",
			fallback: "fallback",
			want:     "test_value",
		},
		{
			name:     "env not set",
			key:      "NON_EXISTENT_KEY",
			setValue: "",
			fallback: "fallback",
			want:     "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setValue != "" {
				t.Setenv(tt.key, tt.setValue)
			}
			got := getEnv(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf(
					"getEnv(%q, %q) = %q, want %q",
					tt.key, tt.fallback, got, tt.want,
				)
			}
		})
	}
}

func TestGetOptions(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *appOptions
	}{
		{
			name: "all custom values",
			envVars: map[string]string{
				OptionConfigPath: "/tmp/test.yaml",
				OptionServerPort: "1234",
				OptionLogLevel:   LogLevelDebug,
				OptionLogFormat:  LogFormatText,
			},
			want: &appOptions{
				configPath: "/tmp/test.yaml",
				serverPort: "1234",
				logLevel:   LogLevelDebug,
				logFormat:  LogFormatText,
			},
		},
		{
			name:    "default values",
			envVars: map[string]string{},
			want: &appOptions{
				configPath: DefaultConfigPath,
				serverPort: DefaultServerPort,
				logLevel:   DefaultLogLevel,
				logFormat:  DefaultLogFormat,
			},
		},
		{
			name: "mixed values",
			envVars: map[string]string{
				OptionConfigPath: "/custom/config.yaml",
				OptionLogLevel:   LogLevelDebug,
			},
			want: &appOptions{
				configPath: "/custom/config.yaml",
				serverPort: DefaultServerPort,
				logLevel:   LogLevelDebug,
				logFormat:  DefaultLogFormat,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			got := getOptions()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOptions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestHasChanged(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name string
		a, b fakeFileInfo
		want bool
	}{
		{
			name: "identical",
			a:    fakeFileInfo{name: "a", size: 10, mod: now},
			b:    fakeFileInfo{name: "a", size: 10, mod: now},
			want: false,
		},
		{
			name: "different size",
			a:    fakeFileInfo{name: "a", size: 10, mod: now},
			b:    fakeFileInfo{name: "a", size: 20, mod: now},
			want: true,
		},
		{
			name: "different mod",
			a:    fakeFileInfo{name: "a", size: 10, mod: now},
			b:    fakeFileInfo{name: "a", size: 10, mod: now.Add(time.Second)},
			want: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := hasChanged(tt.a, tt.b)
			if got != tt.want {
				t.Errorf(
					"hasChanged(%v, %v) = %v, want %v",
					tt.a, tt.b, got, tt.want,
				)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	cases := []struct {
		input    string
		expected zerolog.Level
		wantErr  bool
	}{
		{"trace", zerolog.TraceLevel, false},
		{"debug", zerolog.DebugLevel, false},
		{"info", zerolog.InfoLevel, false},
		{"warn", zerolog.WarnLevel, false},
		{"error", zerolog.ErrorLevel, false},
		{"fatal", zerolog.FatalLevel, false},
		{"panic", zerolog.PanicLevel, false},
		{"invalid", zerolog.InfoLevel, true},
		{"", zerolog.InfoLevel, true},
	}

	for _, c := range cases {
		parsed, err := parseLogLevel(c.input)
		if parsed != c.expected {
			t.Errorf(
				"parseLogLevel(%q) = %v, want %v",
				c.input,
				parsed,
				c.expected,
			)
		}
		if (err != nil) != c.wantErr {
			t.Errorf(
				"parseLogLevel(%q) error = %v, wantErr %v",
				c.input,
				err,
				c.wantErr,
			)
		}
	}
}

type fakeFileInfo struct {
	name string
	size int64
	mod  time.Time
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return f.size }
func (f fakeFileInfo) Mode() os.FileMode  { return 0 }
func (f fakeFileInfo) ModTime() time.Time { return f.mod }
func (f fakeFileInfo) IsDir() bool        { return false }
func (f fakeFileInfo) Sys() interface{}   { return nil }
