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
	testTimeout      = 1 * time.Second
	testTickInterval = 10 * time.Millisecond
)

// Test helpers

// testContextCancellation runs fn in a goroutine and verifies it returns promptly after
// ctx is canceled.
func testContextCancellation(t *testing.T, fn func(ctx context.Context)) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		fn(ctx)

		// The done channel is closed to signal that fn has returned, which should
		// happen after the context cancellation.
		close(done)
	}()

	// Cancel the context to trigger fn to return
	cancel()

	select {
	case <-done:
		// Success, fn returned after context cancellation
	case <-time.After(testTimeout):
		t.Fatal("function did not complete in time after context cancellation")
	}
}

// Test doubles

// fakeFileInfo implements os.FileInfo for testing file stat comparisons.
type fakeFileInfo struct {
	size int64
	mod  time.Time
}

func (f fakeFileInfo) Name() string       { return "" }
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

// mockUpdater implements the Updater interface for testing.
type mockUpdater struct{}

func (m *mockUpdater) Update(_ context.Context) error {
	return nil
}

// nopConfigReloadCollector is a no-op collector for testing.
type nopConfigReloadCollector struct{}

func (nopConfigReloadCollector) RecordConfigReload(_ bool, _ int) {}

func TestEnvOrDefault(t *testing.T) {
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
			got := envOrDefault(tt.key, tt.fallback)
			if got != tt.want {
				t.Errorf(
					"envOrDefault(%q, %q) = %q, want %q",
					tt.key, tt.fallback, got, tt.want,
				)
			}
		})
	}
}

func TestLoadOptions(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		want    *appOptions
	}{
		{
			name: "all custom values",
			envVars: map[string]string{
				optionConfigPath: "/tmp/test.yaml",
				optionServerPort: "1234",
				optionLogLevel:   logLevelDebug,
				optionLogFormat:  logFormatText,
			},
			want: &appOptions{
				configPath: "/tmp/test.yaml",
				serverPort: "1234",
				logLevel:   logLevelDebug,
				logFormat:  logFormatText,
			},
		},
		{
			name:    "default values",
			envVars: map[string]string{},
			want: &appOptions{
				configPath: defaultConfigPath,
				serverPort: defaultServerPort,
				logLevel:   defaultLogLevel,
				logFormat:  defaultLogFormat,
			},
		},
		{
			name: "mixed values",
			envVars: map[string]string{
				optionConfigPath: "/custom/config.yaml",
				optionLogLevel:   logLevelDebug,
			},
			want: &appOptions{
				configPath: "/custom/config.yaml",
				serverPort: defaultServerPort,
				logLevel:   logLevelDebug,
				logFormat:  defaultLogFormat,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			got := loadOptions()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadOptions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    zerolog.Level
		wantErr bool
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
			if parsed != tt.want {
				t.Errorf(
					"parseLogLevel(%q) = %v, want %v",
					tt.input,
					parsed,
					tt.want,
				)
			}
			if err != nil && !tt.wantErr {
				t.Errorf("parseLogLevel(%q) unexpected error: %v", tt.input, err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("parseLogLevel(%q) expected error, got nil", tt.input)
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
			format:    logFormatJSON,
			level:     logLevelDebug,
			wantLevel: zerolog.DebugLevel,
		},
		{
			name:      "text format with valid level",
			format:    logFormatText,
			level:     logLevelWarn,
			wantLevel: zerolog.WarnLevel,
		},
		{
			name:      "invalid format defaults to text",
			format:    "invalid",
			level:     logLevelError,
			wantLevel: zerolog.ErrorLevel,
		},
		{
			name:      "invalid level defaults to info",
			format:    logFormatJSON,
			level:     "invalid",
			wantLevel: zerolog.InfoLevel,
		},
		{
			name:      "trace level",
			format:    logFormatJSON,
			level:     logLevelTrace,
			wantLevel: zerolog.TraceLevel,
		},
		{
			name:      "info level",
			format:    logFormatJSON,
			level:     logLevelInfo,
			wantLevel: zerolog.InfoLevel,
		},
		{
			name:      "fatal level",
			format:    logFormatJSON,
			level:     logLevelFatal,
			wantLevel: zerolog.FatalLevel,
		},
		{
			name:      "panic level",
			format:    logFormatJSON,
			level:     logLevelPanic,
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
	prevStat := fakeFileInfo{size: 100, mod: now}
	newStat := fakeFileInfo{size: 200, mod: now.Add(time.Second)}
	validCfg := &config.Configuration{
		AccessControl: config.AccessControl{
			DefaultPolicy: "deny",
			Rules: []config.AccessControlRule{
				{Policy: "allow", Domains: []string{"example.com"}},
				{Policy: "deny", Domains: []string{"blocked.com"}},
			},
		},
	}

	tests := []struct {
		name           string
		stat           func(string) (os.FileInfo, error)
		load           func(string) (*config.Configuration, error)
		wantReload     bool
		wantErr        bool
		wantCalled     bool
		wantRulesCount int
	}{
		{
			name:           "file not changed",
			stat:           func(string) (os.FileInfo, error) { return prevStat, nil },
			load:           func(string) (*config.Configuration, error) { return validCfg, nil },
			wantReload:     false,
			wantErr:        false,
			wantCalled:     false,
			wantRulesCount: 0,
		},
		{
			name:           "file changed with valid config",
			stat:           func(string) (os.FileInfo, error) { return newStat, nil },
			load:           func(string) (*config.Configuration, error) { return validCfg, nil },
			wantReload:     true,
			wantErr:        false,
			wantCalled:     true,
			wantRulesCount: 2,
		},
		{
			name: "file changed with invalid config",
			stat: func(string) (os.FileInfo, error) { return newStat, nil },
			load: func(string) (*config.Configuration, error) {
				return nil, errors.New("invalid config")
			},
			wantReload:     false,
			wantErr:        true,
			wantCalled:     false,
			wantRulesCount: 0,
		},
		{
			name: "stat error",
			stat: func(string) (os.FileInfo, error) {
				return nil, errors.New("stat error")
			},
			load:           func(string) (*config.Configuration, error) { return validCfg, nil },
			wantReload:     false,
			wantErr:        true,
			wantCalled:     false,
			wantRulesCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockConfigUpdater{}
			reloader := &configReloader{
				path:     "config.yaml",
				prevStat: prevStat,
				stat:     tt.stat,
				load:     tt.load,
			}

			reloaded, rulesCount, err := reloader.reloadIfChanged(mock)
			if err != nil && !tt.wantErr {
				t.Errorf("reloadIfChanged() unexpected error: %v", err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("reloadIfChanged() expected error, got nil")
			}
			if reloaded != tt.wantReload {
				t.Errorf(
					"reloadIfChanged() reloaded = %v, want %v",
					reloaded,
					tt.wantReload,
				)
			}
			if rulesCount != tt.wantRulesCount {
				t.Errorf(
					"reloadIfChanged() rulesCount = %v, want %v",
					rulesCount,
					tt.wantRulesCount,
				)
			}
			if mock.called != tt.wantCalled {
				t.Errorf(
					"UpdateConfig() called = %v, want %v",
					mock.called,
					tt.wantCalled,
				)
			}
		})
	}
}

func TestStopServer(t *testing.T) {
	t.Run("shuts down server on context cancellation", func(t *testing.T) {
		mock := &mockServer{}
		testContextCancellation(t, func(ctx context.Context) {
			stopServer(ctx, mock)
		})

		if !mock.shutdownCalled {
			t.Errorf("Shutdown() called = %v, want %v", mock.shutdownCalled, true)
		}
	})

	t.Run("logs error on shutdown failure", func(t *testing.T) {
		mock := &mockServer{shutdownErr: errors.New("shutdown error")}
		testContextCancellation(t, func(ctx context.Context) {
			stopServer(ctx, mock)
		})

		if !mock.shutdownCalled {
			t.Errorf("Shutdown() called = %v, want %v", mock.shutdownCalled, true)
		}
	})
}

func TestRunEvery(t *testing.T) {
	t.Run("executes function on each tick", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		callCount := 0
		wantCallCount := 3
		done := make(chan struct{})

		go func() {
			runEvery(ctx, testTickInterval, func() {
				callCount++
				if callCount == wantCallCount {
					cancel()
				}
			})
			close(done)
		}()

		select {
		case <-done:
			if callCount != wantCallCount {
				t.Errorf("callCount = %d, want %d", callCount, wantCallCount)
			}
		case <-time.After(testTimeout):
			t.Fatal("function did not complete in time")
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
	tests := []struct {
		name    string
		path    string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "loads valid config file",
			path:    "testdata/valid-config.yaml",
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "returns error for non-existent file",
			path:    "/non/existent/file.yaml",
			wantNil: true,
			wantErr: true,
		},
		{
			name:    "returns error for invalid YAML",
			path:    "testdata/invalid-config.yaml",
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := loadConfig(tt.path)
			if err != nil && !tt.wantErr {
				t.Errorf("loadConfig(%q) unexpected error: %v", tt.path, err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("loadConfig(%q) expected error, got nil", tt.path)
			}
			if tt.wantNil && cfg != nil {
				t.Errorf("loadConfig(%q) = %v, want nil", tt.path, cfg)
			}
			if !tt.wantNil && cfg == nil {
				t.Errorf("loadConfig(%q) = nil, want non-nil", tt.path)
			}
		})
	}
}

func TestNewConfigReloader(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "creates reloader for existing file",
			path:    "testdata/valid-config.yaml",
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "returns error for non-existent file",
			path:    "/non/existent/file.yaml",
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reloader, err := newConfigReloader(tt.path)
			if err != nil && !tt.wantErr {
				t.Errorf("newConfigReloader(%q) unexpected error: %v", tt.path, err)
			}
			if err == nil && tt.wantErr {
				t.Errorf("newConfigReloader(%q) expected error, got nil", tt.path)
			}
			if tt.wantNil && reloader != nil {
				t.Errorf("newConfigReloader(%q) = %v, want nil", tt.path, reloader)
			}
			if !tt.wantNil && reloader == nil {
				t.Errorf("newConfigReloader(%q) = nil, want non-nil", tt.path)
			}
		})
	}
}

func TestCacheDir(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		setEnv bool
		want   string
	}{
		{
			name:   "returns env value when set",
			envVal: "/custom/cache",
			setEnv: true,
			want:   "/custom/cache",
		},
		{
			name: "returns default when env not set",
			want: defaultCacheDir,
		},
		{
			name:   "returns empty when env set to empty",
			envVal: "",
			setEnv: true,
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(optionCacheDir, tt.envVal)
			}
			if got := cacheDir(); got != tt.want {
				t.Errorf("cacheDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAutoReload(t *testing.T) {
	t.Run("handles non-existent file gracefully", func(t *testing.T) {
		mock := &mockConfigUpdater{}
		done := make(chan struct{})

		go func() {
			autoReload(
				context.Background(),
				mock,
				"testdata/non-existent-config.yaml",
				nopConfigReloadCollector{},
			)
			close(done)
		}()

		select {
		case <-done:
			// autoReload returned promptly after failing to load the file
		case <-time.After(testTimeout):
			t.Fatal("function did not complete in time")
		}
	})

	t.Run("stops when context is canceled", func(t *testing.T) {
		mock := &mockConfigUpdater{}
		testContextCancellation(t, func(ctx context.Context) {
			autoReload(
				ctx,
				mock,
				"testdata/valid-config.yaml",
				nopConfigReloadCollector{},
			)
		})
	})
}
