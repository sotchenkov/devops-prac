package metrics

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type requestKey struct {
	method string
	path   string
	status int
}

type Registry struct {
	mu       sync.Mutex
	requests map[requestKey]uint64
}

func New() *Registry {
	return &Registry{requests: make(map[requestKey]uint64)}
}

func (r *Registry) ObserveRequest(method, path string, status int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[requestKey{method: method, path: path, status: status}]++
}

func (r *Registry) WritePrometheus(w io.Writer, version string) {
	requests := r.snapshot()

	fmt.Fprintln(w, "# HELP app_build_info Build information for the running backend.")
	fmt.Fprintln(w, "# TYPE app_build_info gauge")
	fmt.Fprintf(w, "app_build_info{version=%s} 1\n", prometheusQuote(version))
	fmt.Fprintln(w, "# HELP app_http_requests_total Total HTTP requests handled by route and status.")
	fmt.Fprintln(w, "# TYPE app_http_requests_total counter")
	for _, request := range requests {
		fmt.Fprintf(
			w,
			"app_http_requests_total{method=%s,path=%s,status=%s} %d\n",
			prometheusQuote(request.method),
			prometheusQuote(request.path),
			prometheusQuote(strconv.Itoa(request.status)),
			request.count,
		)
	}
}

type requestCount struct {
	requestKey
	count uint64
}

func (r *Registry) snapshot() []requestCount {
	r.mu.Lock()
	defer r.mu.Unlock()

	requests := make([]requestCount, 0, len(r.requests))
	for key, count := range r.requests {
		requests = append(requests, requestCount{requestKey: key, count: count})
	}
	sort.Slice(requests, func(i, j int) bool {
		if requests[i].method != requests[j].method {
			return requests[i].method < requests[j].method
		}
		if requests[i].path != requests[j].path {
			return requests[i].path < requests[j].path
		}
		return requests[i].status < requests[j].status
	})
	return requests
}

func prometheusQuote(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return `"` + value + `"`
}
