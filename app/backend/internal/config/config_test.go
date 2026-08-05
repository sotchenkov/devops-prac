package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse(mapLookup(nil))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.HTTPAddress != ":8080" {
		t.Errorf("HTTPAddress = %q, want :8080", cfg.HTTPAddress)
	}
	if cfg.Environment != "local" {
		t.Errorf("Environment = %q, want local", cfg.Environment)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
	}
	if cfg.StartupDelay != time.Second {
		t.Errorf("StartupDelay = %s, want 1s", cfg.StartupDelay)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Errorf("ShutdownTimeout = %s, want 10s", cfg.ShutdownTimeout)
	}
	if cfg.MetricsAuthEnabled {
		t.Error("MetricsAuthEnabled = true, want false")
	}
	if cfg.MetricsTokenFile != "" {
		t.Errorf("MetricsTokenFile = %q, want empty", cfg.MetricsTokenFile)
	}
	if cfg.SlogLevel() != slog.LevelInfo {
		t.Errorf("SlogLevel() = %s, want INFO", cfg.SlogLevel())
	}
}

func TestParseCustomValues(t *testing.T) {
	cfg, err := Parse(mapLookup(map[string]string{
		"HTTP_ADDRESS":         "127.0.0.1:9090",
		"APP_ENVIRONMENT":      "staging",
		"LOG_LEVEL":            "DEBUG",
		"STARTUP_DELAY":        "250ms",
		"SHUTDOWN_TIMEOUT":     "15s",
		"METRICS_AUTH_ENABLED": "true",
		"METRICS_TOKEN_FILE":   "/run/secrets/metrics-token",
	}))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.HTTPAddress != "127.0.0.1:9090" {
		t.Errorf("HTTPAddress = %q, want 127.0.0.1:9090", cfg.HTTPAddress)
	}
	if cfg.Environment != "staging" {
		t.Errorf("Environment = %q, want staging", cfg.Environment)
	}
	if cfg.LogLevel != "debug" || cfg.SlogLevel() != slog.LevelDebug {
		t.Errorf("log level = %q/%s, want debug/DEBUG", cfg.LogLevel, cfg.SlogLevel())
	}
	if cfg.StartupDelay != 250*time.Millisecond {
		t.Errorf("StartupDelay = %s, want 250ms", cfg.StartupDelay)
	}
	if cfg.ShutdownTimeout != 15*time.Second {
		t.Errorf("ShutdownTimeout = %s, want 15s", cfg.ShutdownTimeout)
	}
	if !cfg.MetricsAuthEnabled {
		t.Error("MetricsAuthEnabled = false, want true")
	}
	if cfg.MetricsTokenFile != "/run/secrets/metrics-token" {
		t.Errorf("MetricsTokenFile = %q, want /run/secrets/metrics-token", cfg.MetricsTokenFile)
	}
}

func TestParseRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		values      map[string]string
		wantInError string
	}{
		{
			name:        "empty environment",
			values:      map[string]string{"APP_ENVIRONMENT": ""},
			wantInError: "APP_ENVIRONMENT",
		},
		{
			name:        "unsupported log level",
			values:      map[string]string{"LOG_LEVEL": "verbose"},
			wantInError: "LOG_LEVEL",
		},
		{
			name:        "malformed startup delay",
			values:      map[string]string{"STARTUP_DELAY": "soon"},
			wantInError: "STARTUP_DELAY",
		},
		{
			name:        "negative startup delay",
			values:      map[string]string{"STARTUP_DELAY": "-1s"},
			wantInError: "STARTUP_DELAY",
		},
		{
			name:        "zero shutdown timeout",
			values:      map[string]string{"SHUTDOWN_TIMEOUT": "0s"},
			wantInError: "SHUTDOWN_TIMEOUT",
		},
		{
			name:        "invalid address",
			values:      map[string]string{"HTTP_ADDRESS": "8080"},
			wantInError: "HTTP_ADDRESS",
		},
		{
			name:        "invalid port",
			values:      map[string]string{"HTTP_ADDRESS": ":70000"},
			wantInError: "HTTP_ADDRESS",
		},
		{
			name:        "invalid metrics auth flag",
			values:      map[string]string{"METRICS_AUTH_ENABLED": "enabled"},
			wantInError: "METRICS_AUTH_ENABLED",
		},
		{
			name:        "metrics auth missing token file",
			values:      map[string]string{"METRICS_AUTH_ENABLED": "true"},
			wantInError: "METRICS_TOKEN_FILE",
		},
		{
			name:        "unused metrics token file",
			values:      map[string]string{"METRICS_TOKEN_FILE": "/run/secrets/metrics-token"},
			wantInError: "METRICS_TOKEN_FILE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(mapLookup(tt.values))
			if err == nil {
				t.Fatal("Parse() error = nil, want validation error")
			}
			if !strings.Contains(err.Error(), tt.wantInError) {
				t.Fatalf("Parse() error = %q, want it to contain %q", err, tt.wantInError)
			}
		})
	}
}

func mapLookup(values map[string]string) LookupEnv {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
