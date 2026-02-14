package api

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
)

type mockLookuper struct {
	lookup func(net.IP) (string, error)
}

func (m mockLookuper) Lookup(ip net.IP) (string, error) {
	return m.lookup(ip)
}

func TestCheckHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           string
		mockLookup     func(net.IP) (string, error)
		wantStatus     int
		wantBody       string
		wantErrInBody  bool
		checkContentType bool
	}{
		{
			name:   "POST allowed IP",
			method: http.MethodPost,
			body:   `{"ip_address":"8.8.8.8","allowed_countries":["US","CA"]}`,
			mockLookup: func(net.IP) (string, error) { return "US", nil },
			wantStatus:  http.StatusOK,
			wantBody:    `{"allowed":true,"country":"US"}`,
			checkContentType: true,
		},
		{
			name:   "POST blocked IP",
			method: http.MethodPost,
			body:   `{"ip_address":"8.8.8.8","allowed_countries":["GB"]}`,
			mockLookup: func(net.IP) (string, error) { return "US", nil },
			wantStatus:     http.StatusOK,
			wantBody:       `{"allowed":false,"country":"US"}`,
			checkContentType: true,
		},
		{
			name:            "POST invalid JSON",
			method:          http.MethodPost,
			body:            `{invalid}`,
			wantStatus:      http.StatusBadRequest,
			wantErrInBody:   true,
			checkContentType: true,
		},
		{
			name:            "POST invalid IP",
			method:          http.MethodPost,
			body:            `{"ip_address":"not-an-ip","allowed_countries":["US"]}`,
			mockLookup:      func(net.IP) (string, error) { return "", nil },
			wantStatus:      http.StatusBadRequest,
			wantErrInBody:   true,
			checkContentType: true,
		},
		{
			name:            "GET returns 405",
			method:          http.MethodGet,
			body:            "",
			wantStatus:      http.StatusMethodNotAllowed,
			wantErrInBody:   true,
			checkContentType: true,
		},
		{
			name:   "POST unknown IP returns 200 with allowed false",
			method: http.MethodPost,
			body:   `{"ip_address":"192.168.1.1","allowed_countries":["US"]}`,
			mockLookup: func(net.IP) (string, error) { return "", geofence.ErrUnknownIP },
			wantStatus:     http.StatusOK,
			wantBody:       `{"allowed":false,"country":""}`,
			checkContentType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookup := mockLookuper{lookup: func(ip net.IP) (string, error) { return "", nil }}
			if tt.mockLookup != nil {
				lookup = mockLookuper{lookup: tt.mockLookup}
			}
			checker := geofence.NewChecker(lookup)
			handler := NewCheckHandler(checker)

			req := httptest.NewRequest(tt.method, "/v1/check", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" {
				got := strings.TrimSuffix(rec.Body.String(), "\n")
				if got != tt.wantBody {
					t.Errorf("body = %q, want %q", got, tt.wantBody)
				}
			}
			if tt.wantErrInBody {
				var errResp ErrorResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
					t.Errorf("expected JSON error body, got %q", rec.Body.String())
				}
				if errResp.Error == "" {
					t.Errorf("expected error field in response")
				}
			}
			if tt.checkContentType {
				if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
					t.Errorf("Content-Type = %q, want application/json", ct)
				}
			}
		})
	}
}
