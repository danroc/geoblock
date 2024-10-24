package database

import (
	"net"

	"github.com/danroc/geoblock/pkg/utils"
)

const (
	countryIPv4URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-geo-whois-asn-country/geolite2-geo-whois-asn-country-ipv4.csv"
	countryIPv6URL = "https://cdn.jsdelivr.net/npm/@ip-location-db/geolite2-geo-whois-asn-country/geolite2-geo-whois-asn-country-ipv6.csv"
	asnIPv4URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/asn/asn-ipv4.csv"
	asnIPv6URL     = "https://cdn.jsdelivr.net/npm/@ip-location-db/asn/asn-ipv6.csv"
)

type Resolution struct {
	IP           net.IP
	CountryCode  string
	ASN          string
	Organization string
}

type Resolver struct {
	countryDBv4 *Database
	countryDBv6 *Database
	asnDBv4     *Database
	asnDBv6     *Database
}

// NewResolver creates a new resolver.
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

func getIndex(data []string, index int) string {
	if data == nil {
		return ""
	}
	if index >= len(data) {
		return ""
	}
	return data[index]
}

func resolve(ip net.IP, countryDB *Database, asnDB *Database) *Resolution {
	var (
		countryMatch = countryDB.Find(ip)
		asnMatch     = asnDB.Find(ip)
	)
	return &Resolution{
		IP:           ip,
		CountryCode:  getIndex(countryMatch, 0),
		ASN:          getIndex(asnMatch, 0),
		Organization: getIndex(asnMatch, 1),
	}
}

// Resolve resolves the given IP address to a country code and an ASN.
func (r *Resolver) Resolve(ip net.IP) *Resolution {
	// Nothing to do if the IP is nil, it is the caller's responsibility to
	// check if the IP is valid.
	if ip == nil {
		return nil
	}

	if utils.IsIPv4(ip) {
		return resolve(ip, r.countryDBv4, r.asnDBv4)
	} else {
		return resolve(ip, r.countryDBv6, r.asnDBv6)
	}
}
