package singleFrontend

import (
	. "github.com/wacky-tracky/wacky-tracky-server/pkg/runtimeconfig"

	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func startSingleFrontend() {
	log.WithFields(log.Fields{
		"address": RuntimeConfig.ListenAddressSingleHTTPFrontend,
	}).Info("Starting single HTTP frontend")

	apiURL, _ := url.Parse("http://" + RuntimeConfig.ListenAddressRest)
	apiProxy := httputil.NewSingleHostReverseProxy(apiURL)

	webuiURL, _ := url.Parse("http://" + RuntimeConfig.ListenAddressWebUI)
	webuiProxy := httputil.NewSingleHostReverseProxy(webuiURL)

	mux := http.NewServeMux()

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("api req: %v", r.URL)
		apiProxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("ui req: %v", r.URL)
		webuiProxy.ServeHTTP(w, r)
	})

	srv := &http.Server{
		Addr:    RuntimeConfig.ListenAddressSingleHTTPFrontend,
		Handler: mux,
	}

	log.Panic(srv.ListenAndServe())
}
