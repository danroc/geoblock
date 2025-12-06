# Geoblock Development Guide

## Architecture Overview

Geoblock is a lightweight Go-based forward authentication service for geoblocking. It works with reverse proxies (Traefik, NGINX, Caddy) to authorize requests based on client IP, country, ASN, domain, and HTTP method.

### Core Components

- **`cmd/geoblock/main.go`**: Entry point with auto-reload (5s) and auto-update (24h) goroutines
- **`internal/rules`**: Rule matching engine with atomic config updates via `atomic.Pointer`
- **`internal/ipinfo`**: IP resolution using interval trees and CSV databases from CDN
- **`internal/server`**: HTTP server exposing `/v1/forward-auth`, `/v1/health`, `/metrics`
- **`internal/config`**: YAML configuration with validation using `go-playground/validator`
- **`internal/itree`**: Generic AVL-based interval tree for O(log n) IP lookups

### Data Flow

1. Reverse proxy sends request headers (`X-Forwarded-For`, `X-Forwarded-Host`, `X-Forwarded-Method`)
2. Server extracts IP and resolves country/ASN using interval tree lookups
3. Rules engine evaluates rules sequentially until first match
4. Returns `204 No Content` (allowed) or `403 Forbidden` (denied)

## Development Workflow

### Build & Run

```bash
make build              # Build to dist/geoblock
make run                # Run directly with go run
make docker             # Build Docker image
```

### Testing

```bash
make test               # Run unit + e2e tests
make test-unit          # Unit tests with coverage
make test-e2e           # Dockerized e2e tests (see e2e/run.sh)
make test-coverage      # Generate coverage.xml (Cobertura)
```

### Linting

```bash
make lint               # Run all linters
make lint-format        # gofumpt + golines (79 char limit)
make lint-revive        # Custom rules in revive.toml
make lint-sec           # gosec security scanner
make lint-staticcheck   # staticcheck
```

Run `make lint-format` before committing to ensure consistent formatting.

## Code Conventions

### Line Length & Formatting

- **79 characters max** enforced by `golines` with `gofumpt` base formatter
- Always run `make lint-format` to auto-fix

### Concurrency Patterns

- **Atomic updates**: Use `atomic.Pointer[T]` for safe concurrent reads/writes (see `rules.Engine.config`, `ipinfo.Resolver.db`)
- **Long-running tasks**: Launch goroutines in `main()` for auto-reload and auto-update
- **No mutexes**: Prefer atomic operations over locks for performance

### Configuration & Validation

- YAML unmarshaling with struct tags using `goccy/go-yaml`
- Validation tags: `validate:"required,oneof=allow deny"` (see `config/schema.go`)
- Custom types implement `UnmarshalYAML` (e.g., `config.CIDR` wraps `netip.Prefix`)

### Testing Style

- Use `_test` package suffix for black-box tests (e.g., `rules_test`, `config_test`)
- Table-driven tests with `name`, `config`, `query`, `want` fields
- Place test files alongside implementation (`engine.go` → `engine_test.go`)

### Error Handling

- Return errors, don't panic (except in test helpers)
- Use `#nosec G304` comment for justified gosec exceptions
- Log errors with structured fields: `log.Error().Err(err).Str("path", path).Msg(...)`

## Project-Specific Details

### Rule Matching Logic

- Empty conditions match all (e.g., no `domains` = match any domain)
- ALL conditions must match for a rule to apply (AND logic)
- First matching rule wins (sequential evaluation)
- Case-insensitive for domains, methods, countries
- Wildcards supported via `glob.Star()` (simple `*` matching)

### IP Resolution

- Uses two interval trees: one for countries, one for ASN/org
- Databases auto-downloaded from `cdn.jsdelivr.net` (ip-location-db project)
- Resolution merges results from both trees (last non-zero field wins)
- Local IPs detected via `localNetworkCIDRs` in `server/server.go`

### Environment Variables

- `GEOBLOCK_CONFIG`: Config path (default: `/etc/geoblock/config.yaml`)
- `GEOBLOCK_PORT`: Server port (default: `8080`)
- `GEOBLOCK_LOG_LEVEL`: `trace|debug|info|warn|error|fatal|panic` (default: `info`)
- `GEOBLOCK_LOG_FORMAT`: `json|text` (default: `json`)

### Dependencies

- `zerolog`: Structured logging
- `goccy/go-yaml`: YAML parsing
- `go-playground/validator`: Config validation
- No external libraries for interval tree (custom implementation)

## Common Tasks

### Adding a New Rule Condition

1. Add field to `config.AccessControlRule` with validation tag
2. Update `ruleApplies()` in `internal/rules/engine.go` with new `match()` call
3. Add table-driven tests in `engine_test.go`
4. Update example config and README

### Modifying Metrics

1. Update metric definitions in `internal/metrics/metrics.go`
2. Instrument in `internal/server/server.go` handlers
3. Update expected output in `e2e/metrics-expected.prometheus`
4. Run `make test-e2e` to verify

### Changing Database Sources

1. Update URLs in `internal/ipinfo/resolver.go` constants
2. Adjust parser functions if CSV format changes
3. Test with `make run` (auto-downloads on startup)

## Release Process

### Version Management

- Uses **semantic versioning** (major.minor.patch) with `v` prefix (e.g., `v0.4.0`)
- Version embedded at build time via ldflags: `-X 'github.com/danroc/geoblock/internal/version.Version=$(VERSION)'`
- `VERSION` derived from `git describe --tags --dirty --broken --long`
- Version string format: `v0.4.0-0-abc1234` (tag-ahead-commit)
- Parsed by `internal/version` package to extract tag, commit, dirty/broken state

### Changelog

- Follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format
- Categories: Added, Changed, Removed, Fixed
- **BREAKING** prefix for breaking changes
- Link PRs in entries: `([#123](https://github.com/danroc/geoblock/pull/123))`
- Update `[Unreleased]` section during development
- Move to versioned section on release

### Creating a Release

1. Update `CHANGELOG.md`:
   - Move `[Unreleased]` changes to new version section
   - Add release date in format: `## [X.Y.Z] - YYYY-MM-DD`
   - Ensure all PRs are linked
2. Commit changelog: `git commit -m "release: X.Y.Z"`
3. Create tag: `git tag vX.Y.Z`
4. Push tag: `git push origin vX.Y.Z`
5. **Automated**: GitHub Actions workflow triggers on `v*.*.*` tags
6. **Automated**: Docker image built and pushed to `ghcr.io/danroc/geoblock`

### CI/CD Workflows

- **`build-test-lint.yml`**: Runs on push/PR to main
  - Installs Go 1.25.5, runs `make deps-tools`, `make build`, `make lint`, `make test`
  - Fails if working directory is dirty after job (enforces generated code committed)
- **`publish-docker.yml`**: Runs on version tags (`v*.*.*`)
  - Builds multi-arch Docker image
  - Pushes to GitHub Container Registry with semver tag
  - Requires clean working directory

### Breaking Changes

- Mark as **BREAKING** in changelog
- Increment major version (or minor for 0.x.y)
- Document migration path in changelog entry
