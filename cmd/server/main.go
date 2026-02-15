package main

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/api"
	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
	"github.com/jadenmounteer/avoxi-geo-fence/internal/pb"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const version = "1.0.0"

type config struct {
	httpPort string
	grpcPort string
	dbPath   string
	logLevel slog.Level
}

func loadConfig() config {
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = os.Getenv("APP_PORT")
	}
	if httpPort == "" {
		httpPort = os.Getenv("PORT")
	}
	if httpPort == "" {
		httpPort = "8080"
	}
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = "9090"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/GeoLite2-Country.mmdb"
	}
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	return config{httpPort: httpPort, grpcPort: grpcPort, dbPath: dbPath, logLevel: level}
}

func parseLogLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func main() {
	cfg := loadConfig()

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.logLevel})
	logger := slog.New(jsonHandler).With("service", "geo-fence-service", "version", version)
	slog.SetDefault(logger)

	if _, err := os.Stat(cfg.dbPath); os.IsNotExist(err) {
		slog.Error("database file does not exist", "path", cfg.dbPath)
		os.Exit(1)
	}

	store, err := geofence.NewGeoStore(cfg.dbPath)
	if err != nil {
		slog.Error("failed to open GeoIP database", "err", err)
		os.Exit(1)
	}

	checker := geofence.NewChecker(store)
	healthHandler := api.NewHealthHandler(store)

	mux := http.NewServeMux()
	mux.Handle("/v1/check", api.LoggingMiddleware(api.NewCheckHandler(checker)))
	mux.HandleFunc("/health", healthHandler.Liveness)
	mux.HandleFunc("/ready", healthHandler.Ready)

	httpServer := &http.Server{
		Addr:    ":" + cfg.httpPort,
		Handler: mux,
	}

	grpcServer := grpc.NewServer()
	pb.RegisterGeoFenceServiceServer(grpcServer, api.NewGeoFenceServer(checker))
	pb.RegisterHealthServiceServer(grpcServer, healthHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":"+cfg.grpcPort)
	if err != nil {
		slog.Error("gRPC listen failed", "err", err)
		os.Exit(1)
	}

	g, _ := errgroup.WithContext(context.Background())
	g.Go(func() error {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	g.Go(func() error {
		return grpcServer.Serve(lis)
	})

	slog.Info("server starting", "http_port", cfg.httpPort, "grpc_port", cfg.grpcPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	slog.Info("shutting down gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.GracefulStop()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("http server shutdown", "err", err)
	}
	if err := store.Close(); err != nil {
		slog.Error("close GeoStore", "err", err)
	}

	if err := g.Wait(); err != nil {
		slog.Error("server error", "err", err)
	}

	slog.Info("shutdown complete")
	os.Exit(0)
}
