package singleFrontend

import (
	log "github.com/sirupsen/logrus"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/db/dummy"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/db/neo4j"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/grpcapi"
)

func StartServers(db string) {
	log.Infof("DB Backend: %v", db)

	switch db {
	case "neo4j":
		dbi := neo4j.Neo4jDB{}

		log.Infof("dbi: %+v", dbi)

		go grpcapi.Start(dbi)
	default:
		go grpcapi.Start(dummy.Dummy{})
	}

	go startRestGateway()
	go startWebUIServer()

	startSingleFrontend()
}
