package dummy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDummy_Connect(t *testing.T) {
	db := &Dummy{}
	err := db.Connect()
	require.NoError(t, err)
}

func TestDummy_GetLists(t *testing.T) {
	db := &Dummy{}
	require.NoError(t, db.Connect())

	lists, err := db.GetLists()
	require.NoError(t, err)
	require.Len(t, lists, 2)
	assert.Equal(t, "First list", lists[0].Title)
	assert.Equal(t, "1", lists[0].ID)
	assert.Equal(t, "Second list", lists[1].Title)
}

func TestDummy_GetTasks(t *testing.T) {
	db := &Dummy{}
	require.NoError(t, db.Connect())

	tasks, err := db.GetTasks("1")
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "First List Task One", tasks[0].Content)
	assert.Equal(t, "1", tasks[0].ID)

	tasks, err = db.GetTasks("2")
	require.NoError(t, err)
	require.Len(t, tasks, 2)
}

func TestDummy_GetSubtasks(t *testing.T) {
	db := &Dummy{}
	require.NoError(t, db.Connect())

	subID, err := db.CreateTask("subtask under 1", "", "1")
	require.NoError(t, err)
	require.NotEmpty(t, subID)

	subtasks, err := db.GetSubtasks("1")
	require.NoError(t, err)
	require.Len(t, subtasks, 1)
	assert.Equal(t, "subtask under 1", subtasks[0].Content)
}

func TestDummy_GetTask(t *testing.T) {
	db := &Dummy{}
	require.NoError(t, db.Connect())

	task, err := db.GetTask("1")
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "First List Task One", task.Content)

	task, err = db.GetTask("nonexistent")
	require.NoError(t, err)
	assert.Nil(t, task)
}

func TestDummy_GetTags(t *testing.T) {
	db := &Dummy{}
	require.NoError(t, db.Connect())

	tags, err := db.GetTags()
	require.NoError(t, err)
	require.Len(t, tags, 1)
	assert.Equal(t, "First tag", tags[0].Title)
	assert.Equal(t, "1", tags[0].ID)
}

func TestDummy_CreateList(t *testing.T) {
	db := &Dummy{}
	require.NoError(t, db.Connect())

	err := db.CreateList("New list")
	require.NoError(t, err)

	lists, _ := db.GetLists()
	var found bool
	for _, l := range lists {
		if l.Title == "New list" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestDummy_CreateTask(t *testing.T) {
	db := &Dummy{}
	require.NoError(t, db.Connect())

	id, err := db.CreateTask("new task", "1", "")
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	task, _ := db.GetTask(id)
	require.NotNil(t, task)
	assert.Equal(t, "new task", task.Content)
	assert.Equal(t, "1", task.ParentId)

	subID, err := db.CreateTask("subtask", "", "1")
	require.NoError(t, err)
	task, _ = db.GetTask(subID)
	require.NotNil(t, task)
	assert.Equal(t, "task", task.ParentType)
	assert.Equal(t, "1", task.ParentId)
}

func TestDummy_UpdateList(t *testing.T) {
	db := &Dummy{}
	require.NoError(t, db.Connect())

	err := db.UpdateList("1", "Updated First list")
	require.NoError(t, err)

	lists, _ := db.GetLists()
	assert.Equal(t, "Updated First list", lists[0].Title)
}

func TestDummy_DeleteList(t *testing.T) {
	db := &Dummy{}
	require.NoError(t, db.Connect())

	err := db.DeleteList("2")
	require.NoError(t, err)

	lists, _ := db.GetLists()
	require.Len(t, lists, 1)
	assert.Equal(t, "1", lists[0].ID)
}
