"""IP Geolocation Blocker."""

__version__ = "0.1.0"

import csv
from bisect import bisect_right
from ipaddress import ip_address

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
        timeout respectively. Defaults to None.

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
            [
                (ip_address(row[0]), ip_address(row[1]), row[2])
                for row in csv.reader(file)
            ],
        )


def lookup_ip(db, ip):
    """Lookup the country of the given IP address in the given database.

    Args:
        db (list): Database of IP address ranges and their corresponding
        country codes.

        ip (object): IP object of the address to lookup.

    Returns:
        str | None: The country code of the given IP address, or None if the IP
        address is not found in the database.
    """
    i = bisect_right(db, ip, key=lambda x: x[0])
    if i:
        row = db[i - 1]
        if row[0] <= ip <= row[1]:
            return row[2]
    return None


if __name__ == "__main__":
    download_file(COUNTRY_IPV4_DB_URL, COUNTRY_IPV4_DB_FILE, 10)
    db = read_db(COUNTRY_IPV4_DB_FILE)
    print(lookup_ip(db, ip_address("62.35.85.135")))
    print(lookup_ip(db, ip_address("34.149.229.210")))
    print(lookup_ip(db, ip_address("142.251.220.163")))

    download_file(COUNTRY_IPV6_DB_URL, COUNTRY_IPV6_DB_FILE, 10)
    db = read_db(COUNTRY_IPV6_DB_FILE)
    print(lookup_ip(db, ip_address("2a02:26f7:c9c8:4000:950e:981f:cef4:b0ed")))
