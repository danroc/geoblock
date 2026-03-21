# --------------------------------------------------------------------------------------
# Build
# --------------------------------------------------------------------------------------

FROM --platform=$BUILDPLATFORM golang:1.26.1@sha256:595c7847cff97c9a9e76f015083c481d26078f961c9c8dca3923132f51fe12f1 AS builder

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
