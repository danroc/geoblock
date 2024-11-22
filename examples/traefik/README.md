# Traefik Example

## Deploying

Run `docker compose up` to start the following services defined in the
[`compose.yaml`](./compose.yaml) file:

- `whoami-1`: Example service (allowed)
- `whoami-2`: Example service (blocked)
- `geoblock`: Geoblock service
- `traefik`: Reverse proxy

This example will use the configuration defined in the
[`config.yaml`](./config.yaml) file.

## Testing

In a different console, use `curl` to test the services:

**✅ Allowed:**

HTTP request:

```bash
curl -fH "Host: whoami-1.local" http://localhost:8080
```

HTTP response:

```bash
Hostname: 99704d18aa93
IP: 127.0.0.1
IP: 172.18.0.4
RemoteAddr: 172.18.0.5:56060
GET / HTTP/1.1
Host: whoami-1.local
User-Agent: curl/8.7.1
Accept: */*
Accept-Encoding: gzip
X-Forwarded-For: 172.18.0.1
X-Forwarded-Host: whoami-1.local
X-Forwarded-Port: 80
X-Forwarded-Proto: http
X-Forwarded-Server: 750f0338632d
X-Real-Ip: 172.18.0.1
```

Geoblock logs:

```log
geoblock-1  | time="2024-11-19T19:12:40Z" level=info msg="Request authorized" request_domain=whoami-1.local request_method=GET source_asn=0 source_country= source_ip=172.18.0.1 source_org=
```

**❌ Blocked:**

HTTP request:

```bash
curl -fH "Host: whoami-2.local" http://localhost:8080
```

HTTP response:

```bash
curl: (22) The requested URL returned error: 403
```

Geoblock logs:

```log
geoblock-1  | time="2024-11-19T19:12:41Z" level=warning msg="Request denied" request_domain=whoami-2.local request_method=GET source_asn=0 source_country= source_ip=172.18.0.1 source_org=
```
