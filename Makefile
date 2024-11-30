.PHONY: help lint tidy update run build test docker

help: ## Show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

lint: tidy ## Run linter
	golines -w -m 79 --base-formatter=gofumpt .
	revive -config revive.toml  ./...
	gosec ./...
	go vet ./...

tidy: ## Tidy up dependencies
	go mod tidy

update: ## Update dependencies
	go get -u ./...

run: ## Run the main program
	go run ./cmd/geoblock/

build: ## Build the binary
	mkdir -p dist
	go build -ldflags="-s -w" -o ./dist/geoblock ./cmd/geoblock/

test:
	go test -coverprofile=coverage.out ./...
	gocover-cobertura < coverage.out > coverage.xml

# We use the latest version of the tools since they cannot be automatically
# updated by Renovate. Once development dependencies are managed by `go tool`,
# we can remove this target and track the tools in the `go.mod` file.
deps.dev: ## Install development dependencies
	go install github.com/segmentio/golines@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/mgechev/revive@latest
	go install github.com/boumenot/gocover-cobertura@latest

docker: ## Build docker image
	docker build -t geoblock .
