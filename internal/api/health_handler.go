package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
)

// HealthHandler serves liveness and readiness probes.
type HealthHandler struct {
	store *geofence.GeoStore
}

// NewHealthHandler creates a HealthHandler with the given GeoStore.
func NewHealthHandler(store *geofence.GeoStore) *HealthHandler {
	return &HealthHandler{store: store}
}

// Liveness returns 200 if the server is alive. Used by Kubernetes liveness probe.
func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: "up"})
}

// Ready returns 200 only if the GeoIP database is loaded. Used by Kubernetes readiness probe.
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if h.store == nil {
		slog.Warn("readiness check failed", "reason", "database handle is nil")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "database not ready"})
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ready"})
}

// HealthResponse is the JSON body for health endpoints.
type HealthResponse struct {
	Status string `json:"status"`
}
