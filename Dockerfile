FROM golang:1.23.2 AS builder

ARG CGO_ENABLED=0

WORKDIR /app
COPY . .

RUN make build

FROM alpine:3.20.3

WORKDIR /geoblock
COPY --from=builder /app/dist/server /geoblock/server
COPY config/geoblock.yaml /geoblock/config/geoblock.yaml

ENTRYPOINT [ "/geoblock/server" ]
