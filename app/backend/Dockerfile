FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.26.5-alpine3.23@sha256:622e56dbc11a8cfe87cafa2331e9a201877271cbff918af53d3be315f3da88cc AS build

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=dev

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /usr/local/bin/app \
    ./cmd/api

FROM gcr.io/distroless/static-debian13:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6

ARG VERSION=dev

LABEL org.opencontainers.image.title="app-backend" \
      org.opencontainers.image.description="Demo Go backend for DevOps tooling exercises" \
      org.opencontainers.image.source="https://github.com/sotchenkov/devops-prac" \
      org.opencontainers.image.version="${VERSION}"

COPY --from=build --chown=65532:65532 /usr/local/bin/app /app

EXPOSE 8080/tcp

USER 65532:65532

ENTRYPOINT ["/app"]
