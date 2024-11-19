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
    --interval=10s \
    --timeout=10s \
    --start-period=30s \
    --start-interval=5s \
    --retries=3 \
  CMD wget -qO- http://localhost:8080/v1/health || exit 1

COPY --from=builder /app/dist/geoblock /app/geoblock

RUN addgroup -S app && adduser -S app -G app
USER app

WORKDIR /app
ENTRYPOINT [ "/app/geoblock" ]
