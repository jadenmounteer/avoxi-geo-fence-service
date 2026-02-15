package api

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
	"github.com/jadenmounteer/avoxi-geo-fence/internal/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GeoFenceServer implements pb.GeoFenceServiceServer.
type GeoFenceServer struct {
	pb.UnimplementedGeoFenceServiceServer
	checker *geofence.Checker
}

// NewGeoFenceServer creates a GeoFenceServer with the given Checker.
func NewGeoFenceServer(checker *geofence.Checker) *GeoFenceServer {
	return &GeoFenceServer{checker: checker}
}

// CheckAccess checks whether the given IP is in one of the allowed countries.
func (s *GeoFenceServer) CheckAccess(ctx context.Context, req *pb.CheckRequest) (*pb.CheckResponse, error) {
	result, err := s.checker.Check(req.GetIpAddress(), req.GetAllowedCountries())
	if err != nil {
		if errors.Is(err, geofence.ErrUnknownIP) {
			return &pb.CheckResponse{Allowed: false, Country: ""}, nil
		}
		if errors.Is(err, geofence.ErrEmptyAllowedCountries) || errors.Is(err, geofence.ErrInvalidIP) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		slog.Error("check failed", "err", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.CheckResponse{Allowed: result.Allowed, Country: result.Country}, nil
}
