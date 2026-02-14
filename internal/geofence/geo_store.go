package geofence

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/netip"

	"github.com/oschwald/geoip2-golang/v2"
)

// ErrUnknownIP is returned when an IP address is not found in the GeoIP database
// (e.g., private ranges like 192.168.x.x or reserved addresses).
var ErrUnknownIP = errors.New("ip not found in database")

// DefaultDBPath is the default path to the MaxMind GeoLite2-Country database.
const DefaultDBPath = "data/GeoLite2-Country.mmdb"

// GeoStore encapsulates the MaxMind GeoIP reader and provides a clean interface
// for country lookups. It is safe for concurrent use.
type GeoStore struct {
	reader *geoip2.Reader
}

// NewGeoStore opens the GeoIP database at the given path and returns a GeoStore.
// It fails fast if the file is missing or corrupted.
func NewGeoStore(dbPath string) (*GeoStore, error) {
	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open geoip database: %w", err)
	}
	slog.Info("GeoIP database opened successfully", "path", dbPath)
	return &GeoStore{reader: reader}, nil
}

// Lookup returns the ISO 3166-1 alpha-2 country code (e.g., "US", "FR") for the
// given IP address. Returns ErrUnknownIP if the IP is not in the database.
func (g *GeoStore) Lookup(ip net.IP) (string, error) {
	if ip == nil {
		return "", fmt.Errorf("invalid IP: nil address")
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return "", fmt.Errorf("invalid IP: %s", ip.String())
	}

	record, err := g.reader.Country(addr)
	if err != nil {
		return "", fmt.Errorf("lookup country: %w", err)
	}
	if !record.HasData() {
		return "", ErrUnknownIP
	}
	return record.Country.ISOCode, nil
}

// Close releases the underlying database reader and any memory-mapped resources.
// Callers should invoke Close when the GeoStore is no longer needed (e.g., defer store.Close()).
func (g *GeoStore) Close() error {
	if g.reader == nil {
		return nil
	}
	return g.reader.Close()
}
