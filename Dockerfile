# -----------------------------------------------------------------------------
# Build

FROM golang:1.23.2 AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN make build

# -----------------------------------------------------------------------------
# Run

FROM alpine:3.20.3

COPY --from=builder /app/dist/geoblock /app/geoblock

WORKDIR /app
ENTRYPOINT [ "/app/geoblock" ]
