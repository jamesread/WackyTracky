package db

import (
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/dummy"
	db "github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/todotxt"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/yamlfiles"
	"github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"

	log "github.com/sirupsen/logrus"
)

func GetNewDatabaseConnection() db.DB {
	log.WithFields(log.Fields{
		"Driver": runtimeconfig.RuntimeConfig.Database.Driver,
	}).Infof("DB Backend")

	switch runtimeconfig.RuntimeConfig.Database.Driver {
	case "todotxt":
		return &todotxt.TodoTxt{}
	case "yamlfiles":
		return &yamlfiles.YamlFileDriver{}
	default:
		return &dummy.Dummy{}
	}
}
