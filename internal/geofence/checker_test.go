package geofence

import (
	"errors"
	"net"
	"testing"
)

type mockLookuper struct {
	lookup func(net.IP) (string, error)
}

func (m mockLookuper) Lookup(ip net.IP) (string, error) {
	return m.lookup(ip)
}

func TestChecker_Check(t *testing.T) {
	tests := []struct {
		name             string
		ipStr            string
		allowedCountries []string
		mockLookup       func(net.IP) (string, error)
		wantAllowed      bool
		wantCountry      string
		wantErr          bool
		expectErrUnknown bool
	}{
		{
			name:             "successful match",
			ipStr:            "8.8.8.8",
			allowedCountries: []string{"US", "CA"},
			mockLookup:       func(net.IP) (string, error) { return "US", nil },
			wantAllowed:      true,
			wantCountry:      "US",
			wantErr:          false,
		},
		{
			name:             "blocked match",
			ipStr:            "8.8.8.8",
			allowedCountries: []string{"GB"},
			mockLookup:       func(net.IP) (string, error) { return "US", nil },
			wantAllowed:      false,
			wantCountry:      "US",
			wantErr:          false,
		},
		{
			name:             "invalid IP",
			ipStr:            "not-an-ip",
			allowedCountries: []string{"US"},
			mockLookup:       nil,
			wantErr:          true,
		},
		{
			name:             "unknown IP",
			ipStr:            "192.168.1.1",
			allowedCountries: []string{"US"},
			mockLookup:       func(net.IP) (string, error) { return "", ErrUnknownIP },
			wantAllowed:      false,
			wantCountry:      "",
			wantErr:          true,
			expectErrUnknown: true,
		},
		{
			name:             "empty allowed list",
			ipStr:            "8.8.8.8",
			allowedCountries: []string{},
			mockLookup:       func(net.IP) (string, error) { return "US", nil },
			wantAllowed:      false,
			wantCountry:      "US",
			wantErr:          false,
		},
		{
			name:             "case insensitive",
			ipStr:            "8.8.8.8",
			allowedCountries: []string{"us"},
			mockLookup:       func(net.IP) (string, error) { return "US", nil },
			wantAllowed:      true,
			wantCountry:      "US",
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var lookup CountryLookuper
			if tt.mockLookup != nil {
				lookup = mockLookuper{lookup: tt.mockLookup}
			} else {
				lookup = mockLookuper{lookup: func(net.IP) (string, error) { return "", nil }}
			}
			checker := NewChecker(lookup)
			result, err := checker.Check(tt.ipStr, tt.allowedCountries)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.expectErrUnknown && !errors.Is(err, ErrUnknownIP) {
					t.Errorf("expected ErrUnknownIP, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Allowed != tt.wantAllowed {
				t.Errorf("Allowed = %v, want %v", result.Allowed, tt.wantAllowed)
			}
			if result.Country != tt.wantCountry {
				t.Errorf("Country = %q, want %q", result.Country, tt.wantCountry)
			}
		})
	}
}
