# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned

- RED middleware (`http_requests_total` counter, `http_request_duration_seconds` histogram) wrapping `http.Handler`.
- OpenTelemetry tracing layer with Jaeger gRPC exporter and stdout fallback for development.
- `go_goroutines_active` gauge updated by the background worker.
- Docker Compose stack (Prometheus, Jaeger, Grafana) with provisioned dashboards.
- GitHub Actions CI pipeline (build, test, lint, govulncheck).

---

## [0.2.0] — 2025-01-01

### Changed

- Bumped `google.golang.org/protobuf` to `v1.33.0` to resolve a known vulnerability (CVE addressed upstream).

---

## [0.1.0] — 2025-01-01

### Added

- `internal/metrics` package:
  - `Registry` — wraps a custom `prometheus.Registry` (not the global default) with the `app_events_total` counter registered at construction time.
  - `Worker` — increments a `prometheus.Counter` on a configurable fixed interval; stops cleanly via `context.Context` cancellation.
- `cmd/server/main.go`:
  - Application HTTP server on `HTTP_PORT` (default `8080`) with `/healthz` health check endpoint.
  - Dedicated metrics HTTP server on `METRICS_PORT` (default `9091`) exposing `/metrics` for Prometheus scraping.
  - Graceful shutdown with a 10-second drain timeout on `SIGTERM` / `SIGINT`.
  - Explicit `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` on both servers.
- `tests/metrics` package:
  - `TestNewRegistry` — verifies registry construction and `app_events_total` presence.
  - `TestWorker` — verifies the counter increments at least twice within two tick intervals.
  - `TestNewWorker_Errors` — verifies validation of nil counter and non-positive interval.
- `Makefile` targets: `deps`, `build`, `test`, `lint`, `up`.
- `SECURITY.md`, `CONTRIBUTING.md`, `CHANGELOG.md`, `README.md`.
