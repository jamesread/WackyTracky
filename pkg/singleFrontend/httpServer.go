package singleFrontend

import (
	"github.com/wacky-tracky/wacky-tracky-server/pkg/grpcapi"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/db/dummy"
)

func StartServers() {
	go grpcapi.Start(dummy.Dummy{})

	go startRestGateway()
	go startWebUIServer()

	startSingleFrontend()
}
