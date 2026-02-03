# -----------------------------------------------------------------------------
# Build

FROM golang:1.25.6@sha256:0c87ea6991c06552ca5f516e3aeb434056bac3b674f32f612691692668e57074 AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN make clean build

# -----------------------------------------------------------------------------
# Run

FROM alpine:3.23.3@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659

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
