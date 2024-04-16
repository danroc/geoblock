"""IP Geolocation Blocker."""

__version__ = "0.1.0"

import csv
from bisect import bisect_right
from dataclasses import dataclass
from ipaddress import IPv4Address, IPv6Address, ip_address
from typing import Generic, NewType, TypeGuard, TypeVar

import requests


CountryCode = NewType("CountryCode", str)


AddressType = TypeVar("AddressType", bound=IPv4Address | IPv6Address)


@dataclass
class RangeData:
    """IP Range Data class."""

    country_code: CountryCode


@dataclass(order=True)
class RangeEntry(Generic[AddressType]):
    """IP Range class."""

    start: AddressType
    end: AddressType
    data: RangeData


RowType = RangeEntry[IPv4Address] | RangeEntry[IPv6Address]


DatabaseType = list[RangeEntry[IPv4Address]] | list[RangeEntry[IPv6Address]]


@dataclass
class Databases:
    v4: list[RangeEntry[IPv4Address]]
    v6: list[RangeEntry[IPv6Address]]


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


def ip_range_from_csv(row) -> RangeEntry[IPv4Address] | RangeEntry[IPv6Address]:
    """Create an IPRange object from a CSV row.

    Args:
        row (list): A row from a CSV file containing an IP address range and a
        country code.

    Returns:
        IPRange: An IPRange object.
    """
    start, end, *data = row

    start_ip = ip_address(row[0])
    end_ip = ip_address(row[1])

    if start_ip.version == 4 and end_ip.version == 4:
        return RangeEntry[IPv4Address](start, end, RangeData(*data))

    if start_ip.version == 6 and end_ip.version == 6:
        return RangeEntry[IPv6Address](start, end, RangeData(*data))

    raise ValueError("IP address versions do not match")


def all_same_version(ip_ranges: list[RowType]) -> TypeGuard[DatabaseType]:
    """Check if all IP ranges in the list have the same IP version.

    Args:
        ip_ranges (list): A list of IPRange objects.

    Returns:
        bool: True if all IP ranges have the same IP version, False otherwise.
    """
    first = ip_ranges[0]
    return all(
        first.start.version == ip.start.version for ip in ip_ranges
    ) and (first.start.version == 4 or first.start.version == 6)


def read_db(path: str) -> DatabaseType:
    """Read the CSV database from the given file path.

    Args:
        path (str): The file path of the CSV database.

    Returns:
        list: A sorted list of tuples containing IP address ranges and their
        corresponding location.
    """
    with open(path, newline="") as file:
        db = sorted(
            [ip_range_from_csv(row) for row in csv.reader(file)],
        )
        if not all_same_version(db):
            raise ValueError("IP address versions do not match")
        return db


def search_ip_range(database: DatabaseType, ip) -> RangeData | None:
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


def country_code(databases: Databases, address) -> str | None:
    """Lookup the country of the given IP address in the given databases.

    Args:
        databases (tuple): Tuple of databases for IPv4 and IPv6 IP address.

        address (str): IP address to lookup.

    Returns:
        str | None: The country code of the given IP address, or None if the IP
        address is not found in the database.
    """
    ip = ip_address(address)
    db: DatabaseType = databases.v4 if ip.version == 4 else databases.v6

    match = search_ip_range(db, ip)
    return match.country_code if match else None
