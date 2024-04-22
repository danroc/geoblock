.PHONY: lint tidy update run build

lint: tidy
	golines -w -m 88 --base-formatter=gofumpt .

tidy:
	go mod tidy

update:
	go get -u ./...

run:
	go run ./cmd/

build:
	mkdir -p dist
	go build -ldflags="-s -w" -o ./dist/geoblock ./cmd/
