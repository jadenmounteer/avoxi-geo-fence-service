package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
)

// CheckHandler handles POST /v1/check requests.
type CheckHandler struct {
	checker *geofence.Checker
}

// NewCheckHandler creates a CheckHandler with the given Checker.
func NewCheckHandler(checker *geofence.Checker) *CheckHandler {
	return &CheckHandler{checker: checker}
}

// ServeHTTP implements http.Handler.
func (h *CheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
		return
	}

	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("failed to decode request body", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "malformed JSON"})
		return
	}

	result, err := h.checker.Check(req.IPAddress, req.AllowedCountries)
	if err != nil {
		if errors.Is(err, geofence.ErrUnknownIP) {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(CheckResponse{Allowed: false, Country: ""})
			return
		}
		if errors.Is(err, geofence.ErrEmptyAllowedCountries) || errors.Is(err, geofence.ErrInvalidIP) {
			slog.Info("validation error", "ip_address", req.IPAddress, "err", err)
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
			return
		}
		slog.Error("check failed", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "internal server error"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(CheckResponse{Allowed: result.Allowed, Country: result.Country})
}
