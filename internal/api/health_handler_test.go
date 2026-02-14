package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
)

func TestHealthHandler_Liveness(t *testing.T) {
	handler := NewHealthHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.Liveness(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	got := strings.TrimSuffix(rec.Body.String(), "\n")
	if got != `{"status":"up"}` {
		t.Errorf("body = %q, want %q", got, `{"status":"up"}`)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestHealthHandler_Liveness_MethodNotAllowed(t *testing.T) {
	handler := NewHealthHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()
	handler.Liveness(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestHealthHandler_Ready_WithStore(t *testing.T) {
	dbPath := filepath.Join("..", "..", "data", "GeoLite2-Country.mmdb")
	store, err := geofence.NewGeoStore(dbPath)
	if err != nil {
		t.Skipf("GeoLite2-Country.mmdb not found at %s; skip readiness test with store", dbPath)
	}
	defer store.Close()

	handler := NewHealthHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	handler.Ready(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	got := strings.TrimSuffix(rec.Body.String(), "\n")
	if got != `{"status":"ready"}` {
		t.Errorf("body = %q, want %q", got, `{"status":"ready"}`)
	}
}

func TestHealthHandler_Ready_NilStore(t *testing.T) {
	handler := NewHealthHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	rec := httptest.NewRecorder()
	handler.Ready(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	var errResp ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &errResp); err != nil {
		t.Errorf("expected JSON error body: %v", err)
	}
	if errResp.Error == "" {
		t.Error("expected error field in response")
	}
}
