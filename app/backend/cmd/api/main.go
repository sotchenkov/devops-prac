package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sotchenkov/devops-prac/app/backend/internal/app"
	"github.com/sotchenkov/devops-prac/app/backend/internal/config"
	"github.com/sotchenkov/devops-prac/app/backend/internal/metricsauth"
)

var version = "dev"

func main() {
	bootstrapLogger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cfg, err := config.Load()
	if err != nil {
		bootstrapLogger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	var metricsVerifier *metricsauth.Verifier
	if cfg.MetricsAuthEnabled {
		metricsVerifier, err = metricsauth.Load(cfg.MetricsTokenFile)
		if err != nil {
			bootstrapLogger.Error("could not load metrics authentication credential", "error", err)
			os.Exit(1)
		}
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.SlogLevel(),
	}))

	instance, err := os.Hostname()
	if err != nil {
		logger.Warn("could not determine instance hostname", "error", err)
		instance = "unknown"
	}

	listener, err := net.Listen("tcp", cfg.HTTPAddress)
	if err != nil {
		logger.Error("could not open HTTP listener", "address", cfg.HTTPAddress, "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application := app.New(cfg, app.BuildInfo{Version: version}, instance, logger, metricsVerifier)
	logger.Info(
		"starting HTTP server",
		"address", listener.Addr().String(),
		"environment", cfg.Environment,
		"instance", instance,
		"version", version,
		"metrics_auth_enabled", cfg.MetricsAuthEnabled,
	)

	if err := application.Run(ctx, listener); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("HTTP server stopped with an error", "error", err)
		os.Exit(1)
	}

	logger.Info("HTTP server stopped")
}
