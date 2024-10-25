.PHONY: help lint tidy update vendor run build test

help: ## Show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-12s\033[0m %s\n", $$1, $$2}'

lint: tidy ## Run linter
	golines -w -m 88 --base-formatter=gofumpt .

tidy: ## Tidy up dependencies
	go mod tidy

update: ## Update dependencies
	go get -u ./...

vendor: ## Vendor dependencies
	go mod vendor

run: ## Run the main program
	go run ./cmd/server/

build: ## Build the binary
	mkdir -p dist
	go build -ldflags="-s -w" -o ./dist/server ./cmd/server/

test:
	go test ./...
