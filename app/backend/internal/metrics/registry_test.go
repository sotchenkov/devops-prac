package metrics

import (
	"strings"
	"testing"
)

func TestWritePrometheusIsDeterministicAndEscapesLabels(t *testing.T) {
	registry := New()
	registry.ObserveRequest("GET", "/version", 200)
	registry.ObserveRequest("GET", "/", 200)
	registry.ObserveRequest("GET", "/", 200)

	var output strings.Builder
	registry.WritePrometheus(&output, "v1\nlocal")

	want := []string{
		`app_build_info{version="v1\nlocal"} 1`,
		`app_http_requests_total{method="GET",path="/",status="200"} 2`,
		`app_http_requests_total{method="GET",path="/version",status="200"} 1`,
	}
	for _, line := range want {
		if !strings.Contains(output.String(), line) {
			t.Errorf("output does not contain %q:\n%s", line, output.String())
		}
	}
	if strings.Index(output.String(), `path="/"`) > strings.Index(output.String(), `path="/version"`) {
		t.Errorf("request metrics are not sorted:\n%s", output.String())
	}
}
