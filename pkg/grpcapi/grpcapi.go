package grpcapi

import (
	"net"
	"google.golang.org/grpc"
	pb "github.com/wacky-tracky/wacky-tracky-server/gen/grpc"

	"context"

	log "github.com/sirupsen/logrus"

	. "github.com/wacky-tracky/wacky-tracky-server/pkg/runtimeconfig"
	
	"github.com/wacky-tracky/wacky-tracky-server/pkg/neo4j"
)

type wackyTrackyClientApi struct {
//	pb.UnimplementedWackyTrackyClientApiServer
}

func Start() {
	log.Infof("Starting API: %v", RuntimeConfig.ListenAddressGrpc)

	lis, err := net.Listen("tcp", RuntimeConfig.ListenAddressGrpc)

	if err != nil {
		log.Fatalf("Failed to listen - %v", err)
		return
	} else {
		log.Infof("Listening")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWackyTrackyClientApiServer(grpcServer, newServer())

	go grpcServer.Serve(lis)

	neo4j.GetItems(418)
}

func (api *wackyTrackyClientApi) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	return nil, nil
}

func (api *wackyTrackyClientApi) GetLists(ctx context.Context, req *pb.GetListsRequest) (*pb.GetListsResponse, error) {
	neo4j.GetLists()

	return nil, nil
}

func (api *wackyTrackyClientApi) Tag(ctx context.Context, req *pb.TagRequest) (*pb.TagResponse, error) {
	return nil, nil
}

func (api *wackyTrackyClientApi) UpdateList(ctx context.Context, req *pb.UpdateListRequest) (*pb.UpdateListResponse, error) {
	return nil, nil
}

func newServer() *wackyTrackyClientApi {
	server := wackyTrackyClientApi{}
	return &server
}
