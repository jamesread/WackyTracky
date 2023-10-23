package singleFrontend

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func findWebuiDir() string {
	return "../wacky-tracky-client-html5/src/"
}

func startWebUIServer() {
	uidir := findWebuiDir()

	log.WithFields(log.Fields{
		"address": "0.0.0.0:8084",
		"uidir":   uidir,
	}).Info("Starting WebUI server")

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(uidir)))
	//	mux.HandleFunc("/webUiSettings.json", generateWebUISettings)

	srv := &http.Server{
		Addr:    "0.0.0.0:8084",
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}
