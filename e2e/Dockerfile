# -----------------------------------------------------------------------------
# E2E Tests
# -----------------------------------------------------------------------------

FROM golang:1.25.7@sha256:cc737435e2742bd6da3b7d575623968683609a3d2e0695f9d85bee84071c08e6 AS builder

RUN apt-get update && apt-get install -y jq

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

ENTRYPOINT [ "/app/e2e/run.sh" ]
