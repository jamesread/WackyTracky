package clientapi

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
	clientv1 "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1/clientv1connect"
	"github.com/wacky-tracky/wacky-tracky-server/internal/buildinfo"
	dbmdl "github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
	"net/http"

	"connectrpc.com/connect"
	log "github.com/sirupsen/logrus"

	"context"
	//	. "github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

var metricListTasksCount = promauto.NewCounter(prometheus.CounterOpts{
	Name: "list_tasks_count",
	Help: "The total number of ListTasks API calls",
})

type wackyTrackyClientService struct {
	dbconn dbmdl.DB
}

func GetNewClientAPI(newdb dbmdl.DB) *wackyTrackyClientService {
	api := newServer()
	api.dbconn = newdb
	api.dbconn.Connect()
	api.dbconn.Print()
	api.dbconn.GetTasks("418")

	log.Infof("Client API initialized %+v", api.dbconn)

	return api
}

func (api *wackyTrackyClientService) GetNewHandler() (string, http.Handler) {
	path, handler := clientv1.NewWackyTrackyClientServiceHandler(api)

	return path, handler
}

func (api *wackyTrackyClientService) ListTasks(ctx context.Context, req *connect.Request[pb.ListTasksRequest]) (*connect.Response[pb.ListTasksResponse], error) {
	metricListTasksCount.Inc()

	var items []dbmdl.DBTask
	var err error

	if req.Msg.ParentType == "task" {
		parentId := req.Msg.ParentId

		log.Infof("ListTasks item: %+v", api.dbconn)

		items, err = api.dbconn.GetSubtasks(parentId)
	} else if req.Msg.ParentType == "list" {
		log.Infof("ListTasks list: %+v", api.dbconn)

		items, err = api.dbconn.GetTasks(req.Msg.ParentId)
	} else {
		log.Infof("Unknown parent type: %s", req.Msg.ParentType)
	}

	ret := &pb.ListTasksResponse{}

	if err != nil {
		return connect.NewResponse(ret), err
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

	return connect.NewResponse(ret), nil
}

func (api *wackyTrackyClientService) DeleteTask(ctx context.Context, req *connect.Request[pb.DeleteTaskRequest]) (*connect.Response[pb.DeleteTaskResponse], error) {
	ret := connect.NewResponse(&pb.DeleteTaskResponse{})

	return ret, nil
}

func (api *wackyTrackyClientService) CreateTask(ctx context.Context, req *connect.Request[pb.CreateTaskRequest]) (*connect.Response[pb.CreateTaskResponse], error) {
	api.dbconn.CreateTask(req.Msg.Content)

	res := connect.NewResponse(&pb.CreateTaskResponse{})

	return res, nil
}

func (api *wackyTrackyClientService) CreateList(ctx context.Context, req *connect.Request[pb.CreateListRequest]) (*connect.Response[pb.CreateListResponse], error) {
	api.dbconn.CreateList(req.Msg.Title)

	res := connect.NewResponse(&pb.CreateListResponse{})

	return res, nil
}

func (api *wackyTrackyClientService) GetLists(ctx context.Context, req *connect.Request[pb.GetListsRequest]) (*connect.Response[pb.GetListsResponse], error) {
	lists, _ := api.dbconn.GetLists()

	res := &pb.GetListsResponse{}

	for _, dblist := range lists {
		l := &pb.List{
			Title:      dblist.Title,
			Id:         dblist.ID,
			CountItems: dblist.CountTasks,
		}

		res.Lists = append(res.Lists, l)
	}

	return connect.NewResponse(res), nil
}

func (api *wackyTrackyClientService) Version(ctx context.Context, req *connect.Request[pb.VersionRequest]) (*connect.Response[pb.VersionResponse], error) {
	return connect.NewResponse(&pb.VersionResponse{
		Version: buildinfo.Version,
		Commit:  buildinfo.Commit,
		Date:    buildinfo.Date,
	}), nil
}

func (api *wackyTrackyClientService) GetTags(ctx context.Context, req *connect.Request[pb.GetTagsRequest]) (*connect.Response[pb.GetTagsResponse], error) {
	tags, _ := api.dbconn.GetTags()

	res := &pb.GetTagsResponse{}

	for _, dbtag := range tags {
		t := &pb.Tag{
			Id:    dbtag.ID,
			Title: dbtag.Title,
		}

		res.Tags = append(res.Tags, t)
	}

	return connect.NewResponse(res), nil
}

func (api *wackyTrackyClientService) Tag(ctx context.Context, req *connect.Request[pb.TagRequest]) (*connect.Response[pb.TagResponse], error) {
	return nil, nil
}

func (api *wackyTrackyClientService) UpdateList(ctx context.Context, req *connect.Request[pb.UpdateListRequest]) (*connect.Response[pb.UpdateListResponse], error) {
	return nil, nil
}

func (api *wackyTrackyClientService) Init(ctx context.Context, req *connect.Request[pb.InitRequest]) (*connect.Response[pb.InitResponse], error) {
	res := &pb.InitResponse{
		Wallpaper: "wallpaper.jpg",
	}

	return connect.NewResponse(res), nil
}

func newServer() *wackyTrackyClientService {
	server := wackyTrackyClientService{}
	return &server
}
