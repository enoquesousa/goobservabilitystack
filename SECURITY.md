# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest (master) | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in `goobservabilitystack`, please report it responsibly by opening a [GitHub Security Advisory](https://github.com/ESousa97/goobservabilitystack/security/advisories/new) rather than a public issue.

Include in your report:

- A description of the vulnerability and its potential impact.
- Steps to reproduce or a minimal proof-of-concept.
- Affected versions or commits, if known.

You can expect an acknowledgement within **5 business days** and a resolution or mitigation plan within **30 days**, depending on complexity.

## Security Considerations

### Metrics endpoint exposure

The `/metrics` endpoint (default port `9091`) exposes runtime internals — goroutine counts, request rates, error rates, and latency histograms. This endpoint must not be reachable from untrusted networks.

Recommended mitigations:

- Bind the metrics server to a non-public interface (e.g., `127.0.0.1` or an internal overlay network).
- Place the metrics endpoint behind a firewall rule or network policy that allows access only from your Prometheus scraper.
- In Kubernetes, use a `NetworkPolicy` that restricts port `9091` to the monitoring namespace.

### Application HTTP server

The application server (default port `8080`) serves `/healthz` and any future application routes. If exposed externally:

- Deploy behind a TLS-terminating reverse proxy or load balancer.
- Do not expose the raw HTTP port directly to the internet.

### Environment variables

Sensitive configuration is passed via environment variables (`HTTP_PORT`, `METRICS_PORT`, `JAEGER_ENDPOINT`, `SERVICE_NAME`). When deploying with Docker Compose or Kubernetes:

- Do not commit `.env` files containing real values to version control.
- Use Kubernetes `Secret` objects or a secrets manager (Vault, AWS Secrets Manager, etc.) for credentials rather than plain `ConfigMap` entries.

### Jaeger / OTel exporter

Trace data is exported over gRPC to `JAEGER_ENDPOINT` (default `localhost:4317`). In production:

- Enable mutual TLS on the Jaeger collector endpoint.
- Restrict network access so that only the application can reach port `4317`.
- Traces may contain request metadata (routes, status codes); avoid including personally identifiable information in span attributes.

### Dependency management

Dependencies are pinned in `go.sum` and managed via Go modules. To audit for known vulnerabilities:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

Dependency updates (including security patches) are automated via Dependabot for the `go_modules` ecosystem.

### Docker image hygiene

When building a container image:

- Use a minimal base image (e.g., `gcr.io/distroless/static` or `scratch`) to reduce the attack surface.
- Run the process as a non-root user.
- Do not copy the Go toolchain or source code into the final image stage.

## Out of scope

The following are outside the threat model of this project and will not be treated as security issues:

- Denial-of-service via high request rates to the metrics or application server (no rate limiting is implemented by design; handle this at the infrastructure layer).
- Vulnerabilities in Prometheus, Jaeger, or Grafana themselves — report those to their respective upstream projects.
