# Geoblock Development Guide

## Architecture

Go-based forward auth service for geoblocking. Works with reverse proxies (Traefik, NGINX, Caddy) to authorize requests based on IP, country, ASN, domain, and HTTP method.

### Core Components

- `cmd/geoblock/main.go`: Entry point with auto-reload (5s) and auto-update (24h) goroutines
- `internal/rules`: Rule matching engine with atomic config updates via `atomic.Pointer`
- `internal/ipinfo`: IP resolution using interval trees and CSV databases from CDN
- `internal/server`: HTTP server exposing `/v1/forward-auth`, `/v1/health`, `/metrics`
- `internal/config`: YAML configuration with validation using `go-playground/validator`
- `internal/itree`: Generic AVL-based interval tree for O(log n) IP lookups

### Data Flow

1. Proxy sends headers (`X-Forwarded-For`, `X-Forwarded-Host`, `X-Forwarded-Method`)
2. Server extracts IP, resolves country/ASN via interval tree
3. Rules engine evaluates sequentially until first match
4. Returns `204 No Content` (allowed) or `403 Forbidden` (denied)

## Commands

```bash
make build              # Build to dist/geoblock
make run                # Run with go run
make docker             # Build Docker image
make test               # Unit + e2e tests
make test-unit          # Unit tests with coverage
make test-e2e           # Dockerized e2e tests (e2e/run.sh)
make lint               # All linters (tidy + format + golangci-lint)
make format             # Run before committing
```

## Code Conventions

- **88 char max** line length; linter config in `.golangci.yml`
- Use `atomic.Pointer[T]` for concurrent reads/writes (see `rules.Engine.config`, `ipinfo.Resolver.db`)
- YAML via `goccy/go-yaml`; validation tags like `validate:"required,oneof=allow deny"` (see `config/schema.go`)
- Custom types implement `UnmarshalYAML` (e.g., `config.CIDR` wraps `netip.Prefix`)
- Use `_test` package suffix; table-driven tests with `name`, `config`, `query`, `want` fields
- Return errors, don't panic; use `#nosec G304` for justified gosec exceptions
- Structured logging: `log.Error().Err(err).Str("path", path).Msg(...)`

## Rule Matching

- Empty conditions match all (no `domains` = any domain)
- ALL conditions must match (AND logic)
- First matching rule wins
- Case-insensitive for domains, methods, countries
- Wildcards via `glob.Star()` (simple `*` matching)

## IP Resolution

- Two interval trees: countries and ASN/org
- Auto-downloaded from `cdn.jsdelivr.net` (ip-location-db)
- Resolution merges results (last non-zero field wins)
- Local IPs via `localNetworkCIDRs` in `server/server.go`

## Environment Variables

- `GEOBLOCK_CACHE_DIR`: IP database cache directory (default: `/var/cache/geoblock`, empty to disable)
- `GEOBLOCK_CONFIG_FILE`: Config path (default: `/etc/geoblock/config.yaml`)
- `GEOBLOCK_PORT`: Server port (default: `8080`)
- `GEOBLOCK_LOG_LEVEL`: trace|debug|info|warn|error|fatal|panic (default: `info`)
- `GEOBLOCK_LOG_FORMAT`: json|text (default: `json`)

## Common Tasks

### Adding a Rule Condition

1. Add field to `config.AccessControlRule` with validation tag
2. Update `ruleApplies()` in `internal/rules/engine.go` with `match()` call
3. Add table-driven tests in `engine_test.go`
4. Update example config and README

### Modifying Metrics

1. Update definitions in `internal/metrics/metrics.go`
2. Instrument in `internal/server/server.go`
3. Update `e2e/metrics-expected.prometheus`
4. Run `make test-e2e`

### Changing Database Sources

1. Update URLs in `internal/ipinfo/resolver.go` constants
2. Adjust parser functions if CSV format changes
3. Test with `make run` (auto-downloads on startup)

## Release Process

- Semantic versioning with `v` prefix; embedded via ldflags from `git describe --tags`
- Format: `<tag>-<commits_ahead>-g<short_hash>[-dirty|-broken]`
- Parsed by `internal/version` package

### Changelog

- [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format
- Categories: Added, Changed, Removed, Fixed
- **BREAKING** prefix for breaking changes
- Link PRs: `([#123](https://github.com/danroc/geoblock/pull/123))`

### Creating a Release

1. `git checkout -b release/X.Y.Z`
2. Update `CHANGELOG.md`: move Unreleased to `## [X.Y.Z] - YYYY-MM-DD`
3. `git commit -m "release: X.Y.Z"`
4. Push, create PR, merge
5. `git tag vX.Y.Z && git push origin vX.Y.Z`
6. CI builds and pushes to `ghcr.io/danroc/geoblock:X.Y.Z` (no `v` prefix)

### CI Workflows

- `build-test-lint.yml`: On push/PR - build, lint, test, check clean working dir
- `publish-docker.yml`: On `v*.*.*` tags - multi-arch Docker to GHCR
