# -----------------------------------------------------------------------------
# E2E Tests
# -----------------------------------------------------------------------------

FROM golang:1.25.5@sha256:0ece421d4bb2525b7c0b4cad5791d52be38edf4807582407525ca353a429eccc AS builder

RUN apt-get update && apt-get install -y jq

WORKDIR /app
COPY . .

ENTRYPOINT [ "/app/e2e/run.sh" ]
