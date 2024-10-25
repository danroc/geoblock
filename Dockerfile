FROM golang:1.23.2 AS builder

ARG CGO_ENABLED=0

WORKDIR /app
COPY . .
RUN make build

FROM alpine:3.20.3

COPY --from=builder /app/dist/geoblock /app/geoblock

WORKDIR /app
ENTRYPOINT [ "/app/geoblock" ]
