package singleFrontend

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func findWebuiDir() string {
	return "../wacky-tracky-client-html5/src/"
}

func findWallpaperDir() string {
	return "/var/www/html/wallpapers/"
}

func startWebUIServer() {
	uidir := findWebuiDir()
	wallpaperDir := findWallpaperDir()

	log.WithFields(log.Fields{
		"address":      "0.0.0.0:8084",
		"uidir":        uidir,
		"wallpaperDir": wallpaperDir,
	}).Info("Starting WebUI server")

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(uidir)))
	mux.Handle("/wallpapers/", http.FileServer(http.Dir(wallpaperDir)))

	srv := &http.Server{
		Addr:    "0.0.0.0:8084",
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
