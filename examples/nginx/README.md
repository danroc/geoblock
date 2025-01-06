# NGINX Example

## Deploying

Run `docker compose up` to start the following services defined in the
[`compose.yaml`](./compose.yaml) file:

- `whoami-1`: Example service (allowed)
- `whoami-2`: Example service (blocked)
- `geoblock`: Geoblock service
- `nginx`: Reverse proxy

This example will use the configuration defined in the
[`config.yaml`](./config.yaml) file. NGINX configuration is defined in the
[`nginx.conf`](./nginx.conf) file.

## Testing

In a different console, use `curl` to test the services:

**✅ Allowed:**

Request:

```bash
$ curl -fH "Host: whoami-1.local" http://localhost:8080
Hostname: faf5680d9138
IP: 127.0.0.1
IP: 172.20.0.3
RemoteAddr: 172.20.0.5:35620
GET / HTTP/1.1
Host: whoami-1.local
User-Agent: curl/8.7.1
Accept: */*
Connection: close
```

Logs:

```log
geoblock-1  | time="2025-01-06T09:30:19Z" level=info msg="Request authorized" request_domain=whoami-1.local request_method=GET source_asn=0 source_country= source_ip=172.20.0.1 source_org=
```

**❌ Blocked:**

Request:

```bash
$ curl -fH "Host: whoami-2.local" http://localhost:8080
curl: (22) The requested URL returned error: 403
```

Logs:

```log
geoblock-1  | time="2025-01-06T09:30:38Z" level=warning msg="Request denied" request_domain=whoami-2.local request_method=GET source_asn=0 source_country= source_ip=172.20.0.1 source_org=
```
