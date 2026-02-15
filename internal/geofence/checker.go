package geofence

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// ErrEmptyAllowedCountries is returned when allowed_countries is empty.
var ErrEmptyAllowedCountries = errors.New("allowed_countries must contain at least one country")

// ErrInvalidIP is returned when the IP address string cannot be parsed.
var ErrInvalidIP = errors.New("invalid IP")

// CountryLookuper provides IP-to-country lookup. GeoStore implements this interface.
type CountryLookuper interface {
	Lookup(ip net.IP) (string, error)
}

// CheckResult holds the geo-fencing decision and metadata for logging.
type CheckResult struct {
	Allowed bool   // true if the IP's country is in the allowed list
	Country string // the ISO country code found (empty if unknown)
}

// Checker validates IP addresses against an allowed list of countries.
type Checker struct {
	lookup CountryLookuper
}

// NewChecker creates a Checker with the given country lookup dependency.
func NewChecker(lookup CountryLookuper) *Checker {
	return &Checker{lookup: lookup}
}

// Check determines whether the given IP address is in one of the allowed countries.
// It parses IPv4 and IPv6 addresses, looks up the country, and compares case-insensitively.
// Returns an error for malformed IP strings, empty allowed list, or when the IP is not found in the database.
func (c *Checker) Check(ipStr string, allowedCountries []string) (CheckResult, error) {
	if len(allowedCountries) == 0 {
		return CheckResult{}, ErrEmptyAllowedCountries
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return CheckResult{}, fmt.Errorf("%w: %s", ErrInvalidIP, ipStr)
	}

	country, err := c.lookup.Lookup(ip)
	if err != nil {
		if errors.Is(err, ErrUnknownIP) {
			return CheckResult{Allowed: false, Country: ""}, fmt.Errorf("lookup: %w", err)
		}
		return CheckResult{}, fmt.Errorf("lookup: %w", err)
	}

	countryUpper := strings.ToUpper(country)
	for _, allowed := range allowedCountries {
		if strings.ToUpper(allowed) == countryUpper {
			return CheckResult{Allowed: true, Country: country}, nil
		}
	}
	return CheckResult{Allowed: false, Country: country}, nil
}
