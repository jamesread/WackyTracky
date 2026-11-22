package clientapi

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pb "github.com/wacky-tracky/wacky-tracky-server/gen/wacky-tracky/clientapi/v1"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/dummy"
	"github.com/wacky-tracky/wacky-tracky-server/internal/db/todotxt"
	. "github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

func TestGetNewClientAPI_WithDummyDB(t *testing.T) {
	db := &dummy.Dummy{}
	api := GetNewClientAPI(db, nil)
	require.NotNil(t, api)
}

func TestClientAPI_ListTasks(t *testing.T) {
	db := &dummy.Dummy{}
	require.NoError(t, db.Connect())
	api := GetNewClientAPI(db, nil)

	req := connect.NewRequest(&pb.ListTasksRequest{ParentId: "1", ParentType: "list"})
	resp, err := api.ListTasks(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Msg.Tasks, 1)
	assert.Equal(t, "First List Task One", resp.Msg.Tasks[0].Content)
	assert.Equal(t, "1", resp.Msg.Tasks[0].Id)
}

func TestClientAPI_GetLists(t *testing.T) {
	db := &dummy.Dummy{}
	require.NoError(t, db.Connect())
	api := GetNewClientAPI(db, nil)

	req := connect.NewRequest(&pb.GetListsRequest{})
	resp, err := api.GetLists(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Msg.Lists, 2)
}

func TestClientAPI_GetTags(t *testing.T) {
	db := &dummy.Dummy{}
	require.NoError(t, db.Connect())
	api := GetNewClientAPI(db, nil)

	req := connect.NewRequest(&pb.GetTagsRequest{})
	resp, err := api.GetTags(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Msg.Tags, 1)
	assert.Equal(t, "First tag", resp.Msg.Tags[0].Title)
}

func TestClientAPI_Version(t *testing.T) {
	db := &dummy.Dummy{}
	require.NoError(t, db.Connect())
	api := GetNewClientAPI(db, nil)

	req := connect.NewRequest(&pb.VersionRequest{})
	resp, err := api.Version(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Msg.Version)
}

func TestClientAPI_Init_FooterFields(t *testing.T) {
	db := &dummy.Dummy{}
	require.NoError(t, db.Connect())
	api := GetNewClientAPI(db, nil)

	origFooter := RuntimeConfig.ShowFooter
	origNew := RuntimeConfig.ShowNewVersions
	origVer := RuntimeConfig.ShowVersionNumber
	t.Cleanup(func() {
		RuntimeConfig.ShowFooter = origFooter
		RuntimeConfig.ShowNewVersions = origNew
		RuntimeConfig.ShowVersionNumber = origVer
	})

	RuntimeConfig.ShowFooter = true
	RuntimeConfig.ShowNewVersions = true
	RuntimeConfig.ShowVersionNumber = true

	resp, err := api.Init(context.Background(), connect.NewRequest(&pb.InitRequest{}))
	require.NoError(t, err)
	require.True(t, resp.Msg.ShowFooter)
	require.True(t, resp.Msg.ShowVersionNumber)
	require.True(t, resp.Msg.ShowNewVersions)
	assert.NotEmpty(t, resp.Msg.CurrentVersion)
	assert.Equal(t, "WackyTracky", resp.Msg.PageTitle)
	assert.Empty(t, resp.Msg.AvailableVersion)

	RuntimeConfig.ShowVersionNumber = false
	resp, err = api.Init(context.Background(), connect.NewRequest(&pb.InitRequest{}))
	require.NoError(t, err)
	assert.False(t, resp.Msg.ShowVersionNumber)
	assert.False(t, resp.Msg.ShowNewVersions)
	assert.Empty(t, resp.Msg.CurrentVersion)
	assert.Empty(t, resp.Msg.AvailableVersion)

	RuntimeConfig.ShowFooter = false
	RuntimeConfig.ShowVersionNumber = true
	resp, err = api.Init(context.Background(), connect.NewRequest(&pb.InitRequest{}))
	require.NoError(t, err)
	assert.False(t, resp.Msg.ShowFooter)
}

func TestClientAPI_RepoSync_WithDummyDB(t *testing.T) {
	db := &dummy.Dummy{}
	require.NoError(t, db.Connect())
	api := GetNewClientAPI(db, nil)

	req := connect.NewRequest(&pb.RepoSyncRequest{})
	resp, err := api.RepoSync(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Msg.Success)
	assert.Contains(t, resp.Msg.Message, "todo.txt backend")
}

func TestClientAPI_CreateTaskReturnsParsedMetadata(t *testing.T) {
	dir := t.TempDir()
	orig := RuntimeConfig.Database.Database
	RuntimeConfig.Database.Database = dir
	t.Cleanup(func() { RuntimeConfig.Database.Database = orig })

	api := GetNewClientAPI(&todotxt.TodoTxt{}, nil)

	resp, err := api.CreateTask(context.Background(), connect.NewRequest(&pb.CreateTaskRequest{
		Content:      "ship feature due:2025-07-15",
		ParentListId: "inbox",
	}))
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.Task)
	assert.NotEmpty(t, resp.Msg.Task.Id)
	assert.Equal(t, "2025-07-15", resp.Msg.Task.DueDate)
	assert.Equal(t, "ship feature", resp.Msg.Task.Content)
}

func newTodoTxtClientAPI(t *testing.T) *wackyTrackyClientService {
	t.Helper()
	orig := RuntimeConfig.Database.Database
	RuntimeConfig.Database.Database = t.TempDir()
	t.Cleanup(func() { RuntimeConfig.Database.Database = orig })
	return GetNewClientAPI(&todotxt.TodoTxt{}, nil)
}

func TestClientAPI_MoveTaskReturnsMovedTask(t *testing.T) {
	api := newTodoTxtClientAPI(t)
	ctx := context.Background()

	_, err := api.CreateList(ctx, connect.NewRequest(&pb.CreateListRequest{Title: "Other"}))
	require.NoError(t, err)
	lists, err := api.GetLists(ctx, connect.NewRequest(&pb.GetListsRequest{}))
	require.NoError(t, err)
	var otherID string
	for _, l := range lists.Msg.Lists {
		if l.Title == "Other" {
			otherID = l.Id
		}
	}
	require.NotEmpty(t, otherID)

	created, err := api.CreateTask(ctx, connect.NewRequest(&pb.CreateTaskRequest{Content: "move me", ParentListId: "inbox"}))
	require.NoError(t, err)

	resp, err := api.MoveTask(ctx, connect.NewRequest(&pb.MoveTaskRequest{TaskId: created.Msg.Task.Id, TargetListId: otherID}))
	require.NoError(t, err)
	require.NotNil(t, resp.Msg.Task)
	assert.Equal(t, created.Msg.Task.Id, resp.Msg.Task.Id)
	assert.Equal(t, otherID, resp.Msg.Task.ParentId)
	assert.Equal(t, "move me", resp.Msg.Task.Content)
}

func TestClientAPI_MoveTaskErrorCodes(t *testing.T) {
	api := newTodoTxtClientAPI(t)
	ctx := context.Background()

	_, err := api.MoveTask(ctx, connect.NewRequest(&pb.MoveTaskRequest{TaskId: "x"}))
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))

	_, err = api.MoveTask(ctx, connect.NewRequest(&pb.MoveTaskRequest{TaskId: "missing", TargetListId: "inbox"}))
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestClientAPI_MoveTaskUnimplementedForBackendWithoutMover(t *testing.T) {
	api := GetNewClientAPI(&dummy.Dummy{}, nil)
	_, err := api.MoveTask(context.Background(), connect.NewRequest(&pb.MoveTaskRequest{TaskId: "x", TargetListId: "y"}))
	assert.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
}

func TestClientAPI_MutationsReturnErrorCodes(t *testing.T) {
	api := newTodoTxtClientAPI(t)
	ctx := context.Background()

	cases := []struct {
		name string
		call func() error
		code connect.Code
	}{
		{"DoneTask missing id", func() error {
			_, err := api.DoneTask(ctx, connect.NewRequest(&pb.DoneTaskRequest{}))
			return err
		}, connect.CodeInvalidArgument},
		{"DoneTask unknown task", func() error {
			_, err := api.DoneTask(ctx, connect.NewRequest(&pb.DoneTaskRequest{Id: "missing"}))
			return err
		}, connect.CodeNotFound},
		{"UpdateTask unknown task", func() error {
			_, err := api.UpdateTask(ctx, connect.NewRequest(&pb.UpdateTaskRequest{Id: "missing", Content: "x"}))
			return err
		}, connect.CodeNotFound},
		{"UpdateList unknown list", func() error {
			_, err := api.UpdateList(ctx, connect.NewRequest(&pb.UpdateListRequest{Id: "missing", Title: "x"}))
			return err
		}, connect.CodeNotFound},
		{"DeleteList missing id", func() error {
			_, err := api.DeleteList(ctx, connect.NewRequest(&pb.DeleteListRequest{}))
			return err
		}, connect.CodeInvalidArgument},
		{"DeleteList inbox", func() error {
			_, err := api.DeleteList(ctx, connect.NewRequest(&pb.DeleteListRequest{Id: "inbox"}))
			return err
		}, connect.CodeFailedPrecondition},
		{"SetTaskMetadata unknown task", func() error {
			_, err := api.SetTaskMetadata(ctx, connect.NewRequest(&pb.SetTaskMetadataRequest{TaskId: "missing", Field: "notes"}))
			return err
		}, connect.CodeNotFound},
		{"SetTaskMetadata missing field", func() error {
			_, err := api.SetTaskMetadata(ctx, connect.NewRequest(&pb.SetTaskMetadataRequest{TaskId: "t"}))
			return err
		}, connect.CodeInvalidArgument},
		{"SetTaskPropertyProperty bad type", func() error {
			_, err := api.SetTaskPropertyProperty(ctx, connect.NewRequest(&pb.SetTaskPropertyPropertyRequest{PropertyType: "colour", PropertyName: "work", Key: "k"}))
			return err
		}, connect.CodeInvalidArgument},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.code, connect.CodeOf(tc.call()))
		})
	}
}

func TestClientAPI_DoneTaskSucceedsForExistingTask(t *testing.T) {
	api := newTodoTxtClientAPI(t)
	ctx := context.Background()
	created, err := api.CreateTask(ctx, connect.NewRequest(&pb.CreateTaskRequest{Content: "x", ParentListId: "inbox"}))
	require.NoError(t, err)

	_, err = api.DoneTask(ctx, connect.NewRequest(&pb.DoneTaskRequest{Id: created.Msg.Task.Id}))
	assert.NoError(t, err)
}

func createTestTask(t *testing.T, api *wackyTrackyClientService, content, parentTaskID string) string {
	t.Helper()
	resp, err := api.CreateTask(context.Background(), connect.NewRequest(&pb.CreateTaskRequest{Content: content, ParentListId: "inbox", ParentTaskId: parentTaskID}))
	require.NoError(t, err)
	return resp.Msg.Task.Id
}

func TestClientAPI_MoveTaskToParentTask(t *testing.T) {
	api := newTodoTxtClientAPI(t)
	parentID := createTestTask(t, api, "parent", "")
	taskID := createTestTask(t, api, "task", "")

	resp, err := api.MoveTask(context.Background(), connect.NewRequest(&pb.MoveTaskRequest{TaskId: taskID, TargetParentTaskId: parentID}))
	require.NoError(t, err)
	assert.Equal(t, "task", resp.Msg.Task.ParentType)
	assert.Equal(t, parentID, resp.Msg.Task.ParentId)
}

func TestClientAPI_MoveTaskTargetValidation(t *testing.T) {
	api := newTodoTxtClientAPI(t)
	ctx := context.Background()
	taskID := createTestTask(t, api, "task", "")

	_, err := api.MoveTask(ctx, connect.NewRequest(&pb.MoveTaskRequest{TaskId: taskID, TargetListId: "inbox", TargetParentTaskId: taskID}))
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "both targets")

	_, err = api.MoveTask(ctx, connect.NewRequest(&pb.MoveTaskRequest{TaskId: taskID, TargetParentTaskId: taskID}))
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err), "under itself")

	_, err = api.MoveTask(ctx, connect.NewRequest(&pb.MoveTaskRequest{TaskId: taskID, TargetParentTaskId: "missing"}))
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
}

func TestClientAPI_DeleteTask(t *testing.T) {
	api := newTodoTxtClientAPI(t)
	ctx := context.Background()
	taskID := createTestTask(t, api, "delete me", "")

	_, err := api.DeleteTask(ctx, connect.NewRequest(&pb.DeleteTaskRequest{Id: taskID}))
	require.NoError(t, err)

	_, err = api.DeleteTask(ctx, connect.NewRequest(&pb.DeleteTaskRequest{Id: taskID}))
	assert.Equal(t, connect.CodeNotFound, connect.CodeOf(err))
	_, err = api.DeleteTask(ctx, connect.NewRequest(&pb.DeleteTaskRequest{}))
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	_, err = GetNewClientAPI(&dummy.Dummy{}, nil).DeleteTask(ctx, connect.NewRequest(&pb.DeleteTaskRequest{Id: "x"}))
	assert.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
}

func TestClientAPI_ListCompletedTasks(t *testing.T) {
	api := newTodoTxtClientAPI(t)
	ctx := context.Background()
	doneID := createTestTask(t, api, "finished", "")
	createTestTask(t, api, "still open", "")
	_, err := api.DoneTask(ctx, connect.NewRequest(&pb.DoneTaskRequest{Id: doneID}))
	require.NoError(t, err)
	today := time.Now().UTC().Format(time.DateOnly)

	for _, since := range []string{"", today, today + "T00:00:00Z"} {
		resp, err := api.ListCompletedTasks(ctx, connect.NewRequest(&pb.ListCompletedTasksRequest{Since: since}))
		require.NoError(t, err, "since %q", since)
		require.Len(t, resp.Msg.Tasks, 1, "since %q", since)
		assert.Equal(t, doneID, resp.Msg.Tasks[0].Id)
		assert.True(t, resp.Msg.Tasks[0].Completed)
		assert.Equal(t, today, resp.Msg.Tasks[0].CompletedAt)
	}

	resp, err := api.ListCompletedTasks(ctx, connect.NewRequest(&pb.ListCompletedTasksRequest{Since: "2999-01-01"}))
	require.NoError(t, err)
	assert.Empty(t, resp.Msg.Tasks)

	_, err = api.ListCompletedTasks(ctx, connect.NewRequest(&pb.ListCompletedTasksRequest{Since: "yesterday"}))
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	_, err = GetNewClientAPI(&dummy.Dummy{}, nil).ListCompletedTasks(ctx, connect.NewRequest(&pb.ListCompletedTasksRequest{}))
	assert.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
}

func TestClientAPI_TasksIncludeNotes(t *testing.T) {
	api := newTodoTxtClientAPI(t)
	ctx := context.Background()
	taskID := createTestTask(t, api, "with notes", "")
	_, err := api.SetTaskMetadata(ctx, connect.NewRequest(&pb.SetTaskMetadataRequest{TaskId: taskID, Field: "notes", Value: "remember this"}))
	require.NoError(t, err)

	list, err := api.ListTasks(ctx, connect.NewRequest(&pb.ListTasksRequest{ParentId: "inbox", ParentType: "list"}))
	require.NoError(t, err)
	require.Len(t, list.Msg.Tasks, 1)
	assert.Equal(t, "remember this", list.Msg.Tasks[0].Notes)
	assert.False(t, list.Msg.Tasks[0].Completed)
}
