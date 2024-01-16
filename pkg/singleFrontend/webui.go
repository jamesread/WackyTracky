package singleFrontend

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
)

func findWebuiDir() string {
	directoriesToSearch := []string{
		"../wacky-tracky-client-html5/dist/",
		"./webui",
		"/webui",
	}

	for _, dir := range directoriesToSearch {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			log.WithFields(log.Fields{
				"dir": dir,
			}).Infof("Found the webui directory")

			return dir
		}
	}

	log.Warnf("Did not find the webui directory! You will get 404 errors")

	return "./webui" // should not exist
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
