# Geoblock

Geoblock is a lightweight authorization service designed to control access to
services based on customizable rules.

It enables blocking or allowing access based on various criteria, providing
enhanced control over service access.

## Rules

Access can be controlled based on the following criteria:

- Client country
- Client autonomous system number (ASN)
- Client IP address
- Requested domain

Various criteria can be combined in a single rule, allowing for fine-grained
access control.

Rules are tested in order, and the first rule that matches the request is
applied. If no rule matches, the default policy is applied.

### Examples

Here are some examples of rules:

1. Allow access from private networks:

   ```yaml
   - networks:
       - 10.0.0.0/8
       - 127.0.0.0/8
       - 172.16.0.0/12
       - 192.168.0.0/16
     policy: allow
   ```

2. Block access from a specific country:

   ```yaml
   - countries:
       - FR
     policy: block
   ```

3. Allow access from a specific country to a list of domains:

   ```yaml
   - countries:
       - FR
     domains:
       - example.com
       - example.org
     policy: allow
   ```

4. Block access from a specific ASN:

   ```yaml
   - asns:
       - 1234
     policy: block
   ```

## Deployment

Here is an example of how to deploy Geoblock using Docker Compose with Traefik.

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

## Configuration file

Geoblock has a single configuration file that defines the access control rules
to apply. Its default location is `config.yaml` in the same directory as the
binary.

```yaml
# config.yaml
---
access_control:
  # Default policy to apply when no rules match, possible values are "allow"
  # and "deny"
  default_policy: deny

  # List of rules to apply, in order, to determine access control. If a rule
  # matches, the policy defined in the rule is applied. If no rule matches, the
  # default policy is applied.
  rules:
    # Example: allow access to example.com and example.org
    - domains:
        - example.com
        - example.org
      policy: allow

    # Example: allow access from private networks
    - networks:
        - 10.0.0.0/8
        - 127.0.0.0/8
        - 172.16.0.0/12
        - 192.168.0.0/16
      policy: allow

    # Example: deny access for clients from ASNs 1234 and 5678
    - autonomous_systems:
        - 1234
        - 5678
      policy: deny

    # Example: allow access from France
    - countries:
        - FR
      policy: allow

    # Example: allow access from France and the US to example.com
    - domains:
        - example.com
      countries:
        - FR
        - US
      policy: allow
```

## Environment variables

The following environment variables can be used to configure Geoblock:

- `GEOBLOCK_CONFIG`: Path to the configuration file (default: `config.yaml`).
- `GEOBLOCK_PORT`: Port to listen on (default: `8080`).

## Manual testing

Start geoblock with the provided example configuration:

```bash
GEOBLOCK_CONFIG=examples/config.yaml GEOBLOCK_PORT=8080 make run
```

### Missing X-Forwarded-For and X-Forwarded-Host headers

```http
GET http://localhost:8080/v1/forward-auth
```

### Missing X-Forwarded-Host header

```http
GET http://localhost:8080/v1/forward-auth
X-Forwarded-For: 127.0.0.1
```

### Missing X-Forwarded-For header

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
