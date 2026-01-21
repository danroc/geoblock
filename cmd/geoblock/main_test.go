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

// Test doubles

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

type mockUpdater struct {
	err error
}

func (m *mockUpdater) Update() error {
	return m.err
}

type mockServer struct {
	shutdownCalled bool
	shutdownErr    error
}

func (m *mockServer) Shutdown(context.Context) error {
	m.shutdownCalled = true
	return m.shutdownErr
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
			t.Errorf("unexpected error: %v", err)
		}
		if reloaded {
			t.Error("should not report reloaded when file unchanged")
		}
		if mock.called {
			t.Error("UpdateConfig should not be called when file unchanged")
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
			t.Errorf("unexpected error: %v", err)
		}
		if !reloaded {
			t.Error("should report reloaded when file changed")
		}
		if !mock.called {
			t.Error("UpdateConfig should be called when file changed")
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
			t.Error("expected error for invalid config")
		}
		if reloaded {
			t.Error("should not report reloaded on error")
		}
		if mock.called {
			t.Error("UpdateConfig should not be called with invalid config")
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
			t.Error("expected error for stat failure")
		}
		if reloaded {
			t.Error("should not report reloaded on error")
		}
		if mock.called {
			t.Error("UpdateConfig should not be called when stat fails")
		}
	})

	t.Run("failed load can be retried", func(t *testing.T) {
		mock := &mockConfigUpdater{}
		loadAttempts := 0
		reloader := &configReloader{
			path:     "config.yaml",
			prevStat: prevStat,
			stat:     func(string) (os.FileInfo, error) { return newStat, nil },
			load: func(string) (*config.Configuration, error) {
				loadAttempts++
				if loadAttempts == 1 {
					return nil, errors.New("transient error")
				}
				return validCfg, nil
			},
		}

		// First attempt fails
		reloaded, err := reloader.reloadIfChanged(mock)
		if err == nil {
			t.Error("expected error on first attempt")
		}
		if reloaded || mock.called {
			t.Error("should not reload on error")
		}

		// Second attempt succeeds (prevStat was not updated after failure)
		reloaded, err = reloader.reloadIfChanged(mock)
		if err != nil {
			t.Errorf("unexpected error on retry: %v", err)
		}
		if !reloaded {
			t.Error("should report reloaded on successful retry")
		}
		if !mock.called {
			t.Error("UpdateConfig should be called on successful retry")
		}
	})
}

func TestStopServer(t *testing.T) {
	t.Run("shuts down server on context cancellation", func(t *testing.T) {
		mock := &mockServer{}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})

		go func() {
			stopServer(ctx, mock)
			close(done)
		}()

		cancel()

		select {
		case <-done:
			if !mock.shutdownCalled {
				t.Error("Shutdown should be called")
			}
		case <-time.After(time.Second):
			t.Error("stopServer did not return after context cancellation")
		}
	})

	t.Run("logs error on shutdown failure", func(t *testing.T) {
		mock := &mockServer{shutdownErr: errors.New("shutdown error")}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})

		go func() {
			stopServer(ctx, mock)
			close(done)
		}()

		cancel()

		select {
		case <-done:
			if !mock.shutdownCalled {
				t.Error("Shutdown should be called even if it returns error")
			}
		case <-time.After(time.Second):
			t.Error("stopServer did not return after context cancellation")
		}
	})
}

func TestAutoUpdate_ContextCancellation(t *testing.T) {
	mock := &mockUpdater{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		autoUpdate(ctx, mock)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Success: autoUpdate returned after context cancellation
	case <-time.After(time.Second):
		t.Error("autoUpdate did not return after context cancellation")
	}
}

func TestAutoReload_ContextCancellation(t *testing.T) {
	mock := &mockConfigUpdater{}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		// Use a nonexistent filepath so `newConfigReloader` fails early and the
		// function returns immediately (testing early exit path).
		autoReload(ctx, mock, "nonexistent-file-for-test")
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Success: autoReload returned
	case <-time.After(time.Second):
		t.Error("autoReload did not return after context cancellation")
	}
}
