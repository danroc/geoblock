# Contributing to Geoblock

## Development Setup

### Prerequisites

- Go 1.24 or later
- Docker (for e2e and integration tests)
- [golangci-lint](https://golangci-lint.run/docs/welcome/install/)

### Installing Tools

```bash
make tools
```

This installs Go-based development tools managed via `go tool`.

## Development Workflow

### Building

```bash
make build    # Build to dist/geoblock
make run      # Run directly with go run
make docker   # Build Docker image
```

### Testing

```bash
make test              # Run unit + e2e + integration tests
make test-unit         # Unit tests with coverage
make test-e2e          # Dockerized e2e tests
make test-integration  # Integration tests with reverse proxies
make test-bench        # Run benchmarks
make test-coverage     # Generate coverage report
```

### Linting and Formatting

```bash
make format        # Run code formatters (gofumpt, goimports, golines)
make lint          # Run all linters (tidy + format + golangci-lint)
make lint-golangci # Run golangci-lint only
```

Run `make format` before committing to ensure consistent formatting.

## Code Style

- **88 characters max** line length enforced by golines
- Run `make format` to auto-fix formatting issues

## Release Process

Releases use semantic versioning with a `v` prefix (e.g. `v1.2.3`).

### Changelog

- Follow the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format
- Categories: `Added`, `Changed`, `Removed`, `Fixed`
- Use a **BREAKING** prefix for breaking changes
- Link PRs: `([#123](https://github.com/danroc/geoblock/pull/123))`

### Creating a Release

1. `git checkout -b release/X.Y.Z`
2. Update `CHANGELOG.md`: move `Unreleased` to `## [X.Y.Z] - YYYY-MM-DD`
3. `git commit -m "release: X.Y.Z"`
4. Push, create PR, merge
5. `git tag vX.Y.Z && git push origin vX.Y.Z`
6. CI builds and pushes the Docker image to `ghcr.io/danroc/geoblock:X.Y.Z`
