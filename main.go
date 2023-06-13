package main

import (
	"github.com/wacky-tracky/wacky-tracky-server/pkg/singleFrontend"
	. "github.com/wacky-tracky/wacky-tracky-server/pkg/runtimeconfig"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("wacky-tracky")

	log.WithFields(log.Fields{
		"ListenAddress": RuntimeConfig.ListenAddressSingleHTTPFrontend,
	}).Infof("config")

	singleFrontend.StartServers()
}
