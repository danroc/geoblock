# --------------------------------------------------------------------------------------
# Build
# --------------------------------------------------------------------------------------

FROM --platform=$BUILDPLATFORM golang:1.27.1@sha256:03fd17ca31f26e3c8d5b83ae0606044133c59e0788784dd6f371da59cdb3b46f AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

WORKDIR /app

# Cache dependency downloads in a separate layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG CGO_ENABLED=0
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH GOARM=${TARGETVARIANT#v} make clean build

# --------------------------------------------------------------------------------------
# Run
# --------------------------------------------------------------------------------------

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6

RUN addgroup -S geoblock \
    && adduser -S geoblock -G geoblock \
    && mkdir /cache \
    && chown geoblock:geoblock /cache

ENV GEOBLOCK_CACHE_DIR=/cache
ENV GEOBLOCK_CONFIG_FILE=/config.yaml

USER geoblock

EXPOSE 8080

HEALTHCHECK \
    --interval=10s \
    --timeout=10s \
    --start-period=30s \
    --start-interval=5s \
    --retries=3 \
    CMD wget --spider --no-verbose --tries=1 http://localhost:8080/v1/health || exit 1

COPY --from=builder /app/dist/geoblock /usr/bin/geoblock

ENTRYPOINT ["/usr/bin/geoblock"]
