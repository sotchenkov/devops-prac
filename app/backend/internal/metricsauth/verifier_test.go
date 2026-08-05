package metricsauth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAcceptsOneOptionalTrailingLineEnding(t *testing.T) {
	for _, tt := range []struct {
		name    string
		content string
	}{
		{name: "no line ending", content: "expected-token"},
		{name: "LF", content: "expected-token\n"},
		{name: "CRLF", content: "expected-token\r\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTokenFile(t, tt.content)
			verifier, err := Load(path)
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if !verifier.Authorize("Bearer expected-token") {
				t.Fatal("loaded verifier rejected expected token")
			}
		})
	}
}

func TestLoadRejectsInvalidOrUnreadableTokenFile(t *testing.T) {
	const secret = "do-not-expose-this-token"

	tests := []struct {
		name        string
		path        func(*testing.T) string
		wantInError string
	}{
		{
			name: "missing file",
			path: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing-token")
			},
			wantInError: "read metrics bearer token file",
		},
		{
			name: "path is directory",
			path: func(t *testing.T) string {
				return t.TempDir()
			},
			wantInError: "read metrics bearer token file",
		},
		{
			name: "empty file",
			path: func(t *testing.T) string {
				return writeTokenFile(t, "")
			},
			wantInError: "must not be empty",
		},
		{
			name: "line ending only",
			path: func(t *testing.T) string {
				return writeTokenFile(t, "\n")
			},
			wantInError: "must not be empty",
		},
		{
			name: "multiple lines",
			path: func(t *testing.T) string {
				return writeTokenFile(t, secret+"\nsecond-line\n")
			},
			wantInError: "without whitespace",
		},
		{
			name: "surrounding spaces",
			path: func(t *testing.T) string {
				return writeTokenFile(t, " "+secret+" ")
			},
			wantInError: "without whitespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(tt.path(t))
			if err == nil {
				t.Fatal("Load() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tt.wantInError) {
				t.Fatalf("Load() error = %q, want it to contain %q", err, tt.wantInError)
			}
			if strings.Contains(err.Error(), secret) {
				t.Fatalf("Load() error exposes token content: %q", err)
			}
		})
	}
}

func TestAuthorizeRequiresOneBearerCredential(t *testing.T) {
	verifier, err := New("expected-token")
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for _, tt := range []struct {
		name   string
		header string
		want   bool
	}{
		{name: "missing", header: "", want: false},
		{name: "wrong scheme", header: "Basic expected-token", want: false},
		{name: "wrong token", header: "Bearer wrong-token", want: false},
		{name: "extra separator", header: "Bearer  expected-token", want: false},
		{name: "extra credential", header: "Bearer expected-token extra", want: false},
		{name: "bearer", header: "Bearer expected-token", want: true},
		{name: "case insensitive scheme", header: "bearer expected-token", want: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := verifier.Authorize(tt.header); got != tt.want {
				t.Fatalf("Authorize() = %t, want %t", got, tt.want)
			}
		})
	}
}

func writeTokenFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "metrics-token")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write token file: %v", err)
	}
	return path
}
