import os
from geoblock import country_code, download_file, read_db


IPV4_COUNTRY_DB_URL = (
    "https://cdn.jsdelivr.net/npm/"
    + "@ip-location-db/"
    + "geolite2-country/"
    + "geolite2-country-ipv4.csv"
)

IPV6_COUNTRY_DB_URL = (
    "https://cdn.jsdelivr.net/npm/"
    + "@ip-location-db/"
    + "geolite2-country/"
    + "geolite2-country-ipv6.csv"
)

IPV4_COUNTRY_DB_FILE = "geolite2-country-ipv4.csv"

IPV6_COUNTRY_DB_FILE = "geolite2-country-ipv6.csv"


class Service:
    def __init__(self):
        if not os.path.exists(IPV4_COUNTRY_DB_FILE):
            print("Downloading IPv4 database...")
            download_file(IPV4_COUNTRY_DB_URL, IPV4_COUNTRY_DB_FILE, 10)

        if not os.path.exists(IPV6_COUNTRY_DB_FILE):
            print("Downloading IPv6 database...")
            download_file(IPV6_COUNTRY_DB_URL, IPV6_COUNTRY_DB_FILE, 10)

        print("Reading databases...")
        self.databases = (
            read_db(IPV4_COUNTRY_DB_FILE),
            read_db(IPV6_COUNTRY_DB_FILE),
        )

        print("Ready!")

    def country_code(self, ip):
        return country_code(self.databases, ip)


service = Service()


if __name__ == "__main__":
    print(service.country_code("62.35.85.135"))
    print(service.country_code("34.149.229.210"))
    print(service.country_code("142.251.220.163"))
    print(service.country_code("2a02:26f7:c9c8:4000:950e:981f:cef4:b0ed"))
