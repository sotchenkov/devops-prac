package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/sotchenkov/devops-prac/app/backend/internal/config"
	"github.com/sotchenkov/devops-prac/app/backend/internal/health"
	"github.com/sotchenkov/devops-prac/app/backend/internal/httpapi"
	"github.com/sotchenkov/devops-prac/app/backend/internal/metrics"
)

type BuildInfo struct {
	Version string
}

type Application struct {
	cfg    config.Config
	health *health.State
	log    *slog.Logger
	server *http.Server
}

func New(cfg config.Config, build BuildInfo, instance string, logger *slog.Logger) *Application {
	healthState := health.New()
	registry := metrics.New()
	handler := httpapi.New(httpapi.Info{
		Environment: cfg.Environment,
		Instance:    instance,
		Version:     build.Version,
	}, healthState, registry)

	return &Application{
		cfg:    cfg,
		health: healthState,
		log:    logger,
		server: &http.Server{
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

func (a *Application) Run(ctx context.Context, listener net.Listener) error {
	startupCtx, cancelStartup := context.WithCancel(ctx)
	startupDone := make(chan struct{})
	go func() {
		defer close(startupDone)
		a.completeStartup(startupCtx)
	}()

	err := runHTTPServer(ctx, a.server, listener, a.health, a.cfg.ShutdownTimeout)
	cancelStartup()
	<-startupDone
	return err
}

func (a *Application) completeStartup(ctx context.Context) {
	timer := time.NewTimer(a.cfg.StartupDelay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		if a.health.MarkReady() {
			a.log.Info("application is ready", "startup_delay", a.cfg.StartupDelay.String())
		}
	}
}

func runHTTPServer(
	ctx context.Context,
	server *http.Server,
	listener net.Listener,
	healthState *health.State,
	shutdownTimeout time.Duration,
) error {
	serveErrors := make(chan error, 1)
	go func() {
		err := server.Serve(listener)
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serveErrors <- err
	}()

	select {
	case err := <-serveErrors:
		healthState.BeginTermination()
		return err
	case <-ctx.Done():
		healthState.BeginTermination()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		<-serveErrors
		return fmt.Errorf("graceful HTTP shutdown: %w", err)
	}

	return <-serveErrors
}
