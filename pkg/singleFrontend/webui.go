package singleFrontend

import (
	"net/http"
	log "github.com/sirupsen/logrus"
)

func findWebuiDir() string {
	return "../wacky-tracky-client-html5/dist/"
}

func startWebUIServer() {
	log.WithFields(log.Fields{
		"address": "0.0.0.0:8084",
	}).Info("Starting WebUI server")

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(findWebuiDir())))
//	mux.HandleFunc("/webUiSettings.json", generateWebUISettings)

	srv := &http.Server{
		Addr:    "0.0.0.0:8084",
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
