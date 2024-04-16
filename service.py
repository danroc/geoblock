import os

from geoblock import CountryCode, Databases, country_code, download_file, read_db

COUNTRY_IPV4_DB_URL = (
    "https://cdn.jsdelivr.net/npm/"
    + "@ip-location-db/"
    + "geolite2-country/"
    + "geolite2-country-ipv4.csv"
)

COUNTRY_IPV6_DB_URL = (
    "https://cdn.jsdelivr.net/npm/"
    + "@ip-location-db/"
    + "geolite2-country/"
    + "geolite2-country-ipv6.csv"
)

COUNTRY_IPV4_DB_FILE = "geolite2-country-ipv4.csv"

COUNTRY_IPV6_DB_FILE = "geolite2-country-ipv6.csv"


class Service:
    """IP geolocation service."""

    def __init__(self) -> None:
        """Initialize the IP geolocation service."""
        if not os.path.exists(COUNTRY_IPV4_DB_FILE):
            print("Downloading IPv4 database...")
            download_file(COUNTRY_IPV4_DB_URL, COUNTRY_IPV4_DB_FILE, 10)

        if not os.path.exists(COUNTRY_IPV6_DB_FILE):
            print("Downloading IPv6 database...")
            download_file(COUNTRY_IPV6_DB_URL, COUNTRY_IPV6_DB_FILE, 10)

        print("Reading databases...")

        self.databases = Databases(
            read_db(COUNTRY_IPV4_DB_FILE),
            read_db(COUNTRY_IPV6_DB_FILE),
        )

        print("Ready!")

    def country_code(self, ip: str) -> CountryCode | None:
        """Get the country code of the given IP address.

        Args:
            ip (str): The IP address to lookup.

        Returns:
            str | None: The country code of the given IP address, or None if
            the IP address is not found in the database.
        """
        return country_code(self.databases, ip)


service = Service()
