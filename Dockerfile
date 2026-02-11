# -----------------------------------------------------------------------------
# Build

FROM golang:1.26.0@sha256:c83e68f3ebb6943a2904fa66348867d108119890a2c6a2e6f07b38d0eb6c25c5 AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN make clean build

# -----------------------------------------------------------------------------
# Run

FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659

RUN addgroup -S geoblock \
    && adduser -S geoblock -G geoblock \
    && mkdir -p /var/cache/geoblock \
    && chown geoblock:geoblock /var/cache/geoblock

COPY --from=builder /app/dist/geoblock /usr/bin/geoblock

USER geoblock

EXPOSE 8080

HEALTHCHECK \
    --interval=10s \
    --timeout=10s \
    --start-period=30s \
    --start-interval=5s \
    --retries=3 \
    CMD wget -qO- http://localhost:8080/v1/health || exit 1

ENTRYPOINT ["/usr/bin/geoblock"]
