package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultHTTPAddress     = ":8080"
	defaultEnvironment     = "local"
	defaultLogLevel        = "info"
	defaultStartupDelay    = time.Second
	defaultShutdownTimeout = 10 * time.Second
)

type Config struct {
	HTTPAddress     string
	Environment     string
	LogLevel        string
	StartupDelay    time.Duration
	ShutdownTimeout time.Duration
}

type LookupEnv func(string) (string, bool)

func Load() (Config, error) {
	return Parse(os.LookupEnv)
}

func Parse(lookup LookupEnv) (Config, error) {
	cfg := Config{
		HTTPAddress:     envOrDefault(lookup, "HTTP_ADDRESS", defaultHTTPAddress),
		Environment:     envOrDefault(lookup, "APP_ENVIRONMENT", defaultEnvironment),
		LogLevel:        strings.ToLower(envOrDefault(lookup, "LOG_LEVEL", defaultLogLevel)),
		StartupDelay:    defaultStartupDelay,
		ShutdownTimeout: defaultShutdownTimeout,
	}

	startupDelay, err := durationFromEnv(lookup, "STARTUP_DELAY", defaultStartupDelay)
	if err != nil {
		return Config{}, err
	}
	cfg.StartupDelay = startupDelay

	shutdownTimeout, err := durationFromEnv(lookup, "SHUTDOWN_TIMEOUT", defaultShutdownTimeout)
	if err != nil {
		return Config{}, err
	}
	cfg.ShutdownTimeout = shutdownTimeout

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Environment) == "" {
		return fmt.Errorf("APP_ENVIRONMENT must not be empty")
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("LOG_LEVEL %q is not supported; use debug, info, warn, or error", c.LogLevel)
	}

	if c.StartupDelay < 0 {
		return fmt.Errorf("STARTUP_DELAY must not be negative")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("SHUTDOWN_TIMEOUT must be greater than zero")
	}

	_, port, err := net.SplitHostPort(c.HTTPAddress)
	if err != nil {
		return fmt.Errorf("HTTP_ADDRESS %q must be in host:port form: %w", c.HTTPAddress, err)
	}
	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return fmt.Errorf("HTTP_ADDRESS %q contains an invalid TCP port", c.HTTPAddress)
	}

	return nil
}

func (c Config) SlogLevel() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func envOrDefault(lookup LookupEnv, key, fallback string) string {
	if value, ok := lookup(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func durationFromEnv(lookup LookupEnv, key string, fallback time.Duration) (time.Duration, error) {
	value, ok := lookup(key)
	if !ok {
		return fallback, nil
	}

	duration, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid Go duration: %w", key, err)
	}
	return duration, nil
}
