package app

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/sotchenkov/devops-prac/app/backend/internal/health"
)

func TestRunHTTPServerDrainsActiveRequest(t *testing.T) {
	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(requestStarted)
		<-releaseRequest
		w.WriteHeader(http.StatusNoContent)
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	healthState := health.New()
	healthState.MarkReady()
	server := &http.Server{Handler: handler}
	ctx, cancel := context.WithCancel(context.Background())
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- runHTTPServer(ctx, server, listener, healthState, 2*time.Second)
	}()

	client := &http.Client{Timeout: 3 * time.Second}
	requestDone := make(chan error, 1)
	go func() {
		response, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			err = response.Body.Close()
		}
		requestDone <- err
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		t.Fatal("request did not reach the handler")
	}

	cancel()
	assertEventuallyNotReady(t, healthState)
	close(releaseRequest)

	if err := <-requestDone; err != nil {
		t.Fatalf("active request failed during shutdown: %v", err)
	}
	if err := <-serverDone; err != nil {
		t.Fatalf("runHTTPServer() error = %v", err)
	}
}

func assertEventuallyNotReady(t *testing.T, state *health.State) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		ready, phase := state.Snapshot()
		if !ready && phase == health.PhaseTerminating {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("health state did not transition to terminating")
}
