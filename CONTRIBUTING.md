# Contributing

Thank you for your interest in contributing to `goobservabilitystack`.

## Prerequisites

- Go 1.26.10 or later
- Docker and Docker Compose
- `golangci-lint` 2.12.1, matching the CI workflow
- `govulncheck` 1.9.0 — install via `go install golang.org/x/vuln/cmd/govulncheck@v1.9.0`

## Workflow

1. Fork the repository and create a branch from `master`.
2. Make your changes.
3. Run the test suite locally before opening a pull request (see below).
4. Open a pull request against `master` with a clear description of what the change does and why.

## Local Verifications

All of the following checks must pass before a pull request is opened:

```bash
# Install / tidy dependencies
make deps

# Build
make build

# Tests (with race detector enabled)
make test

# Lint
make lint

# Vulnerability audit
govulncheck ./...
```

The full-stack demo also requires a non-empty `GRAFANA_ADMIN_PASSWORD`; see the
README for the interactive setup. Do not put real passwords, collector headers,
private keys, or resolved Compose configuration in test output or pull requests.
TLS and collector tests must use temporary certificates and synthetic credentials.

## Coding Standards

### General

- Write idiomatic Go code. Follow the style described in [Effective Go](https://go.dev/doc/effective_go) and [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Keep changes focused. A pull request should do only one thing.
- Do not introduce global state. Pass registries, counters, and other dependencies explicitly via constructors.

### Documentation

- Every exported function, type, constant, and variable **MUST** have a Godoc comment.
- Comments should be complete sentences starting with the name of the documented item.
- Use native Go 1.19+ documentation links (e.g., `[Registry]`, `[NewWorker]`) when referencing other exported symbols.

### Prometheus Metrics

- Metric names must be in `snake_case`.
- Counters must have the `_total` suffix.
- Metrics measuring time should use `_seconds`; metrics measuring bytes should use `_bytes`.
- Register metrics in the custom `*prometheus.Registry` returned by `metrics.NewRegistry()`, never in the default global registry.
- New metric types should be covered by tests using `prometheus/testutil`.

### Testing

- Place tests in `tests/<package>/` as an external `_test` package (e.g., `package metrics_test`).
- Enable the race detector: `go test -race ./...`.
- Do not use mocks for the Prometheus registry or counters in tests — use real instances of `prometheus.NewCounter` / `prometheus.NewRegistry`.
- Table-driven tests are preferred for covering error paths.

### Commit Messages

Use the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <short description>

[optional body]
```

Common types: `feat`, `fix`, `test`, `refactor`, `build`, `docs`, `chore`.

Examples:

```
feat(metrics): add http_requests_total RED counter
fix(worker): stop ticker before returning from Run
test(registry): add table-driven tests for registration errors
docs: update README with stack services table
```

## Pull Request Guidelines

- Target: `master`.
- Keep pull requests small and easy to review. Split unrelated changes into separate PRs.
- Link any relevant GitHub issues in the PR description.
- Do not force-push to a PR branch after a review has started.

## Reporting Bugs

Open a [GitHub issue](https://github.com/enoquesousa/goobservabilitystack/issues) with:

- A clear title.
- Steps to reproduce.
- Expected vs. actual behavior.
- Go version (`go version`) and OS.

For security vulnerabilities, follow the process in [SECURITY.md](SECURITY.md) instead of opening a public issue.
