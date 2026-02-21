package todotxt

import (
	"os"
	"path/filepath"
	"testing"

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

	if err := d.UpdateTask(id, "updated content"); err != nil {
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
