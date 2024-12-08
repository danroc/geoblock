# =============================================================================
# Project Configuration
# =============================================================================

# Project metadata
PROJECT_NAME := geoblock
DESCRIPTION := A simple IP-based geoblocking service

# Directories
ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
DIST_DIR := $(ROOT_DIR)/dist

# Colors
BLUE := \033[34m
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
BOLD := \033[1m
RESET := \033[0m

# =============================================================================
# @Linters
# =============================================================================

.PHONY: lint
lint: lint.lines lint.revive lint.sec lint.vet ## Run all linters

.PHONY: lint.lines
lint.lines: ## Lint lines length
	golines -w -m 79 --base-formatter=gofumpt .

.PHONY: lint.revive
lint.revive: ## Run revive linter
	revive -config revive.toml  ./...

.PHONY: lint.sec
lint.sec: ## Run gosec linter
	gosec ./...

.PHONY: lint.vet
lint.vet: ## Run go-vet linter
	go vet ./...

# =============================================================================
# @Dependencies
# =============================================================================

.PHONY: deps.tidy
deps.tidy: ## Tidy up dependencies
	go mod tidy

.PHONY: deps.update
deps.update: ## Update dependencies
	go get -u ./...

# We use the latest version of the tools since they cannot be automatically
# updated by Renovate. Once development dependencies are managed by `go tool`,
# we can remove this target and track the tools in the `go.mod` file.
.PHONY: deps.install
deps.install: ## Install development dependencies
	go install github.com/segmentio/golines@latest
	go install mvdan.cc/gofumpt@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/mgechev/revive@latest
	go install github.com/boumenot/gocover-cobertura@latest

# =============================================================================
# Directory Creation
# =============================================================================

$(DIST_DIR):
	mkdir -p $@

# =============================================================================
# @Build
# =============================================================================

.PHONY: run
run: ## Run the main program
	go run ./cmd/geoblock/

.PHONY: build
build: $(DIST_DIR) ## Build the binary
	go build -ldflags="-s -w" -o $(DIST_DIR)/geoblock ./cmd/geoblock/

.PHONY: docker
docker: ## Build docker image
	docker build -t geoblock .

# =============================================================================
# @Tests
# =============================================================================

.PHONY: test
test: test.unit test.e2e ## Run all tests

.PHONY: test.unit
test.unit: ## Run unit tests
	go test -coverprofile=coverage.out ./...

.PHONY: test.e2e
test.e2e: ## Run end-to-end tests
	docker build -f tests/Dockerfile -t geoblock-tests .
	docker run --rm geoblock-tests

.PHONY: test.coverage
test.coverage: test.unit ## Generate coverage report
	gocover-cobertura < coverage.out > coverage.xml

# =============================================================================
# @Help
# =============================================================================

.PHONY: help
help: ## Display this help message
	@echo "$(CYAN)$(BOLD)$(PROJECT_NAME) - $(DESCRIPTION)$(RESET)"
	@awk 'BEGIN { FS = ":.*?##" } \
		/^[a-zA-Z0-9._-]+:.*?##/ { printf "  $(CYAN)%-15s$(RESET) %s\n", $$1, $$2 } \
		/^# @/ { printf "\n$(MAGENTA)%s$(RESET)\n\n", substr($$0, 4) }' \
		$(MAKEFILE_LIST)
	@echo

.DEFAULT_GOAL := help
