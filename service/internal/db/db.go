package db

import (
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/dummy"
	db "github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/neo4j"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/todotxt"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/yamlfiles"
	. "github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"

	log "github.com/sirupsen/logrus"
)

func GetNewDatabaseConnection() db.DB {
	log.WithFields(log.Fields{
		"Driver": RuntimeConfig.Database.Driver,
	}).Infof("DB Backend")

	switch RuntimeConfig.Database.Driver {
	case "neo4j":
		return &neo4j.Neo4jDB{}
	case "todotxt":
		return &todotxt.TodoTxt{}
	case "yamlfiles":
		return &yamlfiles.YamlFileDriver{}
	default:
		return &dummy.Dummy{}
	}
}
