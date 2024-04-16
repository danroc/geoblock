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
