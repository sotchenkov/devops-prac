# App

A small stateless Go backend used as a workload for infrastructure and deployment exercises.

This application is LLM-generated.

## HTTP API

| Endpoint | Response | Purpose |
| --- | --- | --- |
| `GET /` | `200` JSON | Returns the instance hostname, build version, and environment |
| `GET /health/live` | `200` JSON | Confirms that the process can serve HTTP requests |
| `GET /health/ready` | `200` or `503` JSON | Reports the current startup or termination readiness state |
| `GET /version` | `200` JSON | Returns the version injected when the binary was built |
| `GET /metrics` | `200` text | Exposes metrics in the Prometheus text format |

Unsupported methods return `405`. Unknown paths return `404`.

## Readiness lifecycle

The application starts in the `starting` state. Its HTTP listener is available, but `/health/ready` returns `503` until `STARTUP_DELAY` has elapsed. It then transitions to `ready` and the endpoint returns `200`.

When the process receives `SIGINT` or `SIGTERM`, readiness changes to `terminating` before HTTP shutdown begins. The server stops accepting new connections and gives active requests up to `SHUTDOWN_TIMEOUT` to finish.

- successful drain: the process exits with code `0`;
- shutdown timeout: remaining connections are closed and the process exits with a non-zero code.

## Prometheus metrics

`GET /metrics` exposes:

- `app_build_info` — gauge with the running build version in the `version` label;
- `app_http_requests_total` — counter labeled by HTTP `method`, normalized `path`, and response `status`.

Known endpoints have stable path labels. All unknown paths use the single `not_found` label to avoid unbounded metric cardinality. Metrics are held in memory and reset when the process restarts.

## Configuration

Runtime configuration is supplied through environment variables. [`backend/config/app.env.example`](backend/config/app.env.example) contains local example values.

| Variable | Default | Validation |
| --- | --- | --- |
| `HTTP_ADDRESS` | `:8080` | Valid `host:port`, port `1-65535` |
| `APP_ENVIRONMENT` | `local` | Non-empty |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, or `error` |
| `STARTUP_DELAY` | `1s` | Non-negative Go duration |
| `SHUTDOWN_TIMEOUT` | `10s` | Positive Go duration |

Configuration is validated before the HTTP listener is opened.

## Run locally

From `app/backend`:

```shell
go test ./...
go run -ldflags="-X main.version=v1" ./cmd/api
```

## Container image

The multi-stage build produces a static binary and copies it into a minimal distroless image. The application runs as an unprivileged user and supports both `linux/amd64` and `linux/arm64` builds.

From `app/backend`:

```shell
docker build --build-arg VERSION=v1 -t app-backend:v1 .
docker run --rm -p 8080:8080 \
  --read-only \
  --cap-drop=ALL \
  --security-opt=no-new-privileges:true \
  app-backend:v1
```

The image does not define a Docker health check. An orchestrator should configure separate liveness and readiness probes using `/health/live` and `/health/ready`.
