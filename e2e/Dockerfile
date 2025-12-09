# -----------------------------------------------------------------------------
# E2E Tests
# -----------------------------------------------------------------------------

FROM golang:1.25.5@sha256:68ee6dfa91a074c6e62a8af3ff0a427a36a78015778247af4f18398f37a5e91b AS builder

RUN apt-get update && apt-get install -y jq

WORKDIR /app
COPY . .

ENTRYPOINT [ "/app/e2e/run.sh" ]
