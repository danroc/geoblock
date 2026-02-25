# --------------------------------------------------------------------------------------
# Build
# --------------------------------------------------------------------------------------

FROM --platform=$BUILDPLATFORM golang:1.26.0@sha256:b75179794e029c128d4496f695325a4c23b29986574ad13dd52a0d3ee9f72a6f AS builder

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

FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659

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
