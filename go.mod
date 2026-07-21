module github.com/danroc/geoblock

go 1.25.0

toolchain go1.26.5

require (
	github.com/go-playground/validator/v10 v10.30.3
	github.com/goccy/go-yaml v1.19.2
	github.com/prometheus/client_golang v1.24.0
	github.com/rs/zerolog v1.35.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/boumenot/gocover-cobertura v1.5.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.0 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	golang.org/x/mod v0.38.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/telemetry v0.0.0-20260708182218-49f421fb7959 // indirect
	golang.org/x/text v0.38.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
	golang.org/x/vuln v1.6.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

tool (
	github.com/boumenot/gocover-cobertura
	golang.org/x/vuln/cmd/govulncheck
)
