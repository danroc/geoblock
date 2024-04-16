from typing import Annotated
from fastapi import FastAPI, Header
from service import service

app = FastAPI()


@app.get("/")
def check_source_ip(
    user_agent: Annotated[str | None, Header()] = None,
    x_forwarded_method: Annotated[str | None, Header()] = None,
    x_forwarded_proto: Annotated[str | None, Header()] = None,
    x_forwarded_host: Annotated[str | None, Header()] = None,
    x_forwarded_uri: Annotated[str | None, Header()] = None,
    x_forwarded_for: Annotated[str | None, Header()] = None,
):
    print(user_agent)
    print("method", x_forwarded_method)
    print("proto", x_forwarded_proto)
    print("host", x_forwarded_host)
    print("uri", x_forwarded_uri)
    print("for", x_forwarded_for)
    print("country_code", service.country_code(x_forwarded_for))
    return service.country_code("62.35.85.135")
