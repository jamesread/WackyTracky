package clientapi

import (
	"context"
	"testing"

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
	api := GetNewClientAPI(db)
	require.NotNil(t, api)
}

func TestClientAPI_ListTasks(t *testing.T) {
	db := &dummy.Dummy{}
	require.NoError(t, db.Connect())
	api := GetNewClientAPI(db)

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
	api := GetNewClientAPI(db)

	req := connect.NewRequest(&pb.GetListsRequest{})
	resp, err := api.GetLists(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Msg.Lists, 2)
}

func TestClientAPI_GetTags(t *testing.T) {
	db := &dummy.Dummy{}
	require.NoError(t, db.Connect())
	api := GetNewClientAPI(db)

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
	api := GetNewClientAPI(db)

	req := connect.NewRequest(&pb.VersionRequest{})
	resp, err := api.Version(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.NotEmpty(t, resp.Msg.Version)
}

func TestClientAPI_RepoSync_WithDummyDB(t *testing.T) {
	db := &dummy.Dummy{}
	require.NoError(t, db.Connect())
	api := GetNewClientAPI(db)

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

	api := GetNewClientAPI(&todotxt.TodoTxt{})

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
