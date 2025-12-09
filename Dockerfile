# -----------------------------------------------------------------------------
# Build

FROM golang:1.25.5@sha256:68ee6dfa91a074c6e62a8af3ff0a427a36a78015778247af4f18398f37a5e91b AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN make clean build

# -----------------------------------------------------------------------------
# Run

FROM alpine:3.23.0@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375

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
