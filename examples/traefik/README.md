# Traefik Example

Run `docker compose up` to start the following services:

- `traefik`: Reverse proxy and load balancer
- `geoblock`: Geoblock service
- `whoami-1`: Example service (allowed)
- `whoami-2`: Example service (blocked)

Use HTTPie to test the services:

✅ Allowed:

```bash
$ http localhost:8080 Host:whoami-1.local
HTTP/1.1 200 OK
Content-Length: 359
Content-Type: text/plain; charset=utf-8
Date: Tue, 19 Nov 2024 18:00:44 GMT

Hostname: 0e9e69b86ee1
IP: 127.0.0.1
IP: 172.18.0.3
RemoteAddr: 172.18.0.5:40416
GET / HTTP/1.1
Host: whoami-1.local
User-Agent: HTTPie/3.2.4
Accept: */*
Accept-Encoding: gzip, deflate
X-Forwarded-For: 172.18.0.1
X-Forwarded-Host: whoami-1.local
X-Forwarded-Port: 80
X-Forwarded-Proto: http
X-Forwarded-Server: 5c8a21a19bdd
X-Real-Ip: 172.18.0.1
```

❌ Blocked:

```bash
http localhost:8080 Host:whoami-2.local
HTTP/1.1 403 Forbidden
Content-Length: 0
Date: Tue, 19 Nov 2024 18:00:48 GMT
```
