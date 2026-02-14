package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

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
	defer store.Close()

	checker := geofence.NewChecker(store)
	handler := api.NewCheckHandler(checker)
	mux := http.NewServeMux()
	mux.Handle("/v1/check", api.LoggingMiddleware(handler))

	slog.Info("server starting", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
