# goobservabilitystack

> Reference implementation for observability in Go — RED metrics, OTel tracing, Prometheus, Jaeger, and Grafana in a single, composed stack.

![CI](https://github.com/ESousa97/goobservabilitystack/actions/workflows/ci.yml/badge.svg)
![Go Report Card](https://goreportcard.com/badge/github.com/ESousa97/goobservabilitystack?=v1)
![Go Reference](https://pkg.go.dev/badge/github.com/ESousa97/goobservabilitystack.svg)
![License](https://img.shields.io/github/license/ESousa97/goobservabilitystack)
![Go Version](https://img.shields.io/github/go-mod/go-version/ESousa97/goobservabilitystack)
![Last Commit](https://img.shields.io/github/last-commit/ESousa97/goobservabilitystack)

---

`goobservabilitystack` is a minimal and correct foundation for instrumenting HTTP services in Go. The project demonstrates the RED pattern (Rate, Errors, Duration) using Prometheus metrics and OpenTelemetry distributed tracing, with a clear separation between transport, business logic, and telemetry layers. Designed for low CPU and memory overhead in high-throughput infrastructures.

## Demo

Start the full stack and verify the running service:

```bash
# Health check
curl -i http://localhost:8080/healthz

# Prometheus metrics
curl -s http://localhost:9091/metrics | grep http_requests_total

# Inspect a trace in Jaeger
# Access http://localhost:16686 in your browser
```

![Grafana dashboard preview](assets/dashboard_preview.png)

## Tech Stack

| Component | Role |
|---|---|
| Go 1.25+ | Application runtime |
| [prometheus/client_golang](https://github.com/prometheus/client_golang) | Custom metrics registry, RED counters, and histograms |
| [open-telemetry/opentelemetry-go](https://github.com/open-telemetry/opentelemetry-go) | Distributed tracing with W3C context propagation |
| [Jaeger](https://www.jaegertracing.io/) | Tracing backend (OTLP gRPC receiver) |
| [Prometheus](https://prometheus.io/) | Metrics backend and scraper |
| [Grafana](https://grafana.com/) | Dashboard frontend with pre-provisioned panels |
| Docker Compose | Local orchestration of the full observability stack |

## Prerequisites

- Go >= 1.25
- Docker and Docker Compose (for the full stack)
- `golangci-lint`: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- `govulncheck`: `go install golang.org/x/vuln/cmd/govulncheck@latest`

## Installation and Usage

### As a binary

```bash
go install github.com/ESousa97/goobservabilitystack/cmd/server@latest
```

### From source

```bash
git clone https://github.com/ESousa97/goobservabilitystack.git
cd goobservabilitystack
make deps
make build
./bin/goobservabilitystack
```

The application server starts at `http://localhost:8080` and the metrics endpoint at `http://localhost:9091/metrics`.

### Full Observability Stack

```bash
make up
```

This command compiles the Go binary and starts Prometheus, Jaeger, and Grafana via Docker Compose.

| Service | URL |
|---|---|
| Application | http://localhost:8080 |
| Metrics | http://localhost:9091/metrics |
| Prometheus | http://localhost:9090 |
| Jaeger UI | http://localhost:16686 |
| Grafana | http://localhost:3000 (admin / admin) |

## Makefile Targets

| Target | Description |
|---|---|
| `make deps` | Downloads and tidies Go module dependencies |
| `make build` | Compiles the server binary to `bin/goobservabilitystack` |
| `make test` | Runs all tests with the race detector enabled |
| `make lint` | Runs `golangci-lint` across all packages |
| `make up` | Compiles and starts the full stack with Docker Compose |

## Architecture

The project follows a modular structure focused on separation of concerns:

- `cmd/server/main.go`: Entry point; orchestrates dependencies, starts HTTP servers, and manages graceful shutdown.
- `internal/metrics/`: Custom Prometheus registry and metric definitions (no global state).
- `internal/middleware/`: RED HTTP middleware; wraps handlers with rate/error/duration metrics and OTel spans.
- `internal/tracing/`: OTel Tracer provider; configured for gRPC export to Jaeger.
- `deploy/`: Infrastructure configurations (Prometheus and Grafana provisioning).

**Architectural Decisions:**

- Metric isolation via `prometheus.NewRegistry()` to avoid global pollution.
- Goroutine lifecycle management via `context.Context`.
- Metrics server on a dedicated port to isolate monitoring traffic.
- Strict use of timeouts in HTTP servers for resource resilience.

## Implementation Journey

The project was built in five incremental milestones, ensuring a solid foundation for observability:

1. **The Expositor**: Foundation with `prometheus/client_golang` and isolated metrics exposure.
2. **Advanced Instrumentation**: Use of semantic metric types (Histogram/Gauge) to measure internal health.
3. **Distributed Tracing**: OpenTelemetry SDK integration for granular operational visibility.
4. **RED Middleware**: Automation of HTTP telemetry for any registered route.
5. **The Full Stack**: Unification of collectors, visualizers, and the application via Docker.

## API Reference

Detailed package documentation is available at [pkg.go.dev](https://pkg.go.dev/github.com/ESousa97/goobservabilitystack).

| Method | Path | Port | Description |
|---|---|---|---|
| `GET` | `/healthz` | `8080` | Health check — returns `200 OK` |
| `GET` | `/metrics` | `9091` | Metrics exposure in Prometheus format |

## Configuration

| Variable | Type | Default | Description |
|---|---|---|---|
| `HTTP_PORT` | string | `8080` | Application HTTP server port |
| `METRICS_PORT` | string | `9091` | Prometheus scrape port `/metrics` |
| `JAEGER_ENDPOINT` | string | `localhost:4317` | Jaeger collector gRPC endpoint |
| `SERVICE_NAME` | string | `goobservabilitystack` | Service name for OpenTelemetry attributes |

## Development Roadmap

- [x] **Phase 1: The Expositor (Prometheus Registry)**: Custom registry configuration and `/metrics` exposure with background event counter.
- [x] **Phase 2: Custom Instrumentation (Gauges and Histograms)**: Processing latency measurement via Histogram and goroutine monitoring via Gauge.
- [x] **Phase 3: Distributed Tracing (OpenTelemetry and Spans)**: Implementation of spans detailing task execution and status attributes.
- [x] **Phase 4: The RED Pattern (Monitoring Middleware)**: Agnostic HTTP middleware for automatic capture of Rate, Errors, and Duration (RED) and context injection.
- [x] **Phase 5: The Ecosystem (Grafana, Jaeger, and Docker)**: Total integration via Docker Compose with gRPC trace export and pre-provisioned Grafana dashboards.

## Contributing

Link to [CONTRIBUTING.md](CONTRIBUTING.md) for details on setting up the environment and opening pull requests.

## License

Distributed under the MIT License. See `LICENSE` for more information.

## Author

Enoque Sousa — [enoquesousa.vercel.app](https://enoquesousa.vercel.app) · [github.com/ESousa97](https://github.com/ESousa97)
