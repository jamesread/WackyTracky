package grpcapi

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	pb "github.com/wacky-tracky/wacky-tracky-server/gen/grpc"
	"github.com/wacky-tracky/wacky-tracky-server/pkg/buildinfo"
	db "github.com/wacky-tracky/wacky-tracky-server/pkg/db"
	"google.golang.org/grpc"
	"net"

	"context"

	log "github.com/sirupsen/logrus"

	. "github.com/wacky-tracky/wacky-tracky-server/pkg/runtimeconfig"
)

var dbconn db.DB

var metricListTasksCount = promauto.NewCounter(prometheus.CounterOpts{
	Name: "list_tasks_count",
	Help: "The total number of ListTasks API calls",
})

type wackyTrackyClientService struct {
	// pb.UnimplementedWackyTrackyClientServiceServer
}

func Start(newdb db.DB) {
	dbconn = newdb
	dbconn.Connect()
	dbconn.Print()

	log.WithFields(log.Fields{
		"address": RuntimeConfig.ListenAddressGrpc,
	}).Infof("Starting GRPC API")

	lis, err := net.Listen("tcp", RuntimeConfig.ListenAddressGrpc)

	if err != nil {
		log.Fatalf("Failed to listen - %v", err)
		return
	}

	grpcServer := grpc.NewServer()
	pb.RegisterWackyTrackyClientServiceServer(grpcServer, newServer())

	go grpcServer.Serve(lis)

	dbconn.GetTasks(418)
}

func (api *wackyTrackyClientService) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	metricListTasksCount.Inc()

	items, err := dbconn.GetTasks(req.ListId)

	ret := &pb.ListTasksResponse{}

	if err != nil {
		return ret, err
	}

	for _, item := range items {
		ret.Tasks = append(ret.Tasks, &pb.Task{
			Id:            item.ID,
			Content:       item.Content,
			ParentId:      item.ParentId,
			ParentType:    item.ParentType,
			CountSubitems: item.CountSubitems,
		})
	}

	return ret, nil
}

func (api *wackyTrackyClientService) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.CreateTaskResponse, error) {
	dbconn.CreateTask(req.Content)

	res := &pb.CreateTaskResponse{}

	return res, nil
}

func (api *wackyTrackyClientService) CreateList(ctx context.Context, req *pb.CreateListRequest) (*pb.CreateListResponse, error) {
	return nil, nil
}

func (api *wackyTrackyClientService) GetLists(ctx context.Context, req *pb.GetListsRequest) (*pb.GetListsResponse, error) {
	lists, _ := dbconn.GetLists()

	res := &pb.GetListsResponse{}

	for _, dblist := range lists {
		l := &pb.List{
			Title:      dblist.Title,
			Id:         dblist.ID,
			CountItems: dblist.CountTasks,
		}

		res.Lists = append(res.Lists, l)
	}

	return res, nil
}

func (api *wackyTrackyClientService) Version(ctx context.Context, req *pb.VersionRequest) (*pb.VersionResponse, error) {
	return &pb.VersionResponse{
		Version: buildinfo.Version,
		Commit:  buildinfo.Commit,
		Date:    buildinfo.Date,
	}, nil
}

func (api *wackyTrackyClientService) GetTags(ctx context.Context, req *pb.GetTagsRequest) (*pb.GetTagsResponse, error) {
	tags, _ := dbconn.GetTags()

	res := &pb.GetTagsResponse{}

	for _, dbtag := range tags {
		t := &pb.Tag{
			Id:    dbtag.ID,
			Title: dbtag.Title,
		}

		res.Tags = append(res.Tags, t)
	}

	return res, nil
}

func (api *wackyTrackyClientService) Tag(ctx context.Context, req *pb.TagRequest) (*pb.TagResponse, error) {
	return nil, nil
}

func (api *wackyTrackyClientService) UpdateList(ctx context.Context, req *pb.UpdateListRequest) (*pb.UpdateListResponse, error) {
	return nil, nil
}

func (api *wackyTrackyClientService) Init(ctx context.Context, req *pb.InitRequest) (*pb.InitResponse, error) {
	res := &pb.InitResponse{
		Wallpaper: "wallpaper.jpg",
	}

	return res, nil
}

func newServer() *wackyTrackyClientService {
	server := wackyTrackyClientService{}
	return &server
}
