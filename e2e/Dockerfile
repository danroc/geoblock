# -----------------------------------------------------------------------------
# E2E Tests
# -----------------------------------------------------------------------------

FROM golang:1.25.6@sha256:ce63a16e0f7063787ebb4eb28e72d477b00b4726f79874b3205a965ffd797ab2 AS builder

RUN apt-get update && apt-get install -y jq

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

ENTRYPOINT [ "/app/e2e/run.sh" ]
