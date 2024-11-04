<h1 align="center">Geoblock</h1>
<p align="center">
  <i>Block clients based on their country, ASN or network.</i>
</p>

## Introduction

Geoblock is a lightweight authorization service that restricts client access
based on:

- Client's country
- Client's IP address
- Client's ASN (Autonomous System Number)
- Requested domain

## Configuration

Geoblock uses a single configuration file (`config.yaml` by default) to set
access control rules. Rules are evaluated sequentially, applying the first
match per request. If no rules match, the default policy applies.

A rule matches if all specified conditions are met. Rules can include one or
more of the following criteria:

- `countries`: List of country codes (ISO 3166-1 alpha-2)
- `domains`: List of domain names
- `networks`: List of IP ranges in CIDR notation
- `autonomous_systems`: List of ASNs

Example configuration file:

```yaml
---
access_control:
  # Default action when no rules match ("allow" or "deny").
  default_policy: deny

  # List of access rules, evaluated in order. The first matching rule’s
  # policy is applied. If no rule matches, the default policy is used.
  #
  # IMPORTANT: Replace these example rules with your own rules.
  rules:
    # Allow access from internal/private networks.
    - networks:
        - 10.0.0.0/8
        - 127.0.0.0/8
        - 172.16.0.0/12
        - 192.168.0.0/16
      policy: allow

    # Deny access for clients from ASNs 1234 and 5678.
    - autonomous_systems:
        - 1234
        - 5678
      policy: deny

    # Allow access to example.com and example.org from clients in
    # France (FR) and the United States (US).
    - domains:
        - example.com
        - example.org
      countries:
        - FR
        - US
      policy: allow
```

## HTTP API

The following HTTP endpoints are exposed by Geoblock.

### `GET /v1/forward-auth`

Check if a client is authorized to access a domain.

#### Request

| Header             | Required | Description         |
| :----------------- | :------: | :------------------ |
| `X-Forwarded-For`  |   Yes    | Client's IP address |
| `X-Forwarded-Host` |   Yes    | Requested domain    |

#### Response

| Status | Description |
| :----- | :---------- |
| `204`  | Authorized  |
| `403`  | Forbidden   |

### `GET /v1/health`

Check if the service is healthy.

#### Response

| Status | Description |
| :----- | :---------- |
| `204`  | Healthy     |

## Deployment

### Traefik

```yaml
# compose.yaml
---
services:
  traefik:
    # Traefik configuration...

  geoblock:
    image: ghcr.io/danroc/geoblock:latest
    container_name: geoblock
    networks:
      - proxy
    volumes:
      - ./config.yaml:/app/config.yaml
    labels:
      - traefik.enable=true
      - traefik.http.middlewares.geoblock.forwardauth.address=http://geoblock:8080/v1/forward-auth
      - traefik.http.middlewares.geoblock.forwardauth.trustForwardHeader=true
    restart: unless-stopped

networks:
  proxy:
    external: true
```

## Environment variables

The following environment variables can be used to configure Geoblock:

| Variable          | Description                    | Default         |
| :---------------- | :----------------------------- | :-------------- |
| `GEOBLOCK_CONFIG` | Path to the configuration file | `./config.yaml` |
| `GEOBLOCK_PORT`   | Port to listen on              | `8080`          |

## Manual testing

Start geoblock with the provided example configuration:

```bash
GEOBLOCK_CONFIG=examples/config.yaml GEOBLOCK_PORT=8080 make run
```

### Missing `X-Forwarded-For` and `X-Forwarded-Host` headers

```http
GET http://localhost:8080/v1/forward-auth
```

### Missing `X-Forwarded-Host` header

```http
GET http://localhost:8080/v1/forward-auth
X-Forwarded-For: 127.0.0.1
```

### Missing `X-Forwarded-For` header

```http
GET http://localhost:8080/v1/forward-auth
X-Forwarded-Host: example.com
```

### Blocked country

```http
GET http://localhost:8080/v1/forward-auth
X-Forwarded-For: 8.8.8.8
X-Forwarded-Host: example.com
```

### Request authorized

```http
GET http://localhost:8080/v1/forward-auth
X-Forwarded-For: 127.0.0.1
X-Forwarded-Host: example.com
```

## Roadmap

- [x] Support environment variables
- [x] Docker image
- [x] Publish Docker image
- [ ] Write documentation
- [ ] Cache responses
- [ ] Cache databases
- [ ] Auto-update databases
- [ ] Add more tests
- [ ] Add metrics
- [ ] Support command line arguments
