package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sotchenkov/devops-prac/app/backend/internal/health"
	"github.com/sotchenkov/devops-prac/app/backend/internal/metrics"
	"github.com/sotchenkov/devops-prac/app/backend/internal/metricsauth"
)

func TestOperationalEndpoints(t *testing.T) {
	healthState := health.New()
	handler := New(Info{
		Environment: "test",
		Instance:    "instance-1",
		Version:     "v1.2.3",
	}, healthState, metrics.New(), nil)

	tests := []struct {
		name        string
		path        string
		wantStatus  int
		contentType string
		wantBody    string
	}{
		{
			name:        "root",
			path:        "/",
			wantStatus:  http.StatusOK,
			contentType: "application/json",
			wantBody:    `{"environment":"test","instance":"instance-1","version":"v1.2.3"}`,
		},
		{
			name:        "liveness",
			path:        "/health/live",
			wantStatus:  http.StatusOK,
			contentType: "application/json",
			wantBody:    `"status":"live"`,
		},
		{
			name:        "readiness while starting",
			path:        "/health/ready",
			wantStatus:  http.StatusServiceUnavailable,
			contentType: "application/json",
			wantBody:    `"status":"starting"`,
		},
		{
			name:        "version",
			path:        "/version",
			wantStatus:  http.StatusOK,
			contentType: "application/json",
			wantBody:    `"version":"v1.2.3"`,
		},
		{
			name:        "metrics",
			path:        "/metrics",
			wantStatus:  http.StatusOK,
			contentType: prometheusContentType,
			wantBody:    `app_build_info{version="v1.2.3"} 1`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if contentType := response.Header().Get("Content-Type"); contentType != tt.contentType {
				t.Errorf("Content-Type = %q, want %q", contentType, tt.contentType)
			}
			if !strings.Contains(response.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", response.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestReadinessTransitions(t *testing.T) {
	healthState := health.New()
	handler := New(Info{Version: "test"}, healthState, metrics.New(), nil)

	assertReadyStatus(t, handler, http.StatusServiceUnavailable, "starting")
	healthState.MarkReady()
	assertReadyStatus(t, handler, http.StatusOK, "ready")
	healthState.BeginTermination()
	assertReadyStatus(t, handler, http.StatusServiceUnavailable, "terminating")
}

func TestRequestMetricsUseBoundedRouteLabels(t *testing.T) {
	registry := metrics.New()
	handler := New(Info{Version: "test"}, health.New(), registry, nil)

	for _, path := range []string{"/", "/does-not-exist"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	body := response.Body.String()

	for _, want := range []string{
		`app_http_requests_total{method="GET",path="/",status="200"} 1`,
		`app_http_requests_total{method="GET",path="not_found",status="404"} 1`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("metrics body does not contain %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "/does-not-exist") {
		t.Errorf("metrics body contains an unbounded raw path:\n%s", body)
	}
}

func TestUnsupportedMethodReturnsMethodNotAllowed(t *testing.T) {
	handler := New(Info{Version: "test"}, health.New(), metrics.New(), nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/", nil))

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
	}
}

func TestMetricsBearerAuthentication(t *testing.T) {
	verifier, err := metricsauth.New("expected-token")
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}
	handler := New(Info{Version: "test"}, health.New(), metrics.New(), verifier)

	tests := []struct {
		name          string
		authorization string
		wantStatus    int
		wantChallenge bool
		wantMetrics   bool
	}{
		{name: "missing token", wantStatus: http.StatusUnauthorized, wantChallenge: true},
		{name: "wrong scheme", authorization: "Basic expected-token", wantStatus: http.StatusUnauthorized, wantChallenge: true},
		{name: "wrong token", authorization: "Bearer wrong-token", wantStatus: http.StatusUnauthorized, wantChallenge: true},
		{name: "correct token", authorization: "Bearer expected-token", wantStatus: http.StatusOK, wantMetrics: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
			if tt.authorization != "" {
				request.Header.Set("Authorization", tt.authorization)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if got := response.Header().Get("WWW-Authenticate"); (got == "Bearer") != tt.wantChallenge {
				t.Errorf("WWW-Authenticate = %q, challenge wanted = %t", got, tt.wantChallenge)
			}
			containsMetrics := strings.Contains(response.Body.String(), "app_build_info")
			if containsMetrics != tt.wantMetrics {
				t.Errorf("metrics body present = %t, want %t", containsMetrics, tt.wantMetrics)
			}
		})
	}
}

func TestMetricsAuthenticationDoesNotProtectOtherEndpoints(t *testing.T) {
	verifier, err := metricsauth.New("expected-token")
	if err != nil {
		t.Fatalf("create verifier: %v", err)
	}
	healthState := health.New()
	healthState.MarkReady()
	handler := New(Info{Version: "test"}, healthState, metrics.New(), verifier)

	for _, path := range []string{"/", "/health/live", "/health/ready", "/version"} {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
		})
	}
}

func assertReadyStatus(t *testing.T, handler http.Handler, wantStatus int, wantPhase string) {
	t.Helper()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/health/ready", nil))
	if response.Code != wantStatus {
		t.Fatalf("status = %d, want %d", response.Code, wantStatus)
	}

	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode readiness response: %v", err)
	}
	if body["status"] != wantPhase {
		t.Fatalf("phase = %q, want %q", body["status"], wantPhase)
	}
}
