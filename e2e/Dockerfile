# -----------------------------------------------------------------------------
# E2E Tests
# -----------------------------------------------------------------------------

FROM golang:1.25.6@sha256:fc24d3881a021e7b968a4610fc024fba749f98fe5c07d4f28e6cfa14dc65a84c AS builder

RUN apt-get update && apt-get install -y jq

WORKDIR /app
COPY . .

ENTRYPOINT [ "/app/e2e/run.sh" ]
