package clientapi

import (
	"context"
	"net/http"
	"os/exec"
	"time"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
	clientv1 "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1/clientv1connect"
	"github.com/wacky-tracky/wacky-tracky-server/internal/buildinfo"
	dbmdl "github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/todotxt"
	"github.com/wacky-tracky/wacky-tracky-server/internal/ruleeval"
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
	if err := api.dbconn.Connect(); err != nil {
		log.Warnf("DB connect: %v", err)
	}
	api.dbconn.Print()
	log.Infof("Client API initialized")
	return api
}

func (api *wackyTrackyClientService) GetNewHandler() (string, http.Handler) {
	path, handler := clientv1.NewWackyTrackyClientServiceHandler(api)

	return path, handler
}

func dbTaskToPb(t *dbmdl.DBTask) *pb.Task {
	if t == nil {
		return nil
	}
	out := &pb.Task{
		Id:            t.ID,
		Content:       t.Content,
		ParentId:      t.ParentId,
		ParentType:    t.ParentType,
		CountSubitems: t.CountSubitems,
		Tags:          append([]string{}, t.Tags...),
		Contexts:      append([]string{}, t.Contexts...),
		WaitUntil:     t.WaitUntil,
		Priority:      t.Priority,
		DueDate:       t.DueDate,
	}
	return out
}

func (api *wackyTrackyClientService) getChildTasks(parentId string, parentType string) ([]dbmdl.DBTask, error) {
	if parentType == "task" {
		return api.dbconn.GetSubtasks(parentId)
	}
	return api.dbconn.GetTasks(parentId)
}

func (api *wackyTrackyClientService) buildTaskTree(parentId string, parentType string) ([]*pb.Task, map[string]*pb.TaskIdList, error) {
	items, err := api.getChildTasks(parentId, parentType)
	if err != nil {
		return nil, nil, err
	}
	tasks := make([]*pb.Task, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, t := range items {
		tasks = append(tasks, dbTaskToPb(&t))
		ids = append(ids, t.ID)
	}
	tree := map[string]*pb.TaskIdList{
		parentId: {Ids: ids},
	}
	for _, t := range items {
		var err error
		tasks, _, err = api.appendSubtrees(tasks, tree, t.ID)
		if err != nil {
			return nil, nil, err
		}
	}
	return tasks, tree, nil
}

func (api *wackyTrackyClientService) appendSubtrees(tasks []*pb.Task, tree map[string]*pb.TaskIdList, taskId string) ([]*pb.Task, map[string]*pb.TaskIdList, error) {
	subTasks, subTree, err := api.buildTaskTree(taskId, "task")
	if err != nil {
		return nil, nil, err
	}
	tasks = append(tasks, subTasks...)
	for k, v := range subTree {
		tree[k] = v
	}
	return tasks, subTree, nil
}

const hideAtTimesKey = "hide-at-times"

func collectHiddenNamesFromProps(props map[string]map[string]string, env ruleeval.RuleEnv) []string {
	var out []string
	for name, m := range props {
		if e := m[hideAtTimesKey]; e != "" {
			ok, _ := ruleeval.Eval(e, env)
			if ok {
				out = append(out, name)
			}
		}
	}
	return out
}

func hideAtTimesExprForProp(props map[string]map[string]string, name string) string {
	m := props[name]
	if m == nil {
		return ""
	}
	return m[hideAtTimesKey]
}

func collectHideAtTimesExprs(task *pb.Task, tagProps, ctxProps map[string]map[string]string) []string {
	var exprs []string
	for _, tag := range task.Tags {
		if e := hideAtTimesExprForProp(tagProps, tag); e != "" {
			exprs = append(exprs, e)
		}
	}
	for _, ctx := range task.Contexts {
		if e := hideAtTimesExprForProp(ctxProps, ctx); e != "" {
			exprs = append(exprs, e)
		}
	}
	return exprs
}

func taskIsHiddenByExprs(exprs []string, env ruleeval.RuleEnv) bool {
	for _, e := range exprs {
		ok, err := ruleeval.Eval(e, env)
		if err != nil {
			continue
		}
		if ok {
			return true
		}
	}
	return false
}

func ensureTppMaps(tagProps, ctxProps map[string]map[string]string) (map[string]map[string]string, map[string]map[string]string) {
	if tagProps == nil {
		tagProps = make(map[string]map[string]string)
	}
	if ctxProps == nil {
		ctxProps = make(map[string]map[string]string)
	}
	return tagProps, ctxProps
}

func getTaskPropertyMaps(store dbmdl.TaskPropertyPropertiesStore) (tagProps, ctxProps map[string]map[string]string, ok bool) {
	tagProps, ctxProps, err := store.GetTaskPropertyProperties()
	if err != nil || (tagProps == nil && ctxProps == nil) {
		return nil, nil, false
	}
	tagProps, ctxProps = ensureTppMaps(tagProps, ctxProps)
	return tagProps, ctxProps, true
}

func buildHiddenSet(tasks []*pb.Task, tagProps, ctxProps map[string]map[string]string, env ruleeval.RuleEnv) map[string]bool {
	hiddenSet := make(map[string]bool)
	for _, t := range tasks {
		exprs := collectHideAtTimesExprs(t, tagProps, ctxProps)
		if taskIsHiddenByExprs(exprs, env) {
			hiddenSet[t.Id] = true
		}
	}
	return hiddenSet
}

func (api *wackyTrackyClientService) listTasksFilterHideAtTimes(tasks []*pb.Task, tree map[string]*pb.TaskIdList) ([]*pb.Task, map[string]*pb.TaskIdList, []string, []string) {
	store, ok := api.dbconn.(dbmdl.TaskPropertyPropertiesStore)
	if !ok {
		return tasks, tree, nil, nil
	}
	tagProps, ctxProps, ok := getTaskPropertyMaps(store)
	if !ok {
		return tasks, tree, nil, nil
	}
	env := ruleeval.NewRuleEnv(time.Now())
	hiddenTagNames := collectHiddenNamesFromProps(tagProps, env)
	hiddenContextNames := collectHiddenNamesFromProps(ctxProps, env)
	hiddenSet := buildHiddenSet(tasks, tagProps, ctxProps, env)
	for id := range hiddenSet {
		api.addDescendantsToHidden(id, tree, hiddenSet)
	}
	filteredTasks := filterTasksByHidden(tasks, hiddenSet)
	filteredTree := filterTreeByHidden(tree, hiddenSet)
	return filteredTasks, filteredTree, hiddenTagNames, hiddenContextNames
}

func filterTasksByHidden(tasks []*pb.Task, hiddenSet map[string]bool) []*pb.Task {
	var out []*pb.Task
	for _, t := range tasks {
		if !hiddenSet[t.Id] {
			out = append(out, t)
		}
	}
	return out
}

func filterTreeByHidden(tree map[string]*pb.TaskIdList, hiddenSet map[string]bool) map[string]*pb.TaskIdList {
	out := make(map[string]*pb.TaskIdList)
	for parentID, idList := range tree {
		var ids []string
		for _, id := range idList.Ids {
			if !hiddenSet[id] {
				ids = append(ids, id)
			}
		}
		out[parentID] = &pb.TaskIdList{Ids: ids}
	}
	return out
}

func (api *wackyTrackyClientService) addDescendantsToHidden(id string, tree map[string]*pb.TaskIdList, hiddenSet map[string]bool) {
	idList := tree[id]
	if idList == nil {
		return
	}
	for _, childID := range idList.Ids {
		hiddenSet[childID] = true
		api.addDescendantsToHidden(childID, tree, hiddenSet)
	}
}

func (api *wackyTrackyClientService) ListTasks(ctx context.Context, req *connect.Request[pb.ListTasksRequest]) (*connect.Response[pb.ListTasksResponse], error) {
	metricListTasksCount.Inc()

	parentId := req.Msg.ParentId
	parentType := req.Msg.ParentType
	if parentType == "" {
		parentType = "list"
	}
	if parentId == "" {
		return connect.NewResponse(&pb.ListTasksResponse{}), nil
	}

	tasks, tree, err := api.buildTaskTree(parentId, parentType)
	if err != nil {
		return connect.NewResponse(&pb.ListTasksResponse{}), err
	}
	tasks, tree, hiddenTags, hiddenContexts := api.listTasksFilterHideAtTimes(tasks, tree)
	return connect.NewResponse(&pb.ListTasksResponse{
		Tasks:              tasks,
		Tree:               tree,
		HiddenTagNames:     hiddenTags,
		HiddenContextNames: hiddenContexts,
	}), nil
}

func (api *wackyTrackyClientService) DoneTask(ctx context.Context, req *connect.Request[pb.DoneTaskRequest]) (*connect.Response[pb.DoneTaskResponse], error) {
	if upd, ok := api.dbconn.(dbmdl.Updatable); ok && req.Msg.Id != "" {
		_ = upd.DoneTask(req.Msg.Id)
	}
	return connect.NewResponse(&pb.DoneTaskResponse{}), nil
}

func (api *wackyTrackyClientService) CreateTask(ctx context.Context, req *connect.Request[pb.CreateTaskRequest]) (*connect.Response[pb.CreateTaskResponse], error) {
	id, err := api.dbconn.CreateTask(req.Msg.Content, req.Msg.ParentListId, req.Msg.ParentTaskId)
	if err != nil {
		return nil, err
	}
	task, err := api.dbconn.GetTask(id)
	if err != nil || task == nil {
		return connect.NewResponse(&pb.CreateTaskResponse{
			Task: &pb.Task{Id: id, Content: req.Msg.Content},
		}), nil
	}
	return connect.NewResponse(&pb.CreateTaskResponse{Task: dbTaskToPb(task)}), nil
}

func (api *wackyTrackyClientService) CreateList(ctx context.Context, req *connect.Request[pb.CreateListRequest]) (*connect.Response[pb.CreateListResponse], error) {
	if err := api.dbconn.CreateList(req.Msg.Title); err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.CreateListResponse{}), nil
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
	_ = req
	return connect.NewResponse(&pb.TagResponse{}), nil
}

func (api *wackyTrackyClientService) UpdateList(ctx context.Context, req *connect.Request[pb.UpdateListRequest]) (*connect.Response[pb.UpdateListResponse], error) {
	if req.Msg.Id != "" {
		_ = api.dbconn.UpdateList(req.Msg.Id, req.Msg.Title)
	}
	return connect.NewResponse(&pb.UpdateListResponse{}), nil
}

func (api *wackyTrackyClientService) DeleteList(ctx context.Context, req *connect.Request[pb.DeleteListRequest]) (*connect.Response[pb.DeleteListResponse], error) {
	if req.Msg.Id != "" {
		_ = api.dbconn.DeleteList(req.Msg.Id)
	}
	return connect.NewResponse(&pb.DeleteListResponse{}), nil
}

func (api *wackyTrackyClientService) SearchTasks(ctx context.Context, req *connect.Request[pb.SearchTasksRequest]) (*connect.Response[pb.SearchTasksResponse], error) {
	var tasks []*pb.Task
	if s, ok := api.dbconn.(dbmdl.Searchable); ok && req.Msg.Query != "" {
		items, err := s.SearchTasks(req.Msg.Query)
		if err == nil {
			for i := range items {
				tasks = append(tasks, dbTaskToPb(&items[i]))
			}
		}
	}
	return connect.NewResponse(&pb.SearchTasksResponse{Tasks: tasks}), nil
}

func (api *wackyTrackyClientService) UpdateTask(ctx context.Context, req *connect.Request[pb.UpdateTaskRequest]) (*connect.Response[pb.UpdateTaskResponse], error) {
	var task *pb.Task
	if upd, ok := api.dbconn.(dbmdl.Updatable); ok && req.Msg.Id != "" {
		_ = upd.UpdateTask(req.Msg.Id, req.Msg.Content)
		if t, _ := api.dbconn.GetTask(req.Msg.Id); t != nil {
			task = dbTaskToPb(t)
		}
	}
	return connect.NewResponse(&pb.UpdateTaskResponse{Task: task}), nil
}

func (api *wackyTrackyClientService) RepoStatus(ctx context.Context, req *connect.Request[pb.RepoStatusRequest]) (*connect.Response[pb.RepoStatusResponse], error) {
	_ = req
	out := ""
	if t, ok := api.dbconn.(*todotxt.TodoTxt); ok {
		dir := t.Dir()
		if dir != "" {
			cmd := exec.CommandContext(ctx, "git", "-C", dir, "status")
			b, err := cmd.CombinedOutput()
			if err != nil {
				out = string(b) + "\n" + err.Error()
			} else {
				out = string(b)
			}
		}
	}
	return connect.NewResponse(&pb.RepoStatusResponse{Output: out}), nil
}

func (api *wackyTrackyClientService) GetSavedSearches(ctx context.Context, req *connect.Request[pb.GetSavedSearchesRequest]) (*connect.Response[pb.GetSavedSearchesResponse], error) {
	var list []*pb.SavedSearch
	if s, ok := api.dbconn.(dbmdl.SavedSearchesStore); ok {
		searches, err := s.GetSavedSearches()
		if err == nil {
			for i := range searches {
				list = append(list, &pb.SavedSearch{Id: searches[i].ID, Name: searches[i].Title, Query: searches[i].Query})
			}
		}
	}
	return connect.NewResponse(&pb.GetSavedSearchesResponse{SavedSearches: list}), nil
}

func (api *wackyTrackyClientService) SetSavedSearches(ctx context.Context, req *connect.Request[pb.SetSavedSearchesRequest]) (*connect.Response[pb.SetSavedSearchesResponse], error) {
	if s, ok := api.dbconn.(dbmdl.SavedSearchesStore); ok && req.Msg.SavedSearches != nil {
		searches := make([]dbmdl.SavedSearch, len(req.Msg.SavedSearches))
		for i, p := range req.Msg.SavedSearches {
			searches[i] = dbmdl.SavedSearch{ID: p.Id, Title: p.Name, Query: p.Query}
		}
		_ = s.SetSavedSearches(searches)
	}
	return connect.NewResponse(&pb.SetSavedSearchesResponse{}), nil
}

func (api *wackyTrackyClientService) GetTaskMetadata(ctx context.Context, req *connect.Request[pb.GetTaskMetadataRequest]) (*connect.Response[pb.GetTaskMetadataResponse], error) {
	var fields map[string]string
	if s, ok := api.dbconn.(dbmdl.TaskMetadataStore); ok && req.Msg.TaskId != "" {
		fields, _ = s.GetTaskMetadata(req.Msg.TaskId)
	}
	return connect.NewResponse(&pb.GetTaskMetadataResponse{Fields: fields}), nil
}

func (api *wackyTrackyClientService) SetTaskMetadata(ctx context.Context, req *connect.Request[pb.SetTaskMetadataRequest]) (*connect.Response[pb.SetTaskMetadataResponse], error) {
	if s, ok := api.dbconn.(dbmdl.TaskMetadataStore); ok && req.Msg.TaskId != "" && req.Msg.Field != "" {
		_ = s.SetTaskMetadata(req.Msg.TaskId, req.Msg.Field, req.Msg.Value)
	}
	return connect.NewResponse(&pb.SetTaskMetadataResponse{}), nil
}

func mapToTaskPropertyProps(m map[string]string) *pb.TaskPropertyProps {
	if m == nil {
		return nil
	}
	props := make(map[string]string, len(m))
	for k, v := range m {
		props[k] = v
	}
	return &pb.TaskPropertyProps{Props: props}
}

func (api *wackyTrackyClientService) GetTaskPropertyProperties(ctx context.Context, req *connect.Request[pb.GetTaskPropertyPropertiesRequest]) (*connect.Response[pb.GetTaskPropertyPropertiesResponse], error) {
	_ = req
	res := &pb.GetTaskPropertyPropertiesResponse{
		TagProperties:     make(map[string]*pb.TaskPropertyProps),
		ContextProperties: make(map[string]*pb.TaskPropertyProps),
	}
	if s, ok := api.dbconn.(dbmdl.TaskPropertyPropertiesStore); ok {
		tagProps, ctxProps, err := s.GetTaskPropertyProperties()
		if err != nil {
			return nil, err
		}
		for name, props := range tagProps {
			res.TagProperties[name] = mapToTaskPropertyProps(props)
		}
		for name, props := range ctxProps {
			res.ContextProperties[name] = mapToTaskPropertyProps(props)
		}
	}
	return connect.NewResponse(res), nil
}

func (api *wackyTrackyClientService) SetTaskPropertyProperty(ctx context.Context, req *connect.Request[pb.SetTaskPropertyPropertyRequest]) (*connect.Response[pb.SetTaskPropertyPropertyResponse], error) {
	if s, ok := api.dbconn.(dbmdl.TaskPropertyPropertiesStore); ok && req.Msg.PropertyType != "" && req.Msg.PropertyName != "" && req.Msg.Key != "" {
		_ = s.SetTaskPropertyProperty(req.Msg.PropertyType, req.Msg.PropertyName, req.Msg.Key, req.Msg.Value)
	}
	return connect.NewResponse(&pb.SetTaskPropertyPropertyResponse{}), nil
}

func (api *wackyTrackyClientService) RuleStatus(ctx context.Context, req *connect.Request[pb.RuleStatusRequest]) (*connect.Response[pb.RuleStatusResponse], error) {
	_ = req
	env := ruleeval.NewRuleEnv(time.Now())
	return connect.NewResponse(&pb.RuleStatusResponse{
		D: env.D,
		H: int32(env.H),
		M: int32(env.M),
	}), nil
}

func (api *wackyTrackyClientService) RuleTest(ctx context.Context, req *connect.Request[pb.RuleTestRequest]) (*connect.Response[pb.RuleTestResponse], error) {
	exprStr := req.Msg.Expression
	if exprStr == "" {
		return connect.NewResponse(&pb.RuleTestResponse{Compiles: false, CompileError: "expression is empty"}), nil
	}
	program, err := ruleeval.Compile(exprStr)
	if err != nil {
		return connect.NewResponse(&pb.RuleTestResponse{
			Compiles:     false,
			CompileError: err.Error(),
		}), nil
	}
	env := ruleeval.NewRuleEnv(time.Now())
	result, err := ruleeval.Run(program, env)
	if err != nil {
		return connect.NewResponse(&pb.RuleTestResponse{
			Compiles:  true,
			Result:    false,
			EvalError: err.Error(),
		}), nil
	}
	return connect.NewResponse(&pb.RuleTestResponse{
		Compiles: true,
		Result:   result,
	}), nil
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
