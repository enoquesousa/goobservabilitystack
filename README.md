# goobservabilitystack

A production-ready observability telemetry reference implementation in Go, demonstrating the RED pattern (Rate, Errors, Duration) with Prometheus metrics and OpenTelemetry distributed tracing.

## Overview

`goobservabilitystack` is designed as a minimal, correct baseline for instrumenting Go HTTP services. It prioritises low CPU and memory overhead, a custom (non-global) Prometheus registry to prevent test pollution, and clean separation between transport, business logic, and telemetry layers.

## Architecture

```
cmd/server/
  main.go               — wires everything together; starts app + metrics HTTP servers

internal/
  metrics/
    registry.go         — custom Prometheus registry + metric definitions
    worker.go           — background goroutine that increments app_events_total every 5 s

tests/
  metrics/
    registry_test.go    — unit tests for registry construction and metric registration
    worker_test.go      — unit tests for worker tick behaviour and error paths
```

**Key design decisions:**

- A `prometheus.NewRegistry()` (not the default global) isolates metrics per process instance and prevents cross-test pollution.
- The `Worker` owns a single ticker goroutine; it stops cleanly via `context.Context` cancellation, preventing goroutine leaks.
- The metrics server (`/metrics`) is served on a dedicated port, separate from the application server, so scrape traffic never competes with application traffic.
- Both HTTP servers are configured with explicit `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` to prevent resource exhaustion.

## Prometheus Metrics

| Metric | Type | Labels | Description |
|---|---|---|---|
| `app_events_total` | Counter | — | Background worker events, incremented every 5 s |

Additional RED metrics (`http_requests_total`, `http_request_duration_seconds`) and a goroutine gauge (`go_goroutines_active`) are planned for the middleware layer.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `HTTP_PORT` | `8080` | Application HTTP server port |
| `METRICS_PORT` | `9091` | Prometheus `/metrics` scrape port |
| `JAEGER_ENDPOINT` | `localhost:4317` | Jaeger gRPC collector endpoint |
| `SERVICE_NAME` | `goobservabilitystack` | OpenTelemetry service name attribute |

## Prerequisites

- Go 1.22 or later
- Docker and Docker Compose (for the full observability stack)
- `golangci-lint` (for linting)

## Getting Started

### Run locally

```bash
# Download dependencies
make deps

# Build the binary
make build

# Run the server
./bin/goobservabilitystack
```

The application server listens on `http://localhost:8080` and the metrics endpoint on `http://localhost:9091/metrics`.

### Run with the full stack

```bash
# Start Go app + Prometheus + Jaeger + Grafana
make up

# Stop the stack
docker compose down
```

| Service | URL |
|---|---|
| Application | http://localhost:8080 |
| Metrics | http://localhost:9091/metrics |
| Prometheus | http://localhost:9090 |
| Jaeger UI | http://localhost:16686 |
| Grafana | http://localhost:3000 |

### Health check

```bash
curl http://localhost:8080/healthz
# HTTP 200
```

## Development

```bash
# Run all tests with the race detector
make test

# Run a single test package verbosely
go test ./tests/metrics/... -v

# Run a single test
go test ./tests/metrics/... -run TestNewRegistry -v

# Lint
make lint

# Audit dependencies for known vulnerabilities
govulncheck ./...
```

## Project Layout Conventions

- Every exported symbol carries a Godoc comment that begins with the symbol name and forms a complete sentence.
- Prometheus metric names follow strict conventions: `snake_case`, `_total` suffix for counters, `_seconds` or `_bytes` for unit suffixes.
- No global state: the Prometheus registry and all metrics are passed explicitly via constructors.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## License

This project is licensed under the MIT License.
