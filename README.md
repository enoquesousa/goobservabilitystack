<div align="center">
  <h1>Go Observability Stack</h1>
  <p>Reference implementation for observability in Go — RED metrics, OTel tracing, Prometheus, Jaeger, and Grafana in a single, composed stack.</p>

  <img src="assets/github-go.png" alt="Go Observability Stack Banner" width="600px">

  <br>

  ![CI](https://img.shields.io/github/actions/workflow/status/ESousa97/goobservabilitystack/ci.yml?style=flat&logo=githubactions&logoColor=white)
  [![Go Report Card](https://goreportcard.com/badge/github.com/ESousa97/goobservabilitystack?style=flat)](https://goreportcard.com/report/github.com/ESousa97/goobservabilitystack)
  [![CodeFactor](https://www.codefactor.io/repository/github/esousa97/goobservabilitystack/badge?style=flat)](https://www.codefactor.io/repository/github/esousa97/goobservabilitystack)
  [![Go Reference](https://img.shields.io/badge/go.dev-reference-007d9c?style=flat&logo=go&logoColor=white)](https://pkg.go.dev/github.com/ESousa97/goobservabilitystack)
  ![License](https://img.shields.io/github/license/ESousa97/goobservabilitystack?style=flat&color=blue)
  ![Go Version](https://img.shields.io/github/go-mod/go-version/ESousa97/goobservabilitystack?style=flat&logo=go&logoColor=white)
  ![Last Commit](https://img.shields.io/github/last-commit/ESousa97/goobservabilitystack?style=flat)
</div>

---

**goobservabilitystack** is a minimal and correct foundation for instrumenting HTTP services in Go. The project demonstrates the RED pattern (Rate, Errors, Duration) using Prometheus metrics and OpenTelemetry distributed tracing, with a clear separation between transport, business logic, and telemetry layers. Designed for low CPU and memory overhead in high-throughput infrastructures.

## Technologies & Frameworks

<div align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=flat&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Docker-2496ED?style=flat&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/Prometheus-E6522C?style=flat&logo=prometheus&logoColor=white" alt="Prometheus" />
  <img src="https://img.shields.io/badge/Grafana-F46800?style=flat&logo=grafana&logoColor=white" alt="Grafana" />
  <img src="https://img.shields.io/badge/Jaeger-60EDC5?style=flat&logo=jaeger&logoColor=black" alt="Jaeger" />
  <img src="https://img.shields.io/badge/OpenTelemetry-000000?style=flat&logo=opentelemetry&logoColor=white" alt="OpenTelemetry" />
  <img src="https://img.shields.io/badge/gRPC-244C5A?style=flat&logo=grpc&logoColor=white" alt="gRPC" />
</div>

## Development Roadmap

- [x] **Phase 1: The Expositor (Prometheus Registry)**
  - **Objective:** Foundation with `prometheus/client_golang` and isolated metrics exposure.
  - **Status:** Custom registry configuration and `/metrics` exposure with background event counter completed.

- [x] **Phase 2: Custom Instrumentation (Gauges and Histograms)**
  - **Objective:** Use of semantic metric types to measure internal health.
  - **Status:** Processing latency measurement via Histogram and goroutine monitoring via Gauge implemented.

- [x] **Phase 3: Distributed Tracing (OpenTelemetry and Spans)**
  - **Objective:** OpenTelemetry SDK integration for granular operational visibility.
  - **Status:** Implementation of spans detailing task execution and status attributes completed.

- [x] **Phase 4: The RED Pattern (Monitoring Middleware)**
  - **Objective:** Automation of HTTP telemetry for any registered route.
  - **Status:** Agnostic HTTP middleware for automatic capture of Rate, Errors, and Duration (RED) finished.

- [x] **Phase 5: The Ecosystem (Grafana, Jaeger, and Docker)**
  - **Objective:** Unification of collectors, visualizers, and the application via Docker.
  - **Status:** Total integration via Docker Compose with gRPC trace export and pre-provisioned Grafana dashboards.

## Quick Start

### Prerequisites

- **Go 1.25+**
- **Docker and Docker Compose**
- **golangci-lint & govulncheck**

### Installation

```bash
git clone https://github.com/ESousa97/goobservabilitystack.git
cd goobservabilitystack
make deps
make build
```

### Running (Full Stack)

To start the server along with Prometheus, Jaeger, and Grafana, run:

```bash
make up
```

| Service | URL |
|---|---|
| **Application** | http://localhost:8080 |
| **Metrics** | http://localhost:9091/metrics |
| **Prometheus** | http://localhost:9090 |
| **Jaeger UI** | http://localhost:16686 |
| **Grafana** | http://localhost:3000 (admin / admin) |

### Demo

Verify the running service and observability data:

```bash
# Health check
curl -i http://localhost:8080/healthz

# Prometheus metrics
curl -s http://localhost:9091/metrics | grep http_requests_total
```

![Grafana dashboard preview](assets/dashboard_preview.png)

## Makefile Targets

| Target | Description |
|---|---|
| `make deps` | Downloads and tidies Go module dependencies |
| `make build` | Compiles the server binary to `bin/goobservabilitystack` |
| `make test` | Runs all tests with the race detector enabled |
| `make lint` | Runs `golangci-lint` across all packages |
| `make up` | Compiles and starts the full stack with Docker Compose |
| `make down` | Stops and removes the Docker containers |

## Architecture

The project follows a modular structure focused on separation of concerns, ensuring telemetry doesn't leak into business logic.

<div align="center">

```mermaid
graph TD
    Client[HTTP Client] --> App[Go Application]
    
    subgraph "Instrumentation Layer"
        App --> RED[RED Middleware]
        RED --> Prom[Prometheus Registry]
        RED --> OTel[OTel Tracer]
    end
    
    subgraph "Observability Stack"
        Prom -- HTTP Scrape --> PBackend[Prometheus]
        OTel -- gRPC Export --> JBackend[Jaeger]
        PBackend --> Grafana[Grafana Dashboards]
        JBackend --> Grafana
    end
    
    style App fill:#2da44e,stroke:#fff,stroke-width:1px,color:#fff
    style Grafana fill:#f39c12,stroke:#fff,stroke-width:1px,color:#fff
    style RED fill:#3498db,stroke:#fff,stroke-width:1px,color:#fff
```

</div>

- `cmd/server`: Entry point; orchestrates dependencies and starts HTTP servers.
- `internal/metrics`: Custom Prometheus registry and metric definitions.
- `internal/middleware`: RED HTTP middleware for rate, error, and latency capture.
- `internal/tracing`: OTel Tracer provider configured for gRPC export.
- `deploy`: Infrastructure configurations (Prometheus and Grafana provisioning).

## Configuration

| Variable | Type | Default | Description |
|---|---|---|---|
| `HTTP_PORT` | string | `8080` | Application HTTP server port |
| `METRICS_PORT` | string | `9091` | Prometheus scrape port `/metrics` |
| `JAEGER_ENDPOINT` | string | `localhost:4317` | Jaeger collector gRPC endpoint |
| `SERVICE_NAME` | string | `goobservabilitystack` | Service name for OpenTelemetry attributes |

## API Reference

| Method | Path | Port | Description |
|---|---|---|---|
| `GET` | `/healthz` | `8080` | Health check — returns `200 OK` |
| `GET` | `/metrics` | `9091` | Metrics exposure in Prometheus format |

## License

Distributed under the MIT License. See `LICENSE` for more information.

<div align="center">

## Author

**Enoque Sousa**

[![LinkedIn](https://img.shields.io/badge/LinkedIn-0077B5?style=flat&logo=linkedin&logoColor=white)](https://www.linkedin.com/in/enoque-sousa-bb89aa168/)
[![GitHub](https://img.shields.io/badge/GitHub-100000?style=flat&logo=github&logoColor=white)](https://github.com/ESousa97)
[![Portfolio](https://img.shields.io/badge/Portfolio-FF5722?style=flat&logo=target&logoColor=white)](https://enoquesousa.vercel.app)

**[⬆ Back to Top](#go-observability-stack)**

Made with ❤️ by [Enoque Sousa](https://github.com/ESousa97)

**Project Status:** Archived — Study Project

</div>
