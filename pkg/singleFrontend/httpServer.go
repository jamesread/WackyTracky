package singleFrontend

import (
	"github.com/wacky-tracky/wacky-tracky-server/pkg/grpcapi"
)

func StartServers() {
	go grpcapi.Start()

	startSingleFrontend()
}
