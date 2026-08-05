package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/sotchenkov/devops-prac/app/backend/internal/health"
	"github.com/sotchenkov/devops-prac/app/backend/internal/metrics"
	"github.com/sotchenkov/devops-prac/app/backend/internal/metricsauth"
)

const prometheusContentType = "text/plain; version=0.0.4; charset=utf-8"

type Info struct {
	Environment string
	Instance    string
	Version     string
}

type API struct {
	health  *health.State
	info    Info
	metrics *metrics.Registry
	auth    *metricsauth.Verifier
}

func New(info Info, healthState *health.State, registry *metrics.Registry, auth *metricsauth.Verifier) http.Handler {
	api := &API{health: healthState, info: info, metrics: registry, auth: auth}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", api.root)
	mux.HandleFunc("GET /health/live", api.live)
	mux.HandleFunc("GET /health/ready", api.ready)
	mux.HandleFunc("GET /version", api.version)
	mux.HandleFunc("GET /metrics", api.prometheusMetrics)

	return api.observeRequests(mux)
}

func (a *API) root(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"environment": a.info.Environment,
		"instance":    a.info.Instance,
		"version":     a.info.Version,
	})
}

func (a *API) live(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "live"})
}

func (a *API) ready(w http.ResponseWriter, _ *http.Request) {
	ready, phase := a.health.Snapshot()
	status := http.StatusServiceUnavailable
	if ready {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]string{"status": string(phase)})
}

func (a *API) version(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": a.info.Version})
}

func (a *API) prometheusMetrics(w http.ResponseWriter, r *http.Request) {
	if a.auth != nil && !a.auth.Authorize(r.Header.Get("Authorization")) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", prometheusContentType)
	w.WriteHeader(http.StatusOK)
	a.metrics.WritePrometheus(w, a.info.Version)
}

func (a *API) observeRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		a.metrics.ObserveRequest(r.Method, normalizedPath(r.URL.Path), recorder.statusCode())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.WriteHeader(http.StatusOK)
	}
	return r.ResponseWriter.Write(body)
}

func (r *statusRecorder) statusCode() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func normalizedPath(path string) string {
	switch path {
	case "/", "/health/live", "/health/ready", "/version", "/metrics":
		return path
	default:
		return "not_found"
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
