package api

import (
	"context"
	"net"
	"testing"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
	"github.com/jadenmounteer/avoxi-geo-fence/internal/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGeoFenceServer_CheckAccess(t *testing.T) {
	tests := []struct {
		name           string
		req            *pb.CheckRequest
		mockLookup     func(net.IP) (string, error)
		wantCode       codes.Code
		wantAllowed    bool
		wantCountry    string
	}{
		{
			name: "allowed IP",
			req: &pb.CheckRequest{
				IpAddress:        "8.8.8.8",
				AllowedCountries: []string{"US", "CA"},
			},
			mockLookup:  func(net.IP) (string, error) { return "US", nil },
			wantCode:    codes.OK,
			wantAllowed: true,
			wantCountry: "US",
		},
		{
			name: "blocked IP",
			req: &pb.CheckRequest{
				IpAddress:        "8.8.8.8",
				AllowedCountries: []string{"GB"},
			},
			mockLookup:  func(net.IP) (string, error) { return "US", nil },
			wantCode:    codes.OK,
			wantAllowed: false,
			wantCountry: "US",
		},
		{
			name: "unknown IP (private)",
			req: &pb.CheckRequest{
				IpAddress:        "192.168.1.1",
				AllowedCountries: []string{"US"},
			},
			mockLookup:  func(net.IP) (string, error) { return "", geofence.ErrUnknownIP },
			wantCode:    codes.OK,
			wantAllowed: false,
			wantCountry: "",
		},
		{
			name: "invalid IP",
			req: &pb.CheckRequest{
				IpAddress:        "not-an-ip",
				AllowedCountries: []string{"US"},
			},
			mockLookup: func(net.IP) (string, error) { return "", nil },
			wantCode:   codes.InvalidArgument,
		},
		{
			name: "empty allowed_countries",
			req: &pb.CheckRequest{
				IpAddress:        "8.8.8.8",
				AllowedCountries: []string{},
			},
			mockLookup: func(net.IP) (string, error) { return "US", nil },
			wantCode:   codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookup := mockLookuper{lookup: tt.mockLookup}
			checker := geofence.NewChecker(lookup)
			server := NewGeoFenceServer(checker)

			resp, err := server.CheckAccess(context.Background(), tt.req)

			if tt.wantCode == codes.OK {
				if err != nil {
					t.Fatalf("CheckAccess() unexpected error: %v", err)
				}
				if resp.Allowed != tt.wantAllowed {
					t.Errorf("Allowed = %v, want %v", resp.Allowed, tt.wantAllowed)
				}
				if resp.Country != tt.wantCountry {
					t.Errorf("Country = %q, want %q", resp.Country, tt.wantCountry)
				}
				return
			}

			if err == nil {
				t.Fatal("CheckAccess() expected error, got nil")
			}
			st, ok := status.FromError(err)
			if !ok {
				t.Fatalf("error is not a gRPC status: %v", err)
			}
			if st.Code() != tt.wantCode {
				t.Errorf("status code = %v, want %v", st.Code(), tt.wantCode)
			}
		})
	}
}
