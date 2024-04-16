from fastapi import FastAPI, Header

from service import service

app = FastAPI()


@app.get("/")
def check_source_ip(
    x_forwarded_method: str | None = Header(default=None),
    x_forwarded_proto: str | None = Header(default=None),
    x_forwarded_host: str | None = Header(default=None),
    x_forwarded_uri: str | None = Header(default=None),
    x_forwarded_for: str | None = Header(default=None),
):
    print("method", x_forwarded_method)
    print("proto", x_forwarded_proto)
    print("host", x_forwarded_host)
    print("uri", x_forwarded_uri)
    print("for", x_forwarded_for)
    print("country_code", service.country_code(x_forwarded_for))
    return service.country_code("62.35.85.135")
