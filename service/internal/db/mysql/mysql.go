package mysql

import (
	//	"github.com/go-sql-driver/mysql"

	"github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
)

type MySQLConnector struct {
	db.DB
}

func (self MySQLConnector) Connect() error {
	return nil
}
