//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/api"
	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
	"github.com/jadenmounteer/avoxi-geo-fence/internal/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func TestIntegration_HTTPAndGRPC(t *testing.T) {
	dbPath := filepath.Join("data", "GeoLite2-Country.mmdb")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Skipf("GeoLite2-Country.mmdb not found at %s; skip integration test", dbPath)
	}

	store, err := geofence.NewGeoStore(dbPath)
	if err != nil {
		t.Fatalf("NewGeoStore: %v", err)
	}
	defer store.Close()

	checker := geofence.NewChecker(store)
	healthHandler := api.NewHealthHandler(store)

	mux := http.NewServeMux()
	mux.Handle("/v1/check", api.LoggingMiddleware(api.NewCheckHandler(checker)))
	mux.HandleFunc("/health", healthHandler.Liveness)
	mux.HandleFunc("/ready", healthHandler.Ready)

	httpLis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("http listen: %v", err)
	}
	defer httpLis.Close()

	grpcLis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("gRPC listen: %v", err)
	}
	defer grpcLis.Close()

	httpServer := &http.Server{Handler: mux}
	grpcServer := grpc.NewServer()
	pb.RegisterGeoFenceServiceServer(grpcServer, api.NewGeoFenceServer(checker))
	pb.RegisterHealthServiceServer(grpcServer, healthHandler)
	reflection.Register(grpcServer)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = httpServer.Serve(httpLis)
	}()
	go func() {
		defer wg.Done()
		_ = grpcServer.Serve(grpcLis)
	}()
	defer func() {
		grpcServer.GracefulStop()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(ctx)
		wg.Wait()
	}()

	time.Sleep(50 * time.Millisecond)

	httpBase := "http://" + httpLis.Addr().String()
	grpcAddr := grpcLis.Addr().String()

	// HTTP: GET /health
	resp, err := http.Get(httpBase + "/health")
	if err != nil {
		t.Fatalf("GET /health: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /health status = %d, want 200", resp.StatusCode)
	}
	body := strings.TrimSpace(readAll(t, resp.Body))
	if body != `{"status":"up"}` {
		t.Errorf("GET /health body = %q, want {\"status\":\"up\"}", body)
	}

	// HTTP: GET /ready
	resp2, err := http.Get(httpBase + "/ready")
	if err != nil {
		t.Fatalf("GET /ready: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Errorf("GET /ready status = %d, want 200", resp2.StatusCode)
	}
	body2 := strings.TrimSpace(readAll(t, resp2.Body))
	if body2 != `{"status":"ready"}` {
		t.Errorf("GET /ready body = %q, want {\"status\":\"ready\"}", body2)
	}

	// HTTP: POST /v1/check
	checkBody := `{"ip_address":"8.8.8.8","allowed_countries":["US","CA"]}`
	resp3, err := http.Post(httpBase+"/v1/check", "application/json", bytes.NewReader([]byte(checkBody)))
	if err != nil {
		t.Fatalf("POST /v1/check: %v", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Errorf("POST /v1/check status = %d, want 200", resp3.StatusCode)
	}
	var checkResp api.CheckResponse
	if err := json.NewDecoder(resp3.Body).Decode(&checkResp); err != nil {
		t.Fatalf("decode check response: %v", err)
	}
	if !checkResp.Allowed || checkResp.Country != "US" {
		t.Errorf("POST /v1/check got allowed=%v country=%q, want allowed=true country=US", checkResp.Allowed, checkResp.Country)
	}

	// gRPC: CheckHealth
	conn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	defer conn.Close()

	healthClient := pb.NewHealthServiceClient(conn)
	healthResp, err := healthClient.CheckHealth(context.Background(), &pb.HealthRequest{})
	if err != nil {
		t.Fatalf("CheckHealth: %v", err)
	}
	if healthResp.Status != "ready" {
		t.Errorf("CheckHealth status = %q, want ready", healthResp.Status)
	}

	// gRPC: CheckAccess
	geoClient := pb.NewGeoFenceServiceClient(conn)
	checkAccessResp, err := geoClient.CheckAccess(context.Background(), &pb.CheckRequest{
		IpAddress:        "8.8.8.8",
		AllowedCountries: []string{"US", "CA"},
	})
	if err != nil {
		t.Fatalf("CheckAccess: %v", err)
	}
	if !checkAccessResp.Allowed || checkAccessResp.Country != "US" {
		t.Errorf("CheckAccess got allowed=%v country=%q, want allowed=true country=US", checkAccessResp.Allowed, checkAccessResp.Country)
	}
}

func readAll(t *testing.T, r io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	return string(b)
}
