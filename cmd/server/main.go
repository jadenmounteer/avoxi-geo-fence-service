package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jadenmounteer/avoxi-geo-fence/internal/api"
	"github.com/jadenmounteer/avoxi-geo-fence/internal/geofence"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	store, err := geofence.NewGeoStore(geofence.DefaultDBPath)
	if err != nil {
		log.Fatal(err)
	}

	checker := geofence.NewChecker(store)
	handler := api.NewCheckHandler(checker)
	mux := http.NewServeMux()
	mux.Handle("/v1/check", api.LoggingMiddleware(handler))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	slog.Info("server starting", "port", port)

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
