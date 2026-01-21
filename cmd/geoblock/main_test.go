package main

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/danroc/geoblock/internal/config"
)

const (
	testTimeout      = time.Second
	testTickInterval = 10 * time.Millisecond
)

// Test helpers

// testContextCancellation runs fn in a goroutine and verifies it returns
// promptly after ctx is canceled.
func testContextCancellation(t *testing.T, fn func(ctx context.Context)) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		fn(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Success
	case <-time.After(testTimeout):
		t.Fatal("function did not return after context cancellation")
	}
}

// Test doubles

// fakeFileInfo implements os.FileInfo for testing file stat comparisons.
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
func (f fakeFileInfo) Sys() any           { return nil }

type mockConfigUpdater struct {
	called bool
}

func (m *mockConfigUpdater) UpdateConfig(*config.AccessControl) {
	m.called = true
}

type mockServer struct {
	shutdownCalled bool
	shutdownErr    error
}

func (m *mockServer) Shutdown(context.Context) error {
	m.shutdownCalled = true
	return m.shutdownErr
}

// mockUpdater implements the updater interface for testing.
type mockUpdater struct {
	callCount int
}

func (m *mockUpdater) Update() error {
	m.callCount++
	return nil
}

// Tests

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

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected zerolog.Level
		wantErr  bool
	}{
		{"trace", "trace", zerolog.TraceLevel, false},
		{"debug", "debug", zerolog.DebugLevel, false},
		{"info", "info", zerolog.InfoLevel, false},
		{"warn", "warn", zerolog.WarnLevel, false},
		{"error", "error", zerolog.ErrorLevel, false},
		{"fatal", "fatal", zerolog.FatalLevel, false},
		{"panic", "panic", zerolog.PanicLevel, false},
		{"invalid", "invalid", zerolog.InfoLevel, true},
		{"empty", "", zerolog.InfoLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parseLogLevel(tt.input)
			if parsed != tt.expected {
				t.Errorf(
					"parseLogLevel(%q) = %v, want %v",
					tt.input,
					parsed,
					tt.expected,
				)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf(
					"parseLogLevel(%q) error = %v, wantErr %v",
					tt.input,
					err,
					tt.wantErr,
				)
			}
		})
	}
}

func TestConfigureLogger(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		level     string
		wantLevel zerolog.Level
	}{
		{
			name:      "json format with valid level",
			format:    LogFormatJSON,
			level:     LogLevelDebug,
			wantLevel: zerolog.DebugLevel,
		},
		{
			name:      "text format with valid level",
			format:    LogFormatText,
			level:     LogLevelWarn,
			wantLevel: zerolog.WarnLevel,
		},
		{
			name:      "invalid format defaults to text",
			format:    "invalid",
			level:     LogLevelError,
			wantLevel: zerolog.ErrorLevel,
		},
		{
			name:      "invalid level defaults to info",
			format:    LogFormatJSON,
			level:     "invalid",
			wantLevel: zerolog.InfoLevel,
		},
		{
			name:      "trace level",
			format:    LogFormatJSON,
			level:     LogLevelTrace,
			wantLevel: zerolog.TraceLevel,
		},
		{
			name:      "info level",
			format:    LogFormatJSON,
			level:     LogLevelInfo,
			wantLevel: zerolog.InfoLevel,
		},
		{
			name:      "fatal level",
			format:    LogFormatJSON,
			level:     LogLevelFatal,
			wantLevel: zerolog.FatalLevel,
		},
		{
			name:      "panic level",
			format:    LogFormatJSON,
			level:     LogLevelPanic,
			wantLevel: zerolog.PanicLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore global level to prevent test pollution
			originalLevel := zerolog.GlobalLevel()
			t.Cleanup(func() {
				zerolog.SetGlobalLevel(originalLevel)
			})

			configureLogger(tt.format, tt.level)
			if zerolog.GlobalLevel() != tt.wantLevel {
				t.Errorf(
					"configureLogger(%q, %q) set level to %v, want %v",
					tt.format, tt.level, zerolog.GlobalLevel(), tt.wantLevel,
				)
			}
		})
	}
}

func TestConfigReloader_ReloadIfChanged(t *testing.T) {
	now := time.Now()
	prevStat := fakeFileInfo{name: "config.yaml", size: 100, mod: now}
	newStat := fakeFileInfo{name: "config.yaml", size: 200, mod: now.Add(time.Second)}
	validCfg := &config.Configuration{
		AccessControl: config.AccessControl{DefaultPolicy: "deny"},
	}

	t.Run("file not changed", func(t *testing.T) {
		mock := &mockConfigUpdater{}
		reloader := &configReloader{
			path:     "config.yaml",
			prevStat: prevStat,
			stat:     func(string) (os.FileInfo, error) { return prevStat, nil },
			load:     func(string) (*config.Configuration, error) { return validCfg, nil },
		}

		reloaded, err := reloader.reloadIfChanged(mock)
		if err != nil {
			t.Fatalf("reloadIfChanged() error = %v, want nil", err)
		}
		if reloaded {
			t.Error("reloadIfChanged() reloaded = true, want false")
		}
		if mock.called {
			t.Error("UpdateConfig() called = true, want false")
		}
	})

	t.Run("file changed with valid config", func(t *testing.T) {
		mock := &mockConfigUpdater{}
		reloader := &configReloader{
			path:     "config.yaml",
			prevStat: prevStat,
			stat:     func(string) (os.FileInfo, error) { return newStat, nil },
			load:     func(string) (*config.Configuration, error) { return validCfg, nil },
		}

		reloaded, err := reloader.reloadIfChanged(mock)
		if err != nil {
			t.Fatalf("reloadIfChanged() error = %v, want nil", err)
		}
		if !reloaded {
			t.Error("reloadIfChanged() reloaded = false, want true")
		}
		if !mock.called {
			t.Error("UpdateConfig() called = false, want true")
		}
	})

	t.Run("file changed with invalid config", func(t *testing.T) {
		mock := &mockConfigUpdater{}
		reloader := &configReloader{
			path:     "config.yaml",
			prevStat: prevStat,
			stat:     func(string) (os.FileInfo, error) { return newStat, nil },
			load: func(string) (*config.Configuration, error) {
				return nil, errors.New("invalid config")
			},
		}

		reloaded, err := reloader.reloadIfChanged(mock)
		if err == nil {
			t.Fatal("reloadIfChanged() error = nil, want error")
		}
		if reloaded {
			t.Error("reloadIfChanged() reloaded = true, want false")
		}
		if mock.called {
			t.Error("UpdateConfig() called = true, want false")
		}
	})

	t.Run("stat error", func(t *testing.T) {
		mock := &mockConfigUpdater{}
		reloader := &configReloader{
			path:     "config.yaml",
			prevStat: prevStat,
			stat: func(string) (os.FileInfo, error) {
				return nil, errors.New("stat error")
			},
			load: func(string) (*config.Configuration, error) { return validCfg, nil },
		}

		reloaded, err := reloader.reloadIfChanged(mock)
		if err == nil {
			t.Fatal("reloadIfChanged() error = nil, want error")
		}
		if reloaded {
			t.Error("reloadIfChanged() reloaded = true, want false")
		}
		if mock.called {
			t.Error("UpdateConfig() called = true, want false")
		}
	})
}

func TestStopServer(t *testing.T) {
	t.Run("shuts down server on context cancellation", func(t *testing.T) {
		mock := &mockServer{}
		testContextCancellation(t, func(ctx context.Context) {
			stopServer(ctx, mock)
		})
		if !mock.shutdownCalled {
			t.Error("Shutdown() called = false, want true")
		}
	})

	t.Run("logs error on shutdown failure", func(t *testing.T) {
		mock := &mockServer{shutdownErr: errors.New("shutdown error")}
		testContextCancellation(t, func(ctx context.Context) {
			stopServer(ctx, mock)
		})
		if !mock.shutdownCalled {
			t.Error("Shutdown() called = false, want true")
		}
	})
}

func TestRunEvery(t *testing.T) {
	t.Run("executes function on each tick", func(t *testing.T) {
		called := false
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})

		go func() {
			runEvery(ctx, testTickInterval, func() {
				called = true
				cancel()
			})
			close(done)
		}()

		select {
		case <-done:
			if !called {
				t.Errorf("called = false, want true")
			}
		case <-time.After(testTimeout):
			cancel()
			t.Fatal("runEvery did not complete in time")
		}
	})

	t.Run("stops when context is canceled", func(t *testing.T) {
		testContextCancellation(t, func(ctx context.Context) {
			runEvery(ctx, time.Hour, func() {})
		})
	})
}

func TestAutoUpdate(t *testing.T) {
	t.Run("stops when context is canceled", func(t *testing.T) {
		mock := &mockUpdater{}
		testContextCancellation(t, func(ctx context.Context) {
			autoUpdate(ctx, mock)
		})
	})
}

func TestLoadConfig(t *testing.T) {
	t.Run("loads valid config file", func(t *testing.T) {
		cfg, err := loadConfig("testdata/valid-config.yaml")
		if err != nil {
			t.Fatalf("loadConfig() error = %v, want nil", err)
		}
		if cfg == nil {
			t.Fatal("loadConfig() config = nil, want non-nil")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := loadConfig("/non/existent/file.yaml")
		if err == nil {
			t.Error("loadConfig() error = nil, want error")
		}
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		_, err := loadConfig("testdata/invalid-config.yaml")
		if err == nil {
			t.Error("loadConfig() error = nil, want error")
		}
	})
}

func TestNewConfigReloader(t *testing.T) {
	t.Run("creates reloader for existing file", func(t *testing.T) {
		reloader, err := newConfigReloader("testdata/valid-config.yaml")
		if err != nil {
			t.Fatalf("newConfigReloader() error = %v, want nil", err)
		}
		if reloader == nil {
			t.Fatal("newConfigReloader() reloader = nil, want non-nil")
		}
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := newConfigReloader("/non/existent/file.yaml")
		if err == nil {
			t.Error("newConfigReloader() error = nil, want error")
		}
	})
}

func TestAutoReload(t *testing.T) {
	t.Run("handles non-existent file gracefully", func(t *testing.T) {
		mock := &mockConfigUpdater{}
		done := make(chan struct{})

		go func() {
			// This should return promptly and close the done channel
			autoReload(context.Background(), mock, "/non/existent/file.yaml")
			close(done)
		}()

		select {
		case <-done:
			// The channel closed, meaning `autoReload` returned promptly
		case <-time.After(testTimeout):
			t.Fatal("autoReload did not return promptly for non-existent file")
		}
	})

	t.Run("stops when context is canceled", func(t *testing.T) {
		mock := &mockConfigUpdater{}
		testContextCancellation(t, func(ctx context.Context) {
			autoReload(ctx, mock, "testdata/valid-config.yaml")
		})
	})
}
