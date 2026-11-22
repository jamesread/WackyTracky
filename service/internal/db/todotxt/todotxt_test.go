package todotxt

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	db "github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
	. "github.com/wacky-tracky/wacky-tracky-server/internal/runtimeconfig"
)

// setTodotxtDir sets RuntimeConfig.Database.Database to a temp dir and returns a cleanup that restores the original.
func setTodotxtDir(t *testing.T) (dir string, cleanup func()) {
	t.Helper()
	dir = t.TempDir()
	orig := RuntimeConfig.Database.Database
	RuntimeConfig.Database.Database = dir
	return dir, func() {
		RuntimeConfig.Database.Database = orig
	}
}

func TestTodoTxt_ConnectAndLoadTasks(t *testing.T) {
	dir, cleanup := setTodotxtDir(t)
	defer cleanup()

	// Write a minimal todo.txt
	todoPath := filepath.Join(dir, "todo.txt")
	if err := os.WriteFile(todoPath, []byte("(A) first task\nsecond task +project\n"), 0644); err != nil {
		t.Fatal(err)
	}

	d := &TodoTxt{}
	if err := d.Connect(); err != nil {
		t.Fatal(err)
	}

	tasks := d.tasks
	if len(tasks) < 2 {
		t.Fatalf("expected at least 2 tasks, got %d", len(tasks))
	}
	if len(tasks[0].Priority) < 3 || tasks[0].Priority[1] != 'A' || tasks[0].Description != "first task" {
		t.Errorf("first task: got %+v", tasks[0])
	}
	if tasks[1].Description != "second task" || len(tasks[1].Projects) != 1 || tasks[1].Projects[0] != "project" {
		t.Errorf("second task: got %+v", tasks[1])
	}
}

func TestTodoTxt_CreateTaskAndGetTasks(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	if err := d.Connect(); err != nil {
		t.Fatal(err)
	}

	id, err := d.CreateTask("new task +work", "inbox", "")
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Error("expected non-empty id")
	}

	roots, err := d.GetTasks("inbox")
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root task, got %d", len(roots))
	}
	if roots[0].ID != id || roots[0].Content != "new task +work" {
		t.Errorf("got task %+v", roots[0])
	}
}

func TestTodoTxt_CreateTaskReturnsInlineMetadata(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	require.NoError(t, d.Connect())

	id, err := d.CreateTask("(B) finish report due:2025-12-31 wait:2025-06-01 @work #urgent", "inbox", "")
	require.NoError(t, err)

	task, err := d.GetTask(id)
	require.NoError(t, err)
	require.NotNil(t, task)
	assert.Equal(t, "B", task.Priority)
	assert.Equal(t, "2025-12-31", task.DueDate)
	assert.Equal(t, "2025-06-01", task.WaitUntil)
	assert.Equal(t, []string{"work"}, task.Contexts)
	assert.Equal(t, []string{"urgent"}, task.Tags)
	assert.Equal(t, "finish report", task.Content)
}

func TestTodoTxt_SearchTasks(t *testing.T) {
	dir, cleanup := setTodotxtDir(t)
	defer cleanup()

	// Write tasks so we don't rely on CreateTask
	todoPath := filepath.Join(dir, "todo.txt")
	content := "buy milk\nbuy bread\ncall mom +family\nmeet @work"
	if err := os.WriteFile(todoPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	d := &TodoTxt{}
	if err := d.Connect(); err != nil {
		t.Fatal(err)
	}

	// Search for "buy" should match first two
	results, err := d.SearchTasks("buy")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("search 'buy': expected 2, got %d", len(results))
	}

	// Search with negation
	results, err = d.SearchTasks("buy -bread")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("search 'buy -bread': expected 1, got %d", len(results))
	}
	if results[0].Content != "buy milk" {
		t.Errorf("expected 'buy milk', got %q", results[0].Content)
	}

	// Search for context (DBTask.Content is description only; contexts are in Contexts)
	results, err = d.SearchTasks("@work")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Content != "meet" {
		t.Errorf("search @work: got %d results, content %q", len(results), results[0].Content)
	}
	if len(results[0].Contexts) != 1 || results[0].Contexts[0] != "work" {
		t.Errorf("search @work: expected context work, got %v", results[0].Contexts)
	}
}

func TestTodoTxt_UpdateTaskAndDoneTask(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	if err := d.Connect(); err != nil {
		t.Fatal(err)
	}

	id, err := d.CreateTask("original", "inbox", "")
	if err != nil {
		t.Fatal(err)
	}

	if err = d.UpdateTask(id, "updated content"); err != nil {
		t.Fatal(err)
	}
	task, err := d.GetTask(id)
	if err != nil || task == nil {
		t.Fatalf("GetTask: err=%v task=%v", err, task)
	}
	if task.Content != "updated content" {
		t.Errorf("expected updated content, got %q", task.Content)
	}

	if err := d.DoneTask(id); err != nil {
		t.Fatal(err)
	}
	roots, _ := d.GetTasks("inbox")
	if len(roots) != 0 {
		t.Errorf("done task should not appear in GetTasks, got %d", len(roots))
	}
}

func TestTodoTxt_MoveTask(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	if err := d.Connect(); err != nil {
		t.Fatal(err)
	}

	require.NoError(t, d.CreateList("Other"))
	lists, err := d.GetLists()
	require.NoError(t, err)
	var otherListID string
	for _, l := range lists {
		if l.Title == "Other" {
			otherListID = l.ID
			break
		}
	}
	require.NotEmpty(t, otherListID, "expected Other list")

	id, err := d.CreateTask("move me", "inbox", "")
	require.NoError(t, err)

	inboxBefore, _ := d.GetTasks("inbox")
	otherBefore, _ := d.GetTasks(otherListID)
	require.Len(t, inboxBefore, 1)
	require.Empty(t, otherBefore)

	require.NoError(t, d.MoveTask(id, otherListID))

	inboxAfter, _ := d.GetTasks("inbox")
	otherAfter, _ := d.GetTasks(otherListID)
	require.Empty(t, inboxAfter, "task should no longer be in inbox")
	require.Len(t, otherAfter, 1, "task should appear in Other list")
	assert.Equal(t, "move me", otherAfter[0].Content)
}

func connectTodoTxtWithOtherList(t *testing.T) (*TodoTxt, string) {
	t.Helper()
	_, cleanup := setTodotxtDir(t)
	t.Cleanup(cleanup)
	d := &TodoTxt{}
	require.NoError(t, d.Connect())
	require.NoError(t, d.CreateList("Other"))
	lists, err := d.GetLists()
	require.NoError(t, err)
	for _, l := range lists {
		if l.Title == "Other" {
			return d, l.ID
		}
	}
	t.Fatal("expected Other list")
	return nil, ""
}

func TestTodoTxt_MoveTaskKeepsIDAndMovesSubtasks(t *testing.T) {
	d, otherListID := connectTodoTxtWithOtherList(t)

	parentID, err := d.CreateTask("parent", "inbox", "")
	require.NoError(t, err)
	childID, err := d.CreateTask("child", "inbox", parentID)
	require.NoError(t, err)
	grandchildID, err := d.CreateTask("grandchild", "inbox", childID)
	require.NoError(t, err)

	require.NoError(t, d.MoveTask(parentID, otherListID))

	moved, err := d.GetTask(parentID)
	require.NoError(t, err)
	require.NotNil(t, moved)
	assert.Equal(t, otherListID, moved.ParentId)

	children, err := d.GetSubtasks(parentID)
	require.NoError(t, err)
	require.Len(t, children, 1)
	assert.Equal(t, childID, children[0].ID)

	for _, task := range d.tasks {
		if id := task.Metadata["id"]; id == childID || id == grandchildID {
			assert.Equal(t, otherListID, task.Metadata["listid"], "descendant %s should follow its root", id)
		}
	}
}

func TestTodoTxt_MoveSubtaskBecomesRootInTargetList(t *testing.T) {
	d, otherListID := connectTodoTxtWithOtherList(t)

	parentID, err := d.CreateTask("parent", "inbox", "")
	require.NoError(t, err)
	childID, err := d.CreateTask("child", "inbox", parentID)
	require.NoError(t, err)

	require.NoError(t, d.MoveTask(childID, otherListID))

	roots, _ := d.GetTasks(otherListID)
	require.Len(t, roots, 1)
	assert.Equal(t, childID, roots[0].ID)
	assert.Equal(t, "list", roots[0].ParentType)

	remaining, _ := d.GetSubtasks(parentID)
	assert.Empty(t, remaining)
}

func TestTodoTxt_MoveTaskErrors(t *testing.T) {
	d, otherListID := connectTodoTxtWithOtherList(t)

	id, err := d.CreateTask("task", "inbox", "")
	require.NoError(t, err)

	require.ErrorIs(t, d.MoveTask("missing", otherListID), db.ErrTaskNotFound)
	require.ErrorIs(t, d.MoveTask(id, "missing-list"), db.ErrListNotFound)

	require.NoError(t, d.DoneTask(id))
	require.ErrorIs(t, d.MoveTask(id, otherListID), db.ErrTaskNotFound, "completed tasks cannot be moved")
}

func TestTodoTxt_GetSetTaskPropertyProperties(t *testing.T) {
	dir, cleanup := setTodotxtDir(t)
	defer cleanup()

	todoPath := filepath.Join(dir, "todo.txt")
	if err := os.WriteFile(todoPath, []byte("task #work @home\n"), 0644); err != nil {
		t.Fatal(err)
	}

	d := &TodoTxt{}
	if err := d.Connect(); err != nil {
		t.Fatal(err)
	}

	var store db.TaskPropertyPropertiesStore = d

	tagProps, ctxProps, err := store.GetTaskPropertyProperties()
	require.NoError(t, err)
	assert.Empty(t, tagProps)
	assert.Empty(t, ctxProps)

	require.NoError(t, store.SetTaskPropertyProperty("tag", "work", "bgcolor", "#ff0000"))
	require.NoError(t, store.SetTaskPropertyProperty("context", "home", "bgcolor", "#00ff00"))

	tagProps, ctxProps, err = store.GetTaskPropertyProperties()
	require.NoError(t, err)
	assert.Equal(t, "#ff0000", tagProps["work"]["bgcolor"])
	assert.Equal(t, "#00ff00", ctxProps["home"]["bgcolor"])

	tppPath := filepath.Join(dir, tppFilename)
	b, err := os.ReadFile(tppPath)
	require.NoError(t, err)
	assert.Contains(t, string(b), "work")
	assert.Contains(t, string(b), "#ff0000")
	assert.Contains(t, string(b), "home")
	assert.Contains(t, string(b), "#00ff00")

	require.NoError(t, store.SetTaskPropertyProperty("tag", "work", "bgcolor", ""))
	tagProps, _, _ = store.GetTaskPropertyProperties()
	_, hasWork := tagProps["work"]
	assert.False(t, hasWork)
}

func TestTodoTxt_GetTags(t *testing.T) {
	dir, cleanup := setTodotxtDir(t)
	defer cleanup()

	todoPath := filepath.Join(dir, "todo.txt")
	if err := os.WriteFile(todoPath, []byte("task @ctx #tag\nother +proj @work\n"), 0644); err != nil {
		t.Fatal(err)
	}

	d := &TodoTxt{}
	require.NoError(t, d.Connect())

	tags, err := d.GetTags()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(tags), 2)
	var tagIDs []string
	for _, tag := range tags {
		tagIDs = append(tagIDs, tag.ID)
	}
	assert.Contains(t, tagIDs, "tag")
	assert.Contains(t, tagIDs, "ctx")
	assert.Contains(t, tagIDs, "work")
}

func TestTodoTxt_GetSubtasks(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	require.NoError(t, d.Connect())

	parentID, err := d.CreateTask("parent task", "inbox", "")
	require.NoError(t, err)

	subID, err := d.CreateTask("subtask", "inbox", parentID)
	require.NoError(t, err)

	subtasks, err := d.GetSubtasks(parentID)
	require.NoError(t, err)
	require.Len(t, subtasks, 1)
	assert.Equal(t, subID, subtasks[0].ID)
	assert.Equal(t, "subtask", subtasks[0].Content)
}

func TestTodoTxt_Dir(t *testing.T) {
	dir, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	require.NoError(t, d.Connect())

	assert.Equal(t, dir, d.Dir())
}

func TestTodoTxt_UpdateList(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	require.NoError(t, d.Connect())

	require.NoError(t, d.CreateList("Original"))
	lists, _ := d.GetLists()
	var listID string
	for _, l := range lists {
		if l.Title == "Original" {
			listID = l.ID
			break
		}
	}
	require.NotEmpty(t, listID)

	require.NoError(t, d.UpdateList(listID, "Updated"))
	lists, _ = d.GetLists()
	for _, l := range lists {
		if l.ID == listID {
			assert.Equal(t, "Updated", l.Title)
			return
		}
	}
	t.Error("list not found after update")
}

func TestTodoTxt_DeleteList(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	require.NoError(t, d.Connect())

	require.NoError(t, d.CreateList("ToDelete"))
	lists, _ := d.GetLists()
	var listID string
	for _, l := range lists {
		if l.Title == "ToDelete" {
			listID = l.ID
			break
		}
	}
	require.NotEmpty(t, listID)

	id, _ := d.CreateTask("task in list", listID, "")
	require.NotEmpty(t, id)

	require.NoError(t, d.DeleteList(listID))
	lists, _ = d.GetLists()
	for _, l := range lists {
		assert.NotEqual(t, listID, l.ID)
	}
	task, _ := d.GetTask(id)
	require.NotNil(t, task)
	assert.Equal(t, "inbox", task.ParentId)
}

func TestTodoTxt_GetSetSavedSearches(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	require.NoError(t, d.Connect())

	searches, err := d.GetSavedSearches()
	require.NoError(t, err)
	assert.Empty(t, searches)

	newSearches := []db.SavedSearch{
		{ID: "s1", Title: "Work", Query: "+work"},
		{ID: "s2", Title: "Home", Query: "@home"},
	}
	require.NoError(t, d.SetSavedSearches(newSearches))

	searches, err = d.GetSavedSearches()
	require.NoError(t, err)
	require.Len(t, searches, 2)
	assert.Equal(t, "Work", searches[0].Title)
	assert.Equal(t, "+work", searches[0].Query)
	assert.Equal(t, "Home", searches[1].Title)
}

func TestTodoTxt_GetSetTaskMetadata(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()

	d := &TodoTxt{}
	require.NoError(t, d.Connect())

	id, err := d.CreateTask("task with metadata", "inbox", "")
	require.NoError(t, err)

	meta, err := d.GetTaskMetadata(id)
	require.NoError(t, err)
	assert.NotNil(t, meta)

	require.NoError(t, d.SetTaskMetadata(id, "notes", "my note"))
	require.NoError(t, d.SetTaskMetadata(id, "wait", "2025-01-01"))
	require.NoError(t, d.SetTaskMetadata(id, "due", "2025-12-31"))
	require.NoError(t, d.SetTaskMetadata(id, "priority", "A"))

	meta, err = d.GetTaskMetadata(id)
	require.NoError(t, err)
	assert.Equal(t, "my note", meta["notes"])
	assert.Equal(t, "2025-01-01", meta["wait"])
	assert.Equal(t, "2025-12-31", meta["due"])
	assert.Equal(t, "A", meta["priority"])
}

func TestParseSearchQuery(t *testing.T) {
	inc, exc := parseSearchQuery("")
	assert.Nil(t, inc)
	assert.Nil(t, exc)

	inc, exc = parseSearchQuery("  hello world  ")
	assert.Equal(t, []string{"hello", "world"}, inc)
	assert.Empty(t, exc)

	inc, exc = parseSearchQuery("a -b c -d")
	assert.Equal(t, []string{"a", "c"}, inc)
	assert.Equal(t, []string{"b", "d"}, exc)

	inc, exc = parseSearchQuery("-only")
	assert.Empty(t, inc)
	assert.Equal(t, []string{"only"}, exc)
}

func TestTodoTxt_MutationsReturnNotFoundErrors(t *testing.T) {
	d, _ := connectTodoTxtWithOtherList(t)

	require.ErrorIs(t, d.UpdateTask("missing", "x"), db.ErrTaskNotFound)
	require.ErrorIs(t, d.DoneTask("missing"), db.ErrTaskNotFound)
	require.ErrorIs(t, d.SetTaskMetadata("missing", "notes", "x"), db.ErrTaskNotFound)
	require.ErrorIs(t, d.SetTaskMetadata("missing", "priority", "A"), db.ErrTaskNotFound)
	require.ErrorIs(t, d.UpdateList("missing", "x"), db.ErrListNotFound)
	require.ErrorIs(t, d.DeleteList("missing"), db.ErrListNotFound)
}

func TestTodoTxt_DeleteDefaultListIsRefused(t *testing.T) {
	d, _ := connectTodoTxtWithOtherList(t)

	require.ErrorIs(t, d.DeleteList("inbox"), db.ErrDefaultListNotDeletable)
}

func TestTodoTxt_SetTaskPropertyPropertyRejectsUnknownType(t *testing.T) {
	d, _ := connectTodoTxtWithOtherList(t)

	require.ErrorIs(t, d.SetTaskPropertyProperty("colour", "work", "k", "v"), db.ErrInvalidPropertyType)
	require.NoError(t, d.SetTaskPropertyProperty("tag", "work", "k", "v"))
}

func TestTodoTxt_SetTaskMetadataOnExistingTask(t *testing.T) {
	d, _ := connectTodoTxtWithOtherList(t)
	id, err := d.CreateTask("task", "inbox", "")
	require.NoError(t, err)

	require.NoError(t, d.SetTaskMetadata(id, "notes", "hello"))
	fields, err := d.GetTaskMetadata(id)
	require.NoError(t, err)
	assert.Equal(t, "hello", fields["notes"])
}

func TestTodoTxt_SetTaskMetadataRejectsPathTraversalInField(t *testing.T) {
	d, _ := connectTodoTxtWithOtherList(t)
	id, err := d.CreateTask("task", "inbox", "")
	require.NoError(t, err)

	for _, field := range []string{"../../escape", "a/b", "", "."} {
		require.ErrorIs(t, d.SetTaskMetadata(id, field, "x"), db.ErrInvalidMetadataField, "field %q", field)
	}
	_, statErr := os.Stat(filepath.Join(d.dir, "..", "escape.txt"))
	assert.True(t, os.IsNotExist(statErr), "no file may be written outside the metadata dir")
}

// Lines added by hand (or by another todo.txt tool) have no id: key. Every
// process and every reload must agree on their ID, otherwise an ID read from
// one is unknown to the next and writes fail with not found.
func TestTodoTxt_TasksWithoutIDGetStableIDsAcrossInstances(t *testing.T) {
	dir, cleanup := setTodotxtDir(t)
	defer cleanup()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "todo.txt"), []byte("hand written task\nhand written task\nanother\n"), 0o600))

	first := &TodoTxt{}
	require.NoError(t, first.Connect())
	second := &TodoTxt{}
	require.NoError(t, second.Connect())

	firstTasks, err := first.GetTasks("inbox")
	require.NoError(t, err)
	secondTasks, err := second.GetTasks("inbox")
	require.NoError(t, err)
	require.Len(t, firstTasks, 3)
	require.Len(t, secondTasks, 3)
	for i := range firstTasks {
		assert.Equal(t, firstTasks[i].ID, secondTasks[i].ID)
	}
	assert.NotEqual(t, firstTasks[0].ID, firstTasks[1].ID, "identical lines must still get distinct IDs")

	require.NoError(t, second.DoneTask(firstTasks[0].ID))
}

func TestTodoTxt_TasksWithoutIDKeepIDsAcrossReload(t *testing.T) {
	dir, cleanup := setTodotxtDir(t)
	defer cleanup()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "todo.txt"), []byte("hand written task\n"), 0o600))

	d := &TodoTxt{}
	require.NoError(t, d.Connect())
	before, err := d.GetTasks("inbox")
	require.NoError(t, err)
	d.doReload()

	require.NoError(t, d.UpdateTask(before[0].ID, "renamed task"))
}

func TestTodoTxt_ReparentTaskMovesTreeUnderNewParent(t *testing.T) {
	d, otherListID := connectTodoTxtWithOtherList(t)

	newParentID, err := d.CreateTask("new parent", otherListID, "")
	require.NoError(t, err)
	taskID, err := d.CreateTask("task", "inbox", "")
	require.NoError(t, err)
	childID, err := d.CreateTask("child", "inbox", taskID)
	require.NoError(t, err)

	require.NoError(t, d.ReparentTask(taskID, newParentID))

	moved, err := d.GetTask(taskID)
	require.NoError(t, err)
	assert.Equal(t, "task", moved.ParentType)
	assert.Equal(t, newParentID, moved.ParentId)
	subs, err := d.GetSubtasks(newParentID)
	require.NoError(t, err)
	require.Len(t, subs, 1)
	assert.Equal(t, taskID, subs[0].ID)
	for _, task := range d.tasks {
		if id := task.Metadata["id"]; id == taskID || id == childID {
			assert.Equal(t, otherListID, task.Metadata["listid"], "%s should join the new parent's list", id)
		}
	}
	roots, err := d.GetTasks("inbox")
	require.NoError(t, err)
	assert.Empty(t, roots)
}

func TestTodoTxt_ReparentTaskErrors(t *testing.T) {
	d, _ := connectTodoTxtWithOtherList(t)
	taskID, err := d.CreateTask("task", "inbox", "")
	require.NoError(t, err)
	childID, err := d.CreateTask("child", "inbox", taskID)
	require.NoError(t, err)

	assert.ErrorIs(t, d.ReparentTask(taskID, taskID), db.ErrInvalidMove)
	assert.ErrorIs(t, d.ReparentTask(taskID, childID), db.ErrInvalidMove)
	assert.ErrorIs(t, d.ReparentTask("missing", taskID), db.ErrTaskNotFound)
	assert.ErrorIs(t, d.ReparentTask(childID, "missing"), db.ErrTaskNotFound)
}

func TestTodoTxt_DeleteTaskRemovesTreeAndMetadata(t *testing.T) {
	dir, cleanup := setTodotxtDir(t)
	defer cleanup()
	d := &TodoTxt{}
	require.NoError(t, d.Connect())
	keepID, err := d.CreateTask("keep", "inbox", "")
	require.NoError(t, err)
	taskID, err := d.CreateTask("delete me", "inbox", "")
	require.NoError(t, err)
	childID, err := d.CreateTask("child", "inbox", taskID)
	require.NoError(t, err)
	require.NoError(t, d.DoneTask(childID))
	require.NoError(t, d.SetTaskMetadata(taskID, "notes", "gone soon"))

	require.NoError(t, d.DeleteTask(taskID))

	for _, id := range []string{taskID, childID} {
		task, err := d.GetTask(id)
		require.NoError(t, err)
		assert.Nil(t, task, "%s should be deleted", id)
	}
	kept, err := d.GetTask(keepID)
	require.NoError(t, err)
	assert.NotNil(t, kept)
	assert.NoDirExists(t, filepath.Join(dir, "metadata", taskID))

	reloaded := &TodoTxt{}
	require.NoError(t, reloaded.Connect())
	assert.Len(t, reloaded.tasks, 1, "deletion should be persisted to todo.txt and done.txt")
	assert.ErrorIs(t, d.DeleteTask(taskID), db.ErrTaskNotFound)
}

func TestTodoTxt_GetCompletedTasksFiltersBySinceAndList(t *testing.T) {
	dir, cleanup := setTodotxtDir(t)
	defer cleanup()
	done := "x 2026-11-01 old task id:old listid:inbox\nx 2026-11-20 recent task id:recent listid:inbox\nx 2026-11-21 other list id:other listid:work\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "done.txt"), []byte(done), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "todo.txt"), []byte("open task id:open\n"), 0o600))
	d := &TodoTxt{}
	require.NoError(t, d.Connect())

	all, err := d.GetCompletedTasks(time.Time{}, "")
	require.NoError(t, err)
	assert.Equal(t, []string{"other", "recent", "old"}, taskIDs(all), "most recently completed first, open tasks excluded")
	assert.True(t, all[0].Completed)
	assert.Equal(t, "2026-11-21", all[0].CompletedAt)

	since, err := d.GetCompletedTasks(time.Date(2026, 11, 20, 15, 0, 0, 0, time.UTC), "")
	require.NoError(t, err)
	assert.Equal(t, []string{"other", "recent"}, taskIDs(since), "since compares by day")

	inList, err := d.GetCompletedTasks(time.Time{}, "inbox")
	require.NoError(t, err)
	assert.Equal(t, []string{"recent", "old"}, taskIDs(inList))
}

func TestTodoTxt_TaskIncludesNotesAndOpenState(t *testing.T) {
	_, cleanup := setTodotxtDir(t)
	defer cleanup()
	d := &TodoTxt{}
	require.NoError(t, d.Connect())
	id, err := d.CreateTask("with notes", "inbox", "")
	require.NoError(t, err)
	require.NoError(t, d.SetTaskMetadata(id, "notes", "line one\nline two"))

	task, err := d.GetTask(id)
	require.NoError(t, err)
	assert.Equal(t, "line one\nline two", task.Notes)
	assert.False(t, task.Completed)
	assert.Empty(t, task.CompletedAt)
}

func taskIDs(tasks []db.DBTask) []string {
	ids := make([]string, 0, len(tasks))
	for _, t := range tasks {
		ids = append(ids, t.ID)
	}
	return ids
}
