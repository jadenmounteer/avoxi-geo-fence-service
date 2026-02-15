package api

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
	"github.com/jadenmounteer/avoxi-geo-fence/internal/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HealthHandler serves liveness and readiness probes for both HTTP and gRPC.
type HealthHandler struct {
	pb.UnimplementedHealthServiceServer
	store *geofence.GeoStore
}

// NewHealthHandler creates a HealthHandler with the given GeoStore.
func NewHealthHandler(store *geofence.GeoStore) *HealthHandler {
	return &HealthHandler{store: store}
}

// isReady returns whether the service is ready to accept traffic (store loaded).
// Shared by HTTP Ready and gRPC CheckHealth.
func (h *HealthHandler) isReady() (ready bool, msg string) {
	if h.store == nil {
		return false, "database not ready"
	}
	return true, "ready"
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
	ready, msg := h.isReady()
	if !ready {
		slog.Warn("readiness check failed", "reason", msg)
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: msg})
}

// CheckHealth implements gRPC HealthService.CheckHealth using the same readiness logic as Ready.
func (h *HealthHandler) CheckHealth(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	ready, msg := h.isReady()
	if !ready {
		slog.Warn("readiness check failed", "reason", msg)
		return nil, status.Error(codes.Unavailable, msg)
	}
	return &pb.HealthResponse{Status: msg}, nil
}

// HealthResponse is the JSON body for health endpoints.
type HealthResponse struct {
	Status string `json:"status"`
}
