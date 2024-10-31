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

EXPOSE 8080

HEALTHCHECK \
    --interval=5s \
    --timeout=5s \
    --start-period=5s \
    --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

COPY --from=builder /app/dist/geoblock /app/geoblock

WORKDIR /app
ENTRYPOINT [ "/app/geoblock" ]
