# -----------------------------------------------------------------------------
# E2E Tests
# -----------------------------------------------------------------------------

FROM golang:1.25.7@sha256:011d6e21edbc198b7aeb06d705f17bc1cc219e102c932156ad61db45005c5d31 AS builder

RUN apt-get update && apt-get install -y jq

WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

ENTRYPOINT [ "/app/e2e/run.sh" ]
