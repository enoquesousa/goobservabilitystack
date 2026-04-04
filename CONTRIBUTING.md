# Contributing

Thank you for your interest in contributing to `goobservabilitystack`.

## Prerequisites

- Go 1.22 or later
- Docker and Docker Compose
- `golangci-lint` — install via `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- `govulncheck` — install via `go install golang.org/x/vuln/cmd/govulncheck@latest`

## Workflow

1. Fork the repository and create a branch from `master`.
2. Make your changes.
3. Run the full check suite locally before opening a pull request (see below).
4. Open a pull request against `master` with a clear description of what the change does and why.

## Local checks

All of the following must pass before a pull request is opened:

```bash
# Install / tidy dependencies
make deps

# Build
make build

# Tests (race detector enabled)
make test

# Lint
make lint

# Vulnerability audit
govulncheck ./...
```

## Coding standards

### General

- Write idiomatic Go. Follow the style described in [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Keep changes focused. A pull request should do one thing.
- Do not introduce global state. Pass registries, counters, and other dependencies explicitly via constructors.

### Documentation

- Every exported function, type, constant, and variable must have a Godoc comment.
- Comments must be complete sentences starting with the name of the item being documented.
- Use Go 1.19+ native doc links (e.g., `[Registry]`, `[NewWorker]`) where referencing other exported symbols.

### Prometheus metrics

- Metric names must be `snake_case`.
- Counters must carry a `_total` suffix.
- Metrics measuring time must use `_seconds`; metrics measuring bytes must use `_bytes`.
- Register metrics on the custom `*prometheus.Registry` returned by `metrics.NewRegistry()`, never on the default global registry.
- New metric types must be covered by tests using `prometheus/testutil`.

### Tests

- Place tests under `tests/<package>/` as an external `_test` package (e.g., `package metrics_test`).
- Enable the race detector: `go test -race ./...`.
- Do not mock the Prometheus registry or counter in tests — use real `prometheus.NewCounter` / `prometheus.NewRegistry` instances.
- Table-driven tests are preferred for error-path coverage.

### Commit messages

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
docs: update README with full stack service table
```

## Pull request guidelines

- Target `master`.
- Keep pull requests small and reviewable. Split unrelated changes into separate PRs.
- Link any relevant GitHub issues in the PR description.
- Do not force-push to a PR branch after review has started.

## Reporting bugs

Open a [GitHub issue](https://github.com/ESousa97/goobservabilitystack/issues) with:

- A clear title.
- Steps to reproduce.
- Expected vs. actual behaviour.
- Go version (`go version`) and OS.

For security vulnerabilities, follow the process in [SECURITY.md](SECURITY.md) instead of opening a public issue.
