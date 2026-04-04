# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

`goobservabilitystack` is a production-ready observability telemetry system built in Go, demonstrating Prometheus metrics, OpenTelemetry distributed tracing, and the RED (Rate, Errors, Duration) pattern. It is designed for minimal CPU/memory overhead on high-performance infrastructure.

## Build and Development Commands

```bash
# Install dependencies
make deps

# Build the binary
make build

# Run all tests
make test

# Run a single test
go test ./internal/metrics/... -run TestCounterIncrement -v

# Lint the code
make lint

# Start the full stack (Go app + Prometheus + Jaeger + Grafana)
make up

# Stop the stack
docker compose down
```

## Architecture

The project follows strict separation of concerns across three layers: HTTP transport, business logic, and telemetry collection.

```
cmd/server/         - main.go: wires everything together, starts HTTP servers
internal/
  metrics/          - Prometheus registry, counter/histogram/gauge definitions, background worker
  tracing/          - OTel tracer setup, Jaeger gRPC exporter configuration
  middleware/        - RED pattern HTTP middleware (wraps handlers with rate/error/duration metrics + OTel spans)
  handler/          - HTTP handlers for application routes
configs/
  prometheus.yml    - Prometheus scrape config targeting the Go service
  grafana/          - Grafana provisioning (datasources + dashboard JSON)
docker-compose.yml  - Orchestrates app, prometheus, jaeger, grafana
Makefile
```

**Key design decisions:**
- A custom Prometheus registry (not the default global one) isolates metrics and prevents test pollution.
- The background worker in `internal/metrics` increments `app_events_total` every 5 seconds via a goroutine started at server init.
- The RED middleware wraps `http.Handler` and records `http_requests_total` (counter), `http_request_duration_seconds` (histogram), and starts an OTel span per request, injecting the trace context into response headers.
- OTel uses a stdout exporter in development and a Jaeger gRPC exporter (`JAEGER_ENDPOINT`) in the Docker Compose environment.

## Prometheus Metrics

| Metric | Type | Description |
|---|---|---|
| `app_events_total` | Counter | Background worker events, incremented every 5s |
| `request_duration_seconds` | Histogram | Simulated process latency with randomized durations |
| `go_goroutines_active` | Gauge | Current number of active goroutines |
| `http_requests_total` | Counter | RED rate/error counter with method, route, status labels |
| `http_request_duration_seconds` | Histogram | RED duration histogram with method, route labels |

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `HTTP_PORT` | `8080` | Main application HTTP port |
| `METRICS_PORT` | `9090` | Prometheus `/metrics` endpoint port |
| `JAEGER_ENDPOINT` | `localhost:4317` | Jaeger gRPC collector endpoint |
| `SERVICE_NAME` | `goobservabilitystack` | OTel service name attribute |

## Coding Standards

- Every exported function, type, constant, and variable must have a Godoc comment as a complete sentence starting with the item name (e.g., `// NewRegistry creates...`).
- Use Go 1.19+ native doc links (e.g., `[Service]`).
- No emojis or inline icons anywhere in code, comments, or markdown files.
- Prometheus metric names must follow strict naming conventions: `snake_case`, `_total` suffix for counters, `_seconds`/`_bytes` suffixes for units.
