package httpserver

import (
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/wacky-tracky/wacky-tracky-server/internal/clientapi"
	"github.com/wacky-tracky/wacky-tracky-server/internal/frontend"

	cors "github.com/rs/cors"
	log "github.com/sirupsen/logrus"

	"github.com/wacky-tracky/wacky-tracky-server/internal/db"
)

func StartServer() {
	mux := http.NewServeMux()

	db := db.GetNewDatabaseConnection()

	api := clientapi.GetNewClientAPI(db)

	apiPath, apiHandler := api.GetNewHandler()
	apiHandler = http.StripPrefix("/api", apiHandler)
	apiHandler = withCors(apiHandler)

	mux.HandleFunc("/api"+apiPath, func(w http.ResponseWriter, r *http.Request) {
		apiHandler.ServeHTTP(w, r)
	})
	mux.Handle("/", http.StripPrefix("/", frontend.GetNewHandler()))
	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/wallpapers", getWallpapersHandler())

	endpoint := "0.0.0.0:8080"

	log.Infof("Starting HTTP server on %s", endpoint)
	log.Infof("API available at %s", endpoint+"/api"+apiPath)

	if err := http.ListenAndServe(
		endpoint,
		h2c.NewHandler(mux, &http2.Server{}),
	); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func findWallpaperDir() string {
	return "/var/www/html/wallpapers/"
}

func getWallpapersHandler() http.Handler {
	wallpaperDir := findWallpaperDir()

	log.WithFields(log.Fields{
		"wallpaperDir": wallpaperDir,
	}).Info("Starting wallpaper server")

	return http.StripPrefix("/wallpapers/", http.FileServer(http.Dir(wallpaperDir)))
}

func withCors(h http.Handler) http.Handler {
	return cors.AllowAll().Handler(h)
}
