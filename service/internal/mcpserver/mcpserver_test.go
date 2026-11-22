package mcpserver

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
)

// fakeService records the requests it receives and returns canned responses.
type fakeService struct {
	lastCreateTask *pb.CreateTaskRequest
	lastListTasks  *pb.ListTasksRequest
	lastDoneTask   *pb.DoneTaskRequest
	lastMoveTask   *pb.MoveTaskRequest
	lastSetMeta    *pb.SetTaskMetadataRequest
	setMetaCalls   []*pb.SetTaskMetadataRequest
	lastDeleteTask *pb.DeleteTaskRequest
	lastDeleteList *pb.DeleteListRequest
	lastCompleted  *pb.ListCompletedTasksRequest
	deleteErr      error
	createTaskErr  error
	moveTaskErr    error
	setMetaErr     error
}

func (f *fakeService) GetLists(_ context.Context, _ *connect.Request[pb.GetListsRequest]) (*connect.Response[pb.GetListsResponse], error) {
	return connect.NewResponse(&pb.GetListsResponse{
		Lists: []*pb.List{{Id: "list-1", Title: "Inbox", CountItems: 2}},
	}), nil
}

func (f *fakeService) ListTasks(_ context.Context, req *connect.Request[pb.ListTasksRequest]) (*connect.Response[pb.ListTasksResponse], error) {
	f.lastListTasks = req.Msg
	return connect.NewResponse(&pb.ListTasksResponse{
		Tasks: []*pb.Task{{Id: "t1", Content: "Buy milk"}},
	}), nil
}

func (f *fakeService) SearchTasks(_ context.Context, req *connect.Request[pb.SearchTasksRequest]) (*connect.Response[pb.SearchTasksResponse], error) {
	return connect.NewResponse(&pb.SearchTasksResponse{
		Tasks: []*pb.Task{{Id: "t2", Content: req.Msg.Query}},
	}), nil
}

func (f *fakeService) CreateTask(_ context.Context, req *connect.Request[pb.CreateTaskRequest]) (*connect.Response[pb.CreateTaskResponse], error) {
	f.lastCreateTask = req.Msg
	if f.createTaskErr != nil {
		return nil, f.createTaskErr
	}
	return connect.NewResponse(&pb.CreateTaskResponse{Task: &pb.Task{Id: "new", Content: req.Msg.Content}}), nil
}

func (f *fakeService) UpdateTask(_ context.Context, req *connect.Request[pb.UpdateTaskRequest]) (*connect.Response[pb.UpdateTaskResponse], error) {
	return connect.NewResponse(&pb.UpdateTaskResponse{Task: &pb.Task{Id: req.Msg.Id, Content: req.Msg.Content}}), nil
}

func (f *fakeService) DoneTask(_ context.Context, req *connect.Request[pb.DoneTaskRequest]) (*connect.Response[pb.DoneTaskResponse], error) {
	f.lastDoneTask = req.Msg
	return connect.NewResponse(&pb.DoneTaskResponse{}), nil
}

func (f *fakeService) MoveTask(_ context.Context, req *connect.Request[pb.MoveTaskRequest]) (*connect.Response[pb.MoveTaskResponse], error) {
	f.lastMoveTask = req.Msg
	if f.moveTaskErr != nil {
		return nil, f.moveTaskErr
	}
	return connect.NewResponse(&pb.MoveTaskResponse{Task: &pb.Task{Id: req.Msg.TaskId, ParentId: req.Msg.TargetListId}}), nil
}

func (f *fakeService) CreateList(_ context.Context, _ *connect.Request[pb.CreateListRequest]) (*connect.Response[pb.CreateListResponse], error) {
	return connect.NewResponse(&pb.CreateListResponse{}), nil
}

func (f *fakeService) SetTaskMetadata(_ context.Context, req *connect.Request[pb.SetTaskMetadataRequest]) (*connect.Response[pb.SetTaskMetadataResponse], error) {
	f.lastSetMeta = req.Msg
	f.setMetaCalls = append(f.setMetaCalls, req.Msg)
	if f.setMetaErr != nil {
		return nil, f.setMetaErr
	}
	return connect.NewResponse(&pb.SetTaskMetadataResponse{}), nil
}

func (f *fakeService) DeleteTask(_ context.Context, req *connect.Request[pb.DeleteTaskRequest]) (*connect.Response[pb.DeleteTaskResponse], error) {
	f.lastDeleteTask = req.Msg
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}
	return connect.NewResponse(&pb.DeleteTaskResponse{}), nil
}

func (f *fakeService) DeleteList(_ context.Context, req *connect.Request[pb.DeleteListRequest]) (*connect.Response[pb.DeleteListResponse], error) {
	f.lastDeleteList = req.Msg
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}
	return connect.NewResponse(&pb.DeleteListResponse{}), nil
}

func (f *fakeService) ListCompletedTasks(_ context.Context, req *connect.Request[pb.ListCompletedTasksRequest]) (*connect.Response[pb.ListCompletedTasksResponse], error) {
	f.lastCompleted = req.Msg
	return connect.NewResponse(&pb.ListCompletedTasksResponse{
		Tasks: []*pb.Task{{Id: "d1", Content: "Done thing", Completed: true, CompletedAt: "2026-11-21"}},
	}), nil
}

func callWith(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: args}}
}

func resultText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	require.NotNil(t, res)
	require.NotEmpty(t, res.Content)
	tc, ok := mcp.AsTextContent(res.Content[0])
	require.True(t, ok, "expected text content")
	return tc.Text
}

func TestMCPRegistersWithoutPanic(t *testing.T) {
	srv := New(&fakeService{}).MCP()
	assert.NotNil(t, srv)
}

func TestHTTPHandler(t *testing.T) {
	h := New(&fakeService{}).HTTPHandler()
	assert.NotNil(t, h)
}

func TestHandleListLists(t *testing.T) {
	res, err := New(&fakeService{}).handleListLists(context.Background(), callWith(nil))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Contains(t, resultText(t, res), "Inbox")
}

func TestHandleListTasksRequiresListID(t *testing.T) {
	res, err := New(&fakeService{}).handleListTasks(context.Background(), callWith(nil))
	require.NoError(t, err)
	assert.True(t, res.IsError, "missing list_id should be an error result")
}

func TestHandleListTasksPassesListID(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleListTasks(context.Background(), callWith(map[string]any{"list_id": "list-1"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotNil(t, f.lastListTasks)
	assert.Equal(t, "list-1", f.lastListTasks.ParentId)
	assert.Equal(t, "list", f.lastListTasks.ParentType)
}

func TestHandleCreateTaskPassesArguments(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleCreateTask(context.Background(), callWith(map[string]any{
		"content": "Buy milk #errands",
		"list_id": "list-1",
	}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotNil(t, f.lastCreateTask)
	assert.Equal(t, "Buy milk #errands", f.lastCreateTask.Content)
	assert.Equal(t, "list-1", f.lastCreateTask.ParentListId)
}

func TestHandleCreateTaskRequiresContent(t *testing.T) {
	res, err := New(&fakeService{}).handleCreateTask(context.Background(), callWith(nil))
	require.NoError(t, err)
	assert.True(t, res.IsError)
}

func TestHandleCreateTaskSurfacesServiceError(t *testing.T) {
	f := &fakeService{createTaskErr: errors.New("backend down")}
	res, err := New(f).handleCreateTask(context.Background(), callWith(map[string]any{"content": "x"}))
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, resultText(t, res), "backend down")
}

func TestHandleCompleteTask(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleCompleteTask(context.Background(), callWith(map[string]any{"id": "t1"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotNil(t, f.lastDoneTask)
	assert.Equal(t, "t1", f.lastDoneTask.Id)
	assert.Contains(t, resultText(t, res), "t1")
}

func TestHandleSearchTasks(t *testing.T) {
	res, err := New(&fakeService{}).handleSearchTasks(context.Background(), callWith(map[string]any{"query": "#work"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Contains(t, resultText(t, res), "#work")
}

func TestHandleMoveTaskPassesArguments(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleMoveTask(context.Background(), callWith(map[string]any{"id": "t1", "list_id": "list-2"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotNil(t, f.lastMoveTask)
	assert.Equal(t, "t1", f.lastMoveTask.TaskId)
	assert.Equal(t, "list-2", f.lastMoveTask.TargetListId)
	assert.Contains(t, resultText(t, res), "list-2")
}

func TestHandleMoveTaskPassesParentTaskID(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleMoveTask(context.Background(), callWith(map[string]any{"id": "t1", "parent_task_id": "t9"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotNil(t, f.lastMoveTask)
	assert.Equal(t, &pb.MoveTaskRequest{TaskId: "t1", TargetParentTaskId: "t9"}, f.lastMoveTask)
}

func TestHandleMoveTaskRequiresIDAndOneTarget(t *testing.T) {
	for _, args := range []map[string]any{nil, {"id": "t1"}, {"list_id": "list-2"}, {"id": "t1", "list_id": "list-2", "parent_task_id": "t9"}} {
		res, err := New(&fakeService{}).handleMoveTask(context.Background(), callWith(args))
		require.NoError(t, err)
		assert.True(t, res.IsError, "args %v should be an error result", args)
	}
}

func TestHandleMoveTaskSurfacesServiceError(t *testing.T) {
	f := &fakeService{moveTaskErr: errors.New("list not found")}
	res, err := New(f).handleMoveTask(context.Background(), callWith(map[string]any{"id": "t1", "list_id": "nope"}))
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, resultText(t, res), "list not found")
}

func TestHandleSnoozeTaskNormalizesUntil(t *testing.T) {
	cases := map[string]string{
		"2026-11-25":                "2026-11-25T00:00:00Z",
		"2026-11-25T09:30:00Z":      "2026-11-25T09:30:00Z",
		"2026-11-25T09:30:00+02:00": "2026-11-25T07:30:00Z",
	}
	for until, want := range cases {
		f := &fakeService{}
		res, err := New(f).handleSnoozeTask(context.Background(), callWith(map[string]any{"id": "t1", "until": until}))
		require.NoError(t, err)
		assert.False(t, res.IsError, "until %q", until)
		require.NotNil(t, f.lastSetMeta)
		assert.Equal(t, "t1", f.lastSetMeta.TaskId)
		assert.Equal(t, "wait", f.lastSetMeta.Field)
		assert.Equal(t, want, f.lastSetMeta.Value)
		assert.Contains(t, resultText(t, res), want)
	}
}

func TestHandleSnoozeTaskClearsWhenUntilEmpty(t *testing.T) {
	for _, args := range []map[string]any{{"id": "t1"}, {"id": "t1", "until": ""}} {
		f := &fakeService{}
		res, err := New(f).handleSnoozeTask(context.Background(), callWith(args))
		require.NoError(t, err)
		assert.False(t, res.IsError)
		require.NotNil(t, f.lastSetMeta)
		assert.Equal(t, "wait", f.lastSetMeta.Field)
		assert.Empty(t, f.lastSetMeta.Value)
		assert.Contains(t, resultText(t, res), "no longer snoozed")
	}
}

func TestHandleSnoozeTaskRejectsBadInput(t *testing.T) {
	for _, args := range []map[string]any{nil, {"until": "2026-11-25"}, {"id": "t1", "until": "next tuesday"}} {
		f := &fakeService{}
		res, err := New(f).handleSnoozeTask(context.Background(), callWith(args))
		require.NoError(t, err)
		assert.True(t, res.IsError, "args %v should be an error result", args)
		assert.Nil(t, f.lastSetMeta, "args %v should not reach the service", args)
	}
}

func TestHandleSnoozeTaskSurfacesServiceError(t *testing.T) {
	f := &fakeService{setMetaErr: errors.New("task not found")}
	res, err := New(f).handleSnoozeTask(context.Background(), callWith(map[string]any{"id": "nope", "until": "2026-11-25"}))
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, resultText(t, res), "task not found")
}

func TestHandleUpdateTaskMetadataPassesArguments(t *testing.T) {
	for _, field := range []string{"notes", "wait", "due", "priority"} {
		f := &fakeService{}
		res, err := New(f).handleUpdateTaskMetadata(context.Background(), callWith(map[string]any{"id": "t1", "field": field, "value": "v"}))
		require.NoError(t, err)
		assert.False(t, res.IsError, "field %q", field)
		require.NotNil(t, f.lastSetMeta)
		assert.Equal(t, &pb.SetTaskMetadataRequest{TaskId: "t1", Field: field, Value: "v"}, f.lastSetMeta)
	}
}

func TestHandleUpdateTaskMetadataClearsWhenValueEmpty(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleUpdateTaskMetadata(context.Background(), callWith(map[string]any{"id": "t1", "field": "notes"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotNil(t, f.lastSetMeta)
	assert.Empty(t, f.lastSetMeta.Value)
}

func TestHandleUpdateTaskMetadataRejectsBadInput(t *testing.T) {
	for _, args := range []map[string]any{nil, {"id": "t1"}, {"field": "notes"}, {"id": "t1", "field": "../../etc"}, {"id": "t1", "field": "id"}} {
		f := &fakeService{}
		res, err := New(f).handleUpdateTaskMetadata(context.Background(), callWith(args))
		require.NoError(t, err)
		assert.True(t, res.IsError, "args %v should be an error result", args)
		assert.Nil(t, f.lastSetMeta, "args %v should not reach the service", args)
	}
}

func TestHandleUpdateTaskMetadataSurfacesServiceError(t *testing.T) {
	f := &fakeService{setMetaErr: errors.New("task not found")}
	res, err := New(f).handleUpdateTaskMetadata(context.Background(), callWith(map[string]any{"id": "nope", "field": "notes", "value": "x"}))
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, resultText(t, res), "task not found")
}

func TestHandleCreateTaskSetsOptionalMetadata(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleCreateTask(context.Background(), callWith(map[string]any{
		"content":    "Plan trip",
		"notes":      "Check passports",
		"wait_until": "2026-12-01",
		"due_date":   "2026-12-10",
		"priority":   "B",
	}))
	require.NoError(t, err)
	assert.False(t, res.IsError, resultText(t, res))
	got := map[string]string{}
	for _, call := range f.setMetaCalls {
		assert.Equal(t, "new", call.TaskId)
		got[call.Field] = call.Value
	}
	assert.Equal(t, map[string]string{"notes": "Check passports", "wait": "2026-12-01T00:00:00Z", "due": "2026-12-10", "priority": "B"}, got)
	text := resultText(t, res)
	for _, want := range []string{`"notes":"Check passports"`, `"waitUntil":"2026-12-01T00:00:00Z"`, `"dueDate":"2026-12-10"`, `"priority":"B"`} {
		assert.Contains(t, text, want)
	}
}

func TestHandleCreateTaskWithoutMetadataSetsNone(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleCreateTask(context.Background(), callWith(map[string]any{"content": "Plain"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Empty(t, f.setMetaCalls)
}

func TestHandleCreateTaskRejectsBadWaitBeforeCreating(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleCreateTask(context.Background(), callWith(map[string]any{"content": "x", "wait_until": "soon"}))
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Nil(t, f.lastCreateTask, "no task should be created when arguments are invalid")
}

func TestHandleCreateTaskReportsMetadataFailureWithTaskID(t *testing.T) {
	f := &fakeService{setMetaErr: errors.New("disk full")}
	res, err := New(f).handleCreateTask(context.Background(), callWith(map[string]any{"content": "x", "notes": "n"}))
	require.NoError(t, err)
	assert.True(t, res.IsError)
	text := resultText(t, res)
	assert.Contains(t, text, "new", "the error should name the task that was created")
	assert.Contains(t, text, "disk full")
}

func TestHandleDeleteTask(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleDeleteTask(context.Background(), callWith(map[string]any{"id": "t1"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotNil(t, f.lastDeleteTask)
	assert.Equal(t, "t1", f.lastDeleteTask.Id)
	assert.Contains(t, resultText(t, res), "t1")

	res, err = New(&fakeService{}).handleDeleteTask(context.Background(), callWith(nil))
	require.NoError(t, err)
	assert.True(t, res.IsError)

	res, err = New(&fakeService{deleteErr: errors.New("task not found")}).handleDeleteTask(context.Background(), callWith(map[string]any{"id": "x"}))
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, resultText(t, res), "task not found")
}

func TestHandleDeleteList(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleDeleteList(context.Background(), callWith(map[string]any{"id": "list-2"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	require.NotNil(t, f.lastDeleteList)
	assert.Equal(t, "list-2", f.lastDeleteList.Id)

	res, err = New(&fakeService{}).handleDeleteList(context.Background(), callWith(nil))
	require.NoError(t, err)
	assert.True(t, res.IsError)

	res, err = New(&fakeService{deleteErr: errors.New("the default list cannot be deleted")}).handleDeleteList(context.Background(), callWith(map[string]any{"id": "inbox"}))
	require.NoError(t, err)
	assert.True(t, res.IsError)
	assert.Contains(t, resultText(t, res), "default list")
}

func TestHandleListCompletedTasks(t *testing.T) {
	f := &fakeService{}
	res, err := New(f).handleListCompletedTasks(context.Background(), callWith(map[string]any{"since": "2026-11-21", "list_id": "list-1"}))
	require.NoError(t, err)
	assert.False(t, res.IsError)
	assert.Equal(t, &pb.ListCompletedTasksRequest{Since: "2026-11-21", ListId: "list-1"}, f.lastCompleted)
	assert.Contains(t, resultText(t, res), `"completed":true`)
}
