# --------------------------------------------------------------------------------------
# E2E Tests
# --------------------------------------------------------------------------------------

FROM golang:1.26.4@sha256:68cb6d68bed024785b69195b89af7ac7a444f27791435f98647edff595aa0479

RUN apt-get update && apt-get install -y jq

WORKDIR /app

# Cache dependency downloads in a separate layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENTRYPOINT ["/app/tests/e2e/run.sh"]
