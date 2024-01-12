package grpcapi

import (
	pb "github.com/wacky-tracky/wacky-tracky-server/gen/grpc"
	db "github.com/wacky-tracky/wacky-tracky-server/pkg/db"
	"google.golang.org/grpc"
	"net"

	"context"

	log "github.com/sirupsen/logrus"

	. "github.com/wacky-tracky/wacky-tracky-server/pkg/runtimeconfig"
)

var dbconn db.DB

type wackyTrackyClientApi struct {
	// pb.UnimplementedWackyTrackyClientApiServer
}

func Start(newdb db.DB) {
	dbconn = newdb
	dbconn.Connect()

	log.WithFields(log.Fields{
		"address": RuntimeConfig.ListenAddressGrpc,
	}).Infof("Starting GRPC API")

	lis, err := net.Listen("tcp", RuntimeConfig.ListenAddressGrpc)

	if err != nil {
		log.Fatalf("Failed to listen - %v", err)
		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWackyTrackyClientApiServer(grpcServer, newServer())

	go grpcServer.Serve(lis)

	dbconn.GetTasks(418)
}

func (api *wackyTrackyClientApi) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	items, err := dbconn.GetTasks(req.ListID)

	ret := &pb.ListTasksResponse{}

	if err != nil {
		return ret, err
	}

	for _, item := range items {
		ret.Tasks = append(ret.Tasks, &pb.Task{
			ID:         item.ID,
			Content:    item.Content,
			ParentId:   item.ParentId,
			ParentType: item.ParentType,
		})
	}

	return ret, nil
}

func (api *wackyTrackyClientApi) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	dbconn.CreateTask(req.Content)

	res := &pb.CreateTaskResponse{}

	return res, nil
}

func (api *wackyTrackyClientApi) CreateList(ctx context.Context, req *pb.CreateListRequest) (*pb.CreateListResponse, error) {
	return nil, nil
}

func (api *wackyTrackyClientApi) GetLists(ctx context.Context, req *pb.GetListsRequest) (*pb.GetListsResponse, error) {
	lists, _ := dbconn.GetLists()

	res := &pb.GetListsResponse{}

	for _, dblist := range lists {
		l := &pb.List{
			Title:      dblist.Title,
			ID:         dblist.ID,
			CountTasks: dblist.CountTasks,
		}

		res.Lists = append(res.Lists, l)
	}

	return res, nil
}

func (api *wackyTrackyClientApi) GetTags(ctx context.Context, req *pb.GetTagsRequest) (*pb.GetTagsResponse, error) {
	tags, _ := dbconn.GetTags()

	res := &pb.GetTagsResponse{}

	for _, dbtag := range tags {
		t := &pb.Tag{
			ID:    dbtag.ID,
			Title: dbtag.Title,
		}

		res.Tags = append(res.Tags, t)
	}

	return res, nil
}

func (api *wackyTrackyClientApi) Tag(ctx context.Context, req *pb.TagRequest) (*pb.TagResponse, error) {
	return nil, nil
}

func (api *wackyTrackyClientApi) UpdateList(ctx context.Context, req *pb.UpdateListRequest) (*pb.UpdateListResponse, error) {
	return nil, nil
}

func (api *wackyTrackyClientApi) Init(ctx context.Context, req *pb.InitRequest) (*pb.InitResponse, error) {
	res := &pb.InitResponse{
		Wallpaper: "wallpaper.jpg",
	}

	return res, nil
}

func newServer() *wackyTrackyClientApi {
	server := wackyTrackyClientApi{}
	return &server
}
