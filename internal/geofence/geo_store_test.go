package geofence

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestNewGeoStore_InvalidPath(t *testing.T) {
	_, err := NewGeoStore("/nonexistent/path/GeoLite2-Country.mmdb")
	if err == nil {
		t.Fatal("expected error for non-existent path, got nil")
	}
}

func TestGeoStore_Lookup(t *testing.T) {
	dbPath := filepath.Join("..", "..", "data", "GeoLite2-Country.mmdb")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Skipf("GeoLite2-Country.mmdb not found at %s; skip Lookup tests", dbPath)
	}

	store, err := NewGeoStore(dbPath)
	if err != nil {
		t.Fatalf("NewGeoStore: %v", err)
	}
	defer store.Close()

	tests := []struct {
		name        string
		ip          net.IP
		wantCountry string
		wantErr     error
		expectAnyErr bool // if true, just verify err != nil
	}{
		{
			name:        "known public IP 8.8.8.8",
			ip:          net.ParseIP("8.8.8.8"),
			wantCountry: "US",
			wantErr:     nil,
		},
		{
			name:        "unknown private IP 192.168.1.1",
			ip:          net.ParseIP("192.168.1.1"),
			wantCountry: "",
			wantErr:     ErrUnknownIP,
		},
		{
			name:         "invalid nil IP",
			ip:           nil,
			wantCountry:  "",
			expectAnyErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			country, err := store.Lookup(tt.ip)
			if tt.expectAnyErr {
				if err == nil {
					t.Fatal("expected error for nil IP, got nil")
				}
				return
			}
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if country != tt.wantCountry {
				t.Errorf("country = %q, want %q", country, tt.wantCountry)
			}
		})
	}
}
