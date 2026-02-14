package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/api"
	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
)

const version = "1.0.0"

type config struct {
	port     string
	dbPath   string
	logLevel slog.Level
}

func loadConfig() config {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = os.Getenv("PORT")
	}
	if port == "" {
		port = "8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/GeoLite2-Country.mmdb"
	}
	level := parseLogLevel(os.Getenv("LOG_LEVEL"))
	return config{port: port, dbPath: dbPath, logLevel: level}
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
	handler := api.NewCheckHandler(checker)
	mux := http.NewServeMux()
	mux.Handle("/v1/check", api.LoggingMiddleware(handler))

	server := &http.Server{
		Addr:    ":" + cfg.port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server listen failed", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("server starting", "port", cfg.port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	slog.Info("shutting down gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown", "err", err)
	}
	if err := store.Close(); err != nil {
		slog.Error("close GeoStore", "err", err)
	}

	slog.Info("shutdown complete")
	os.Exit(0)
}
