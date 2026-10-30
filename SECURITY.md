# Security Policy

## Supported Versions

| Version | Supported |
| ------- | ------------------ |
| latest (master) | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in `goobservabilitystack`, please report it responsibly by opening a [GitHub Security Advisory](https://github.com/enoquesousa/goobservabilitystack/security/advisories/new) instead of a public issue.

Include in your report:

- A description of the vulnerability and its potential impact.
- Steps to reproduce or a minimal proof of concept (PoC).
- Affected versions or commits, if known.

You can expect an acknowledgment within **5 business days** and a resolution or mitigation plan within **30 days**, depending on complexity.

## Security Considerations

### Metrics Endpoint Exposure

The `/metrics` endpoint exposes internal runtime data: goroutine counts, request
rates, error rates and latency histograms. The native server now defaults to
`127.0.0.1:9091`; `METRICS_HOST` and `METRICS_PORT` explicitly configure another
address. It does not authenticate scrapers. Keep it inaccessible from untrusted
networks.

Recommended mitigations:

- Bind the metrics server to a non-public interface (e.g., `127.0.0.1` or an internal network).
- Place the metrics endpoint behind a firewall rule or network policy that allows access only from your Prometheus scraper.
- In Kubernetes, use a `NetworkPolicy` that restricts port `9091` to the monitoring namespace.

### Application HTTP Server

The application server defaults to `127.0.0.1:8080` and serves `/healthz` without
authentication. `HTTP_HOST` and `HTTP_PORT` explicitly configure another address.
There is no application login, JWT, session store or role model. Tracing headers
are correlation data and are not proof of a caller's identity. If exposed
externally:

- Deploy behind a reverse proxy or load balancer with TLS termination.
- Do not expose the raw HTTP port directly to the internet.

### Local Compose Demo and Grafana

All published demo ports bind to `127.0.0.1`. The app listens on `0.0.0.0` inside
the Compose bridge to allow Prometheus scraping. Every container attached to
that bridge can reach the application, metrics and Jaeger without a login; the
demo is intended for a trusted local host, not a shared or public network.

`GRAFANA_ADMIN_PASSWORD` is required and must be non-empty. There is no built-in
password fallback; anonymous Grafana access and public signup are disabled.
The administrator password initializes Grafana's database. For an existing
database, use Grafana's password-change/reset procedure instead of assuming
that an environment update rotates the account password. Grafana authentication
does not protect the other services in this stack.

### Credential Handling

HTTP ports and hostnames are configuration, not credentials. OTLP headers,
Grafana passwords and mTLS private keys are credentials. When deploying:

- Do not commit `.env` files containing actual values to version control.
- Use Kubernetes `Secret` objects or a secrets manager (Vault, AWS Secrets Manager, etc.) for credentials instead of plain `ConfigMap` entries.
- Supply collector tokens through the standard OTLP headers configuration, never in a URL or request path.
- Protect private key files on disk and mount them read-only; the application does not encrypt them at rest.
- Docker administrators can inspect demo container environment variables. Do not share resolved `docker compose config` output, which can contain the Grafana password.
- Read credentials at process startup and restart after changing collector tokens or TLS files; this application has no refresh-token or live certificate-reload mechanism.

The Docker build context excludes `.env` files, `certs/`, `secrets/`, `*.key` and
`*.pem`. Exclusion from an image build does not replace filesystem permissions or
a secret manager.

### Jaeger / OTel Exporter

Trace data is exported via OTLP gRPC. The standard
`OTEL_EXPORTER_OTLP_ENDPOINT` defaults to `https://localhost:4317`; the
`OTEL_EXPORTER_OTLP_TRACES_*` settings take precedence for traces. `JAEGER_ENDPOINT`
is not read by this application.

TLS uses system trust or the configured PEM CA file. Configure both a client
certificate and matching private key for mTLS. The collector must verify that
client identity and implement its own authorization policy. Hostname verification
is preserved, and no failed TLS connection falls back to plaintext.

The exporter validates transport settings and TLS files before invoking the SDK.
Unreadable or unparseable files, incomplete identity pairs, malformed headers and contradictory
scheme/insecure settings fail initialization without including the supplied
values in the error message. A collector's untrusted certificate, wrong SAN or
unavailability is detected when the lazy connection exports/flushes spans.

The Compose demo explicitly uses `http://jaeger:4317` inside its trusted bridge.
CA files, client identity material and non-empty custom headers are rejected with
plaintext export. In production:

- Enable mTLS on the Jaeger collector endpoint.
- Restrict network access to only allow the application to reach port `4317`.
- Traces can contain request metadata (routes, status codes); avoid including personally identifiable information (PII) in span attributes.
- Do not place secrets in URL paths or tracing baggage. These are telemetry inputs, not a protected credential channel.

### Dependency Management

Dependencies are pinned in `go.sum` and managed via Go modules. To audit for known vulnerabilities:

```bash
go install golang.org/x/vuln/cmd/govulncheck@v1.9.0
govulncheck ./...
```

Dependency updates (including security patches) are automated via Dependabot.

### Docker Image Hygiene

When building a container image:

- Use a minimal base image (e.g., `gcr.io/distroless/static` or `scratch`) to reduce the attack surface.
- Run the process as a non-root user.
- Do not copy the Go toolchain or source code into the final image stage.
