package database

import (
	"net"
	"strconv"

	"github.com/danroc/geoblock/pkg/utils"
)

const (
	countryIPv4URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv4.csv"
	countryIPv6URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-country/geolite2-country-ipv6.csv"
	asnIPv4URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-asn/geolite2-asn-ipv4.csv"
	asnIPv6URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-asn/geolite2-asn-ipv6.csv"
)

// ReservedAS0 is the ASN used when the ASN is unknown. Its value is 0.
const ReservedAS0 uint32 = 0

// Resolution contains the result of resolving an IP address.
type Resolution struct {
	CountryCode  string
	Organization string
	ASN          uint32
}

// Resolver is an IP resolver that returns information about an IP address.
type Resolver struct {
	countryDBv4 *Database
	countryDBv6 *Database
	asnDBv4     *Database
	asnDBv6     *Database
}

// NewResolver creates a new IP resolver.
func NewResolver() (*Resolver, error) {
	countryDBv4, err := NewDatabase(countryIPv4URL)
	if err != nil {
		return nil, err
	}

	countryDBv6, err := NewDatabase(countryIPv6URL)
	if err != nil {
		return nil, err
	}

	asnDBv4, err := NewDatabase(asnIPv4URL)
	if err != nil {
		return nil, err
	}

	asnDBv6, err := NewDatabase(asnIPv6URL)
	if err != nil {
		return nil, err
	}

	return &Resolver{
		countryDBv4: countryDBv4,
		countryDBv6: countryDBv6,
		asnDBv4:     asnDBv4,
		asnDBv6:     asnDBv6,
	}, nil
}

// strIndex returns the element at the given index of the data slice. If the
// index is out of bounds, the function returns an empty string.
func strIndex(data []string, index int) string {
	if index < 0 || index >= len(data) {
		return ""
	}
	return data[index]
}

// strToASN converts a string to an ASN. If the string is not a valid ASN, the
// function returns ReservedAS0.
func strToASN(s string) uint32 {
	asn, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return ReservedAS0
	}
	return uint32(asn)
}

// resolve checks the given IP address against the country and ASN databases.
func resolve(ip net.IP, countryDB *Database, asnDB *Database) *Resolution {
	var (
		countryMatch = countryDB.Find(ip)
		asnMatch     = asnDB.Find(ip)
	)
	return &Resolution{
		CountryCode:  strIndex(countryMatch, 0),
		Organization: strIndex(asnMatch, 1),
		ASN:          strToASN(strIndex(asnMatch, 0)),
	}
}

// Resolve resolves the given IP address to a country code and an ASN.
//
// It is the caller's responsibility to check if the IP is valid.
//
// If the country of the IP is not found, the CountryCode field of the result
// will be an empty string. If the ASN of the IP is not found, the ASN field of
// the result will be zero.
//
// The Organization field is present for informational purposes only. It is not
// used by the rules engine.
func (r *Resolver) Resolve(ip net.IP) *Resolution {
	if utils.IsIPv4(ip) {
		return resolve(ip, r.countryDBv4, r.asnDBv4)
	}
	return resolve(ip, r.countryDBv6, r.asnDBv6)
}
