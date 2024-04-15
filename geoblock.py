"""IP Geolocation Blocker."""

__version__ = "0.1.0"

import csv
import os
from bisect import bisect_right
from dataclasses import dataclass
from ipaddress import IPv4Address, IPv6Address, ip_address

import requests

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


def download_file(
    url: str,
    path: str | None = None,
    timeout: float | tuple[float, float] | None = None,
) -> requests.Response:
    """Download a file from the given URL and saves it to the specified path.

    Args:
        url (str): The URL of the file to download.

        path (str, optional): The path where the downloaded file should be
        saved. If not provided, the file will not be saved to disk. Defaults to
        None.

        timeout (float or tuple[float, float], optional): The timeout value for
        the request. Can be a float representing the timeout in seconds, or a
        tuple of two floats representing the connection timeout and read
        timeout respectively. Defaults to None, which means no timeout.

    Returns:
        requests.Response: The response object.

    Raises:
        requests.HTTPError: If the request to download the file fails.
    """
    response = requests.get(url, timeout=timeout)
    response.raise_for_status()
    if path is not None:
        with open(path, "wb") as file:
            file.write(response.content)
    return response


@dataclass
class IPRangeData:
    """IP Range Data class."""

    country_code: str


@dataclass(order=True)
class IPRange:
    """IP Range class."""

    start: IPv4Address | IPv6Address
    end: IPv4Address | IPv6Address
    data: IPRangeData


def ip_range_from_csv(row):
    """Create an IPRange object from a CSV row.

    Args:
        row (list): A row from a CSV file containing an IP address range and a
        country code.

    Returns:
        IPRange: An IPRange object.
    """
    return IPRange(
        ip_address(row[0]),
        ip_address(row[1]),
        IPRangeData(row[2]),
    )


def read_db(path: str):
    """Read the CSV database from the given file path.

    Args:
        path (str): The file path of the CSV database.

    Returns:
        list: A sorted list of tuples containing IP address ranges and their
        corresponding location.
    """
    with open(path, newline="") as file:
        return sorted(
            [ip_range_from_csv(row) for row in csv.reader(file)],
        )


def search_ip_range(database: list[IPRange], ip) -> IPRangeData | None:
    """Search for the IP range containing the given IP address.

    Args:
        database (list): Database of IP address ranges.

        ip (object): IP object of the address to lookup.

    Returns:
        IPRange | None: The IPRange object matching the given IP address, or
        None if the IP address is not found in the database.
    """
    i = bisect_right(database, ip, key=lambda x: x.start)
    if i:
        row = database[i - 1]
        if row.start <= ip <= row.end:
            return row.data
    return None


def country_code(databases, address) -> str | None:
    """Lookup the country of the given IP address in the given databases.

    Args:
        databases (tuple): Tuple of databases for IPv4 and IPv6 IP address.

        address (str): IP address to lookup.

    Returns:
        str | None: The country code of the given IP address, or None if the IP
        address is not found in the database.
    """
    ip = ip_address(address)
    db = databases[0] if ip.version == 4 else databases[1]

    match = search_ip_range(db, ip)
    return match.country_code if match else None


if __name__ == "__main__":
    if not os.path.exists(COUNTRY_IPV4_DB_FILE):
        print("Downloading the IPv4 database...")
        download_file(COUNTRY_IPV4_DB_URL, COUNTRY_IPV4_DB_FILE, 10)

    if not os.path.exists(COUNTRY_IPV6_DB_FILE):
        print("Downloading the IPv6 database...")
        download_file(COUNTRY_IPV6_DB_URL, COUNTRY_IPV6_DB_FILE, 10)

    databases = (
        read_db(COUNTRY_IPV4_DB_FILE),
        read_db(COUNTRY_IPV6_DB_FILE),
    )

    print(country_code(databases, "62.35.85.135"))
    print(country_code(databases, "34.149.229.210"))
    print(country_code(databases, "142.251.220.163"))
    print(country_code(databases, "2a02:26f7:c9c8:4000:950e:981f:cef4:b0ed"))
