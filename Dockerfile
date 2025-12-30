# -----------------------------------------------------------------------------
# Build

FROM golang:1.25.5@sha256:31c1e53dfc1cc2d269deec9c83f58729fa3c53dc9a576f6426109d1e319e9e9a AS builder

WORKDIR /app
COPY . .

ARG CGO_ENABLED=0
RUN make clean build

# -----------------------------------------------------------------------------
# Run

FROM alpine:3.23.2@sha256:865b95f46d98cf867a156fe4a135ad3fe50d2056aa3f25ed31662dff6da4eb62

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
