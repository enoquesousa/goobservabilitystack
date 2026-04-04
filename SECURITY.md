# Security Policy

## Supported Versions

| Version | Supported |
| ------- | ------------------ |
| latest (master) | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in `goobservabilitystack`, please report it responsibly by opening a [GitHub Security Advisory](https://github.com/ESousa97/goobservabilitystack/security/advisories/new) instead of a public issue.

Include in your report:

- A description of the vulnerability and its potential impact.
- Steps to reproduce or a minimal proof of concept (PoC).
- Affected versions or commits, if known.

You can expect an acknowledgment within **5 business days** and a resolution or mitigation plan within **30 days**, depending on complexity.

## Security Considerations

### Metrics Endpoint Exposure

The `/metrics` endpoint (default port `9091`) exposes internal runtime data — goroutine counts, request rates, error rates, and latency histograms. This endpoint must not be accessible from untrusted networks.

Recommended mitigations:

- Bind the metrics server to a non-public interface (e.g., `127.0.0.1` or an internal network).
- Place the metrics endpoint behind a firewall rule or network policy that allows access only from your Prometheus scraper.
- In Kubernetes, use a `NetworkPolicy` that restricts port `9091` to the monitoring namespace.

### Application HTTP Server

The application server (default port `8080`) serves `/healthz` and future routes. If exposed externally:

- Deploy behind a reverse proxy or load balancer with TLS termination.
- Do not expose the raw HTTP port directly to the internet.

### Environment Variables

Sensitive configurations are passed via environment variables (`HTTP_PORT`, `METRICS_PORT`, `JAEGER_ENDPOINT`, `SERVICE_NAME`). When deploying with Docker Compose or Kubernetes:

- Do not commit `.env` files containing actual values to version control.
- Use Kubernetes `Secret` objects or a secrets manager (Vault, AWS Secrets Manager, etc.) for credentials instead of plain `ConfigMap` entries.

### Jaeger / OTel Exporter

Trace data is exported via gRPC to the `JAEGER_ENDPOINT` (default `localhost:4317`). In production:

- Enable mTLS on the Jaeger collector endpoint.
- Restrict network access to only allow the application to reach port `4317`.
- Traces can contain request metadata (routes, status codes); avoid including personally identifiable information (PII) in span attributes.

### Dependency Management

Dependencies are pinned in `go.sum` and managed via Go modules. To audit for known vulnerabilities:

```bash
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

Dependency updates (including security patches) are automated via Dependabot.

### Docker Image Hygiene

When building a container image:

- Use a minimal base image (e.g., `gcr.io/distroless/static` or `scratch`) to reduce the attack surface.
- Run the process as a non-root user.
- Do not copy the Go toolchain or source code into the final image stage.
