package singleFrontend

import (
	. "github.com/wacky-tracky/wacky-tracky-server/pkg/runtimeconfig"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/wacky-tracky/wacky-tracky-server/pkg/grpcapi"

	"github.com/wacky-tracky/wacky-tracky-server/pkg/db/dummy"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/db/neo4j"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/db"

)

func StartServers() {
	go grpcapi.Start(getNewDatabaseConnection())
	go startRestGateway()
	go startWebUIServer()

	startSingleFrontend()
}

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

	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:    RuntimeConfig.ListenAddressSingleHTTPFrontend,
		Handler: mux,
	}

	log.Panic(srv.ListenAndServe())
}

func getNewDatabaseConnection() db.DB {
	log.WithFields(log.Fields{
		"Driver": RuntimeConfig.Database.Driver,
	}).Infof("DB Backend")

	switch RuntimeConfig.Database.Driver {
	case "neo4j":
		return neo4j.Neo4jDB{}
	default:
		return dummy.Dummy{}
	}
}

