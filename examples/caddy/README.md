# Caddy Example

## Deploying

Run `docker compose up` to start the following services defined in the
[`compose.yaml`](./compose.yaml) file:

- `whoami-1`: Example service (allowed)
- `whoami-2`: Example service (blocked)
- `geoblock`: Geoblock service
- `caddy`: Reverse proxy

This example will use the configuration defined in the
[`config.yaml`](./config.yaml) file. Caddy configuration is defined in the
[`Caddyfile`](./Caddyfile) file.

## Testing

In a different console, use `curl` to test the services:

**✅ Allowed:**

Request:

```bash
$ curl -fH "Host: whoami-1.local" http://localhost:8080
Hostname: e3edb3161f2e
IP: 127.0.0.1
IP: 172.19.0.3
RemoteAddr: 172.19.0.5:38128
GET / HTTP/1.1
Host: whoami-1.local
User-Agent: curl/8.7.1
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 172.19.0.1
X-Forwarded-Host: whoami-1.local
X-Forwarded-Proto: http

```

Logs:

```log
geoblock-1  | time="2025-01-06T09:33:34Z" level=info msg="Request authorized" request_domain=whoami-1.local request_method=GET source_asn=0 source_country= source_ip=172.19.0.1 source_org=
```

**❌ Blocked:**

Request:

```bash
$ curl -fH "Host: whoami-2.local" http://localhost:8080
curl: (22) The requested URL returned error: 403
```

Logs:

```log
geoblock-1  | time="2025-01-06T09:33:59Z" level=warning msg="Request denied" request_domain=whoami-2.local request_method=GET source_asn=0 source_country= source_ip=172.19.0.1 source_org=
```
