# --------------------------------------------------------------------------------------
# E2E Tests
# --------------------------------------------------------------------------------------

FROM golang:1.26.0@sha256:c83e68f3ebb6943a2904fa66348867d108119890a2c6a2e6f07b38d0eb6c25c5 AS builder

RUN apt-get update && apt-get install -y jq

WORKDIR /app

# Cache dependency downloads in a separate layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENTRYPOINT [ "/app/e2e/run.sh" ]
