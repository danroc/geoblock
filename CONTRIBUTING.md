# Contributing to Geoblock

## Development Setup

### Prerequisites

- Go 1.24 or later
- Docker (for e2e tests)
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
make test          # Run unit + e2e tests
make test-unit     # Unit tests with coverage
make test-e2e      # Dockerized e2e tests
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
