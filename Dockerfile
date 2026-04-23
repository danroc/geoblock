# --------------------------------------------------------------------------------------
# Build
# --------------------------------------------------------------------------------------

FROM --platform=$BUILDPLATFORM golang:1.26.2@sha256:1e598ea5752ae26c093b746fd73c5095af97d6f2d679c43e83e0eac484a33dc3 AS builder

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

FROM alpine:3.23.4@sha256:5b10f432ef3da1b8d4c7eb6c487f2f5a8f096bc91145e68878dd4a5019afde11

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
