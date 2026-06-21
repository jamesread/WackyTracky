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
	createTaskErr  error
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

func (f *fakeService) CreateList(_ context.Context, _ *connect.Request[pb.CreateListRequest]) (*connect.Response[pb.CreateListResponse], error) {
	return connect.NewResponse(&pb.CreateListResponse{}), nil
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
