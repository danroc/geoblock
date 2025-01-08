# -----------------------------------------------------------------------------
# Build

FROM golang:1.23.4 AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN make build

# -----------------------------------------------------------------------------
# Run

FROM alpine:3.21.2

EXPOSE 8080

HEALTHCHECK \
    --interval=10s \
    --timeout=10s \
    --start-period=30s \
    --start-interval=5s \
    --retries=3 \
  CMD wget -qO- http://localhost:8080/v1/health || exit 1

COPY --from=builder /app/dist/geoblock /usr/bin/geoblock

RUN addgroup -S app && adduser -S app -G app
USER app

ENTRYPOINT [ "/usr/bin/geoblock" ]
