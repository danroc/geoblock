# -----------------------------------------------------------------------------
# E2E Tests
# -----------------------------------------------------------------------------

FROM golang:1.25.7@sha256:85c0ab0b73087fda36bf8692efe2cf67c54a06d7ca3b49c489bbff98c9954d64 AS builder

RUN apt-get update && apt-get install -y jq

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

ENTRYPOINT [ "/app/e2e/run.sh" ]
