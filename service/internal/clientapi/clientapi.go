package clientapi

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"os/exec"
	"time"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
	clientv1 "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1/clientv1connect"
	"github.com/wacky-tracky/wacky-tracky-server/internal/auth"
	"github.com/wacky-tracky/wacky-tracky-server/internal/buildinfo"
	dbmdl "github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/todotxt"
	"github.com/wacky-tracky/wacky-tracky-server/internal/gitsync"
	"github.com/wacky-tracky/wacky-tracky-server/internal/iamstore"
	"github.com/wacky-tracky/wacky-tracky-server/internal/ruleeval"
	"github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

var metricListTasksCount = promauto.NewCounter(prometheus.CounterOpts{
	Name: "list_tasks_count",
	Help: "The total number of ListTasks API calls",
})

type wackyTrackyClientService struct {
	dbconn   dbmdl.DB
	iamStore iamstore.Store
	log      *log.Logger
	// ssoOptOut lets Logout opt a browser out of proxy JWT sign-in (eg Cloudflare Access).
	ssoOptOut *auth.SSOOptOut
}

func GetNewClientAPI(newdb dbmdl.DB, iamStore iamstore.Store) *wackyTrackyClientService {
	api := newServer()
	api.dbconn = newdb
	api.iamStore = iamStore
	if api.log == nil {
		api.log = log.StandardLogger()
	}
	if err := api.dbconn.Connect(); err != nil {
		log.Warnf("DB connect: %v", err)
	}
	api.dbconn.Print()
	log.Infof("Client API initialized")
	return api
}

func (api *wackyTrackyClientService) GetNewHandler(authLayer *auth.Layer, ssoOptOut *auth.SSOOptOut) (string, http.Handler) {
	path, handler := clientv1.NewWackyTrackyClientServiceHandler(api)
	if authLayer != nil {
		handler = authLayer.WrapHandler(handler)
	}
	// Outside the auth layer, so an opted-out browser's JWT is dropped before it is checked.
	api.ssoOptOut = ssoOptOut
	handler = ssoOptOut.Wrap(handler)
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
	maps.Copy(tree, subTree)
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
	upd, err := api.updatableFor(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	if err := upd.DoneTask(req.Msg.Id); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.DoneTaskResponse{}), nil
}

func (api *wackyTrackyClientService) MoveTask(ctx context.Context, req *connect.Request[pb.MoveTaskRequest]) (*connect.Response[pb.MoveTaskResponse], error) {
	mover, err := api.taskMoverFor(req.Msg)
	if err != nil {
		return nil, err
	}
	if err := mover.MoveTask(req.Msg.TaskId, req.Msg.TargetListId); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.MoveTaskResponse{Task: api.getTaskAsPb(req.Msg.TaskId)}), nil
}

func (api *wackyTrackyClientService) taskMoverFor(req *pb.MoveTaskRequest) (dbmdl.TaskMover, error) {
	mover, ok := api.dbconn.(dbmdl.TaskMover)
	if !ok {
		return nil, unimplemented("moving tasks")
	}
	if req.TaskId == "" || req.TargetListId == "" {
		return nil, invalidArgument("task_id and target_list_id are required")
	}
	return mover, nil
}

func (api *wackyTrackyClientService) updatableFor(taskId string) (dbmdl.Updatable, error) {
	upd, ok := api.dbconn.(dbmdl.Updatable)
	if !ok {
		return nil, unimplemented("updating tasks")
	}
	if taskId == "" {
		return nil, invalidArgument("id is required")
	}
	return upd, nil
}

var errorCodes = map[error]connect.Code{
	dbmdl.ErrTaskNotFound:            connect.CodeNotFound,
	dbmdl.ErrListNotFound:            connect.CodeNotFound,
	dbmdl.ErrDefaultListNotDeletable: connect.CodeFailedPrecondition,
	dbmdl.ErrInvalidPropertyType:     connect.CodeInvalidArgument,
	dbmdl.ErrInvalidMetadataField:    connect.CodeInvalidArgument,
}

// toConnectError maps backend sentinel errors to Connect codes; anything else
// is returned as-is and surfaces to clients as CodeUnknown.
func toConnectError(err error) error {
	for sentinel, code := range errorCodes {
		if errors.Is(err, sentinel) {
			return connect.NewError(code, err)
		}
	}
	return err
}

func invalidArgument(msg string) error {
	return connect.NewError(connect.CodeInvalidArgument, errors.New(msg))
}

func unimplemented(feature string) error {
	return connect.NewError(connect.CodeUnimplemented, errors.New("the database backend does not support "+feature))
}

func (api *wackyTrackyClientService) getTaskAsPb(id string) *pb.Task {
	return api.getTaskOrFallback(id, &pb.Task{Id: id})
}

func (api *wackyTrackyClientService) getTaskOrFallback(id string, fallback *pb.Task) *pb.Task {
	task, err := api.dbconn.GetTask(id)
	if err != nil || task == nil {
		return fallback
	}
	return dbTaskToPb(task)
}

func (api *wackyTrackyClientService) CreateTask(ctx context.Context, req *connect.Request[pb.CreateTaskRequest]) (*connect.Response[pb.CreateTaskResponse], error) {
	id, err := api.dbconn.CreateTask(req.Msg.Content, req.Msg.ParentListId, req.Msg.ParentTaskId)
	if err != nil {
		return nil, err
	}
	// The task is already saved; if reading it back fails, echo what was sent rather than fail the create.
	task := api.getTaskOrFallback(id, &pb.Task{Id: id, Content: req.Msg.Content})
	return connect.NewResponse(&pb.CreateTaskResponse{Task: task}), nil
}

func (api *wackyTrackyClientService) CreateList(ctx context.Context, req *connect.Request[pb.CreateListRequest]) (*connect.Response[pb.CreateListResponse], error) {
	if err := api.dbconn.CreateList(req.Msg.Title); err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.CreateListResponse{}), nil
}

func (api *wackyTrackyClientService) GetLists(ctx context.Context, req *connect.Request[pb.GetListsRequest]) (*connect.Response[pb.GetListsResponse], error) {
	lists, err := api.dbconn.GetLists()
	if err != nil {
		return nil, err
	}

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
	tags, err := api.dbconn.GetTags()
	if err != nil {
		return nil, err
	}

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
	if req.Msg.Id == "" {
		return nil, invalidArgument("id is required")
	}
	if err := api.dbconn.UpdateList(req.Msg.Id, req.Msg.Title); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.UpdateListResponse{}), nil
}

func (api *wackyTrackyClientService) DeleteList(ctx context.Context, req *connect.Request[pb.DeleteListRequest]) (*connect.Response[pb.DeleteListResponse], error) {
	if req.Msg.Id == "" {
		return nil, invalidArgument("id is required")
	}
	if err := api.dbconn.DeleteList(req.Msg.Id); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.DeleteListResponse{}), nil
}

func (api *wackyTrackyClientService) SearchTasks(ctx context.Context, req *connect.Request[pb.SearchTasksRequest]) (*connect.Response[pb.SearchTasksResponse], error) {
	s, ok := api.dbconn.(dbmdl.Searchable)
	if !ok {
		return nil, unimplemented("searching tasks")
	}
	if req.Msg.Query == "" {
		return connect.NewResponse(&pb.SearchTasksResponse{}), nil
	}
	items, err := s.SearchTasks(req.Msg.Query)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.SearchTasksResponse{Tasks: dbTasksToPb(items)}), nil
}

func dbTasksToPb(items []dbmdl.DBTask) []*pb.Task {
	tasks := make([]*pb.Task, 0, len(items))
	for i := range items {
		tasks = append(tasks, dbTaskToPb(&items[i]))
	}
	return tasks
}

func (api *wackyTrackyClientService) UpdateTask(ctx context.Context, req *connect.Request[pb.UpdateTaskRequest]) (*connect.Response[pb.UpdateTaskResponse], error) {
	upd, err := api.updatableFor(req.Msg.Id)
	if err != nil {
		return nil, err
	}
	if err := upd.UpdateTask(req.Msg.Id, req.Msg.Content); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.UpdateTaskResponse{Task: api.getTaskAsPb(req.Msg.Id)}), nil
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

func (api *wackyTrackyClientService) RepoSync(ctx context.Context, req *connect.Request[pb.RepoSyncRequest]) (*connect.Response[pb.RepoSyncResponse], error) {
	_ = req
	res := gitsync.Result{Success: false, Message: "Git sync is only available with the todo.txt backend"}
	if t, ok := api.dbconn.(*todotxt.TodoTxt); ok {
		res = gitsync.Sync(ctx, t.Dir(), req.Msg.GetServerName())
	}
	return connect.NewResponse(&pb.RepoSyncResponse{
		Success: res.Success,
		Message: res.Message,
		Steps:   res.Steps,
	}), nil
}

func (api *wackyTrackyClientService) GetSavedSearches(ctx context.Context, req *connect.Request[pb.GetSavedSearchesRequest]) (*connect.Response[pb.GetSavedSearchesResponse], error) {
	s, ok := api.dbconn.(dbmdl.SavedSearchesStore)
	if !ok {
		return connect.NewResponse(&pb.GetSavedSearchesResponse{}), nil
	}
	searches, err := s.GetSavedSearches()
	if err != nil {
		return nil, err
	}
	var list []*pb.SavedSearch
	for i := range searches {
		list = append(list, &pb.SavedSearch{Id: searches[i].ID, Name: searches[i].Title, Query: searches[i].Query})
	}
	return connect.NewResponse(&pb.GetSavedSearchesResponse{SavedSearches: list}), nil
}

func (api *wackyTrackyClientService) SetSavedSearches(ctx context.Context, req *connect.Request[pb.SetSavedSearchesRequest]) (*connect.Response[pb.SetSavedSearchesResponse], error) {
	s, ok := api.dbconn.(dbmdl.SavedSearchesStore)
	if !ok {
		return nil, unimplemented("saved searches")
	}
	searches := make([]dbmdl.SavedSearch, len(req.Msg.SavedSearches))
	for i, p := range req.Msg.SavedSearches {
		searches[i] = dbmdl.SavedSearch{ID: p.Id, Title: p.Name, Query: p.Query}
	}
	if err := s.SetSavedSearches(searches); err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.SetSavedSearchesResponse{}), nil
}

func (api *wackyTrackyClientService) GetTaskMetadata(ctx context.Context, req *connect.Request[pb.GetTaskMetadataRequest]) (*connect.Response[pb.GetTaskMetadataResponse], error) {
	s, ok := api.dbconn.(dbmdl.TaskMetadataStore)
	if !ok {
		return connect.NewResponse(&pb.GetTaskMetadataResponse{}), nil
	}
	if req.Msg.TaskId == "" {
		return nil, invalidArgument("task_id is required")
	}
	fields, err := s.GetTaskMetadata(req.Msg.TaskId)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.GetTaskMetadataResponse{Fields: fields}), nil
}

func (api *wackyTrackyClientService) SetTaskMetadata(ctx context.Context, req *connect.Request[pb.SetTaskMetadataRequest]) (*connect.Response[pb.SetTaskMetadataResponse], error) {
	s, err := api.taskMetadataStoreFor(req.Msg)
	if err != nil {
		return nil, err
	}
	if err := s.SetTaskMetadata(req.Msg.TaskId, req.Msg.Field, req.Msg.Value); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.SetTaskMetadataResponse{}), nil
}

func (api *wackyTrackyClientService) taskMetadataStoreFor(req *pb.SetTaskMetadataRequest) (dbmdl.TaskMetadataStore, error) {
	s, ok := api.dbconn.(dbmdl.TaskMetadataStore)
	if !ok {
		return nil, unimplemented("task metadata")
	}
	if req.TaskId == "" || req.Field == "" {
		return nil, invalidArgument("task_id and field are required")
	}
	return s, nil
}

func mapToTaskPropertyProps(m map[string]string) *pb.TaskPropertyProps {
	if m == nil {
		return nil
	}
	props := make(map[string]string, len(m))
	maps.Copy(props, m)
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
	s, err := api.taskPropertyStoreFor(req.Msg)
	if err != nil {
		return nil, err
	}
	if err := s.SetTaskPropertyProperty(req.Msg.PropertyType, req.Msg.PropertyName, req.Msg.Key, req.Msg.Value); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.SetTaskPropertyPropertyResponse{}), nil
}

func (api *wackyTrackyClientService) taskPropertyStoreFor(req *pb.SetTaskPropertyPropertyRequest) (dbmdl.TaskPropertyPropertiesStore, error) {
	s, ok := api.dbconn.(dbmdl.TaskPropertyPropertiesStore)
	if !ok {
		return nil, unimplemented("tag and context properties")
	}
	if req.PropertyName == "" || req.Key == "" {
		return nil, invalidArgument("property_name and key are required")
	}
	return s, nil
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
	return connect.NewResponse(testRuleExpression(req.Msg.Expression, time.Now())), nil
}

// testRuleExpression reports compile and evaluation problems in the response,
// because a broken expression is the answer the caller asked for, not an RPC failure.
func testRuleExpression(exprStr string, now time.Time) *pb.RuleTestResponse {
	if exprStr == "" {
		return &pb.RuleTestResponse{Compiles: false, CompileError: "expression is empty"}
	}
	program, err := ruleeval.Compile(exprStr)
	if err != nil {
		return &pb.RuleTestResponse{Compiles: false, CompileError: err.Error()}
	}
	result, err := ruleeval.Run(program, ruleeval.NewRuleEnv(now))
	if err != nil {
		return &pb.RuleTestResponse{Compiles: true, Result: false, EvalError: err.Error()}
	}
	return &pb.RuleTestResponse{Compiles: true, Result: result}
}

func (api *wackyTrackyClientService) Init(ctx context.Context, req *connect.Request[pb.InitRequest]) (*connect.Response[pb.InitResponse], error) {
	cfg := runtimeconfig.RuntimeConfig
	showVersion := cfg.ShowVersionNumber
	currentVersion := ""
	availableVersion := ""
	showNew := false
	if showVersion {
		currentVersion = buildinfo.Version
		showNew = cfg.ShowNewVersions
		// No update-check yet; leave available_version empty until one exists.
	}
	res := &pb.InitResponse{
		Wallpaper:         "wallpaper.jpg",
		ShowFooter:        cfg.ShowFooter,
		ShowNewVersions:   showVersion && showNew,
		AvailableVersion:  availableVersion,
		CurrentVersion:    currentVersion,
		PageTitle:         "WackyTracky",
		ShowVersionNumber: showVersion,
	}
	return connect.NewResponse(res), nil
}

func newServer() *wackyTrackyClientService {
	return &wackyTrackyClientService{
		log: log.StandardLogger(),
	}
}
