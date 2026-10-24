# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Security

- Native application and metrics listeners default to loopback and validate host/port configuration before startup.
- Demo port mappings bind to loopback, Grafana requires an operator-supplied password, and credential files are excluded from image builds.
- OTLP defaults to TLS, supports CA and mTLS configuration through standard environment variables, and rejects invalid or contradictory transport settings before SDK initialization.
- Collector headers require TLS and malformed configuration errors do not include credential values.
- CI uses read permissions by default; package publication is limited to the deploy job. Failed PR checks no longer close or comment on the PR automatically.

### Planned

- Configurable OTel sampling rate via environment variable.
- Multi-architecture Docker image published to ghcr.io.

---

## [1.0.0] — 2025-01-01

### Added

- `internal/middleware` package:
  - `Config` — injectable dependency struct for the RED middleware (counter, histogram, tracer).
  - `RED` — HTTP middleware implementing Rate, Errors, and Duration instrumentation via `http_requests_total` and `http_request_duration_seconds`; wraps each request in an OTel server span and injects W3C `traceparent` / `tracestate` headers into the response.
  - `statusRecorder` — minimalist `http.ResponseWriter` wrapper capturing status codes written by handlers.
- `internal/tracing` package:
  - `New` — constructs an OTel `TracerProvider` with OTLP gRPC exporter. Returns a shutdown function for clean span flushing.
- Additions to `internal/metrics`:
  - `process_duration_seconds` Histogram — simulated processing latency with randomized durations.
  - `go_goroutines_active` Gauge — current number of active goroutines in the system.
  - `http_requests_total` Counter vec — RED rate/error counter labeled by `method`, `route`, `status_code`.
  - `http_request_duration_seconds` Histogram vec — RED duration histogram labeled by `method`, `route`.
- Additions to `cmd/server/main.go`:
  - RED middleware connected to the application's `http.ServeMux`.
  - OTel `TracerProvider` initialized at startup.
- `deploy/` directory:
  - `deploy/prometheus/prometheus.yml` — Prometheus scrape configuration.
  - `deploy/grafana/provisioning/` — pre-provisioned Prometheus datasource and dashboard.
- `docker-compose.yml` — orchestrates `app`, `prometheus`, `jaeger`, and `grafana`.
- `Dockerfile` — multi-stage build; final image runs as a non-root user.
- `tests/middleware/red_integration_test.go` — integration tests for the RED middleware.

### Changed

- `Worker` now emits an OTel span per cycle in addition to incrementing the event counter.
- `Registry.NewRegistry` registers the full set of RED metrics upon construction.
- `Makefile`'s `up` target now invokes `docker compose up --build`.

---

## [0.2.0] — 2025-01-01

### Security

- Updated `google.golang.org/protobuf` to resolve a known upstream vulnerability.

---

## [0.1.0] — 2025-01-01

### Added

- `internal/metrics` package:
  - `Registry` — encapsulates a custom `prometheus.Registry` with `app_events_total` registered.
  - `Worker` — increments a Prometheus counter at a fixed interval.
- `cmd/server/main.go`:
  - Application HTTP server on port `8080`.
  - Dedicated metrics HTTP server on port `9091`.
  - Graceful shutdown with a 10-second timeout.
- `tests/metrics` package with unit tests for the registry and worker.
- `Makefile` with targets: `deps`, `build`, `test`, `lint`, `up`.
