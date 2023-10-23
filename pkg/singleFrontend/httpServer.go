package singleFrontend

import (
	log "github.com/sirupsen/logrus"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/db/dummy"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/db/neo4j"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/grpcapi"
)

func StartServers(db string) {
	log.Infof("DB: %v", db)

	switch db {
	case "neo4j":
		go grpcapi.Start(neo4j.Neo4jDB{})
	default:
		go grpcapi.Start(dummy.Dummy{})
	}

	go startRestGateway()
	go startWebUIServer()

	startSingleFrontend()
}
