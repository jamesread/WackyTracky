package db

import (
	"errors"
	"time"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrListNotFound = errors.New("list not found")
	// ErrDefaultListNotDeletable is returned when deleting the default (inbox) list,
	// which orphaned tasks fall back to.
	ErrDefaultListNotDeletable = errors.New("the default list cannot be deleted")
	ErrInvalidMetadataField    = errors.New("metadata field names may only contain letters, digits, '-' and '_'")
	ErrInvalidPropertyType     = errors.New("property type must be \"tag\" or \"context\"")
	ErrInvalidMove             = errors.New("a task cannot be moved under itself or one of its own subtasks")
)

type DBTag struct {
	ID    string
	Title string
}

type DBList struct {
	ID         string
	Title      string
	CountTasks int32
}

type DBTask struct {
	ID            string
	Content       string
	ParentId      string
	ParentType    string
	WaitUntil     string
	Priority      string
	DueDate       string
	Notes         string
	Completed     bool
	CompletedAt   string
	Tags          []string
	Contexts      []string
	CountSubitems int32
}

// SavedSearch is used by backends that support saved searches (e.g. todotxt).
// JSON keys are id, name, query to match the API; when reading, "Title" is also accepted for backward compatibility.
type SavedSearch struct {
	ID    string `json:"id"`
	Title string `json:"name"`
	Query string `json:"query"`
}

type DB interface {
	Connect() error
	Print()
	GetTags() ([]DBTag, error)
	GetTask(id string) (*DBTask, error)
	GetTasks(listId string) ([]DBTask, error)
	GetSubtasks(itemId string) ([]DBTask, error)
	GetLists() ([]DBList, error)
	CreateTask(content string, listId string, parentTaskId string) (string, error)
	CreateList(title string) error
	UpdateList(id string, title string) error
	DeleteList(id string) error
}

// Updatable is optionally implemented by backends that support updating and completing tasks.
type Updatable interface {
	UpdateTask(id string, content string) error
	DoneTask(id string) error
}

// TaskMover is optionally implemented by backends that support moving a task to another list.
type TaskMover interface {
	MoveTask(taskId string, targetListId string) error
}

// TaskReparenter is optionally implemented by backends that can make a task a subtask of another.
type TaskReparenter interface {
	ReparentTask(taskId string, parentTaskId string) error
}

// TaskDeleter is optionally implemented by backends that can permanently delete a task and its subtasks.
type TaskDeleter interface {
	DeleteTask(id string) error
}

// CompletedTaskLister is optionally implemented by backends that keep completed tasks.
// A zero since lists all of them; an empty listId lists every list.
type CompletedTaskLister interface {
	GetCompletedTasks(since time.Time, listId string) ([]DBTask, error)
}

// Searchable is optionally implemented by backends that support task search.
type Searchable interface {
	SearchTasks(query string) ([]DBTask, error)
}

// SavedSearchesStore is optionally implemented by backends that persist saved searches.
type SavedSearchesStore interface {
	GetSavedSearches() ([]SavedSearch, error)
	SetSavedSearches(searches []SavedSearch) error
}

// TaskMetadataStore is optionally implemented by backends that store per-task metadata (notes, wait, due, priority).
type TaskMetadataStore interface {
	GetTaskMetadata(taskId string) (map[string]string, error)
	SetTaskMetadata(taskId string, field string, value string) error
}

// TaskPropertyPropertiesStore is optionally implemented by backends that store key/value props on task properties (tags, contexts).
// PropertyType is "tag" or "context"; propertyName is the tag name (e.g. "work") or context name (e.g. "home").
type TaskPropertyPropertiesStore interface {
	GetTaskPropertyProperties() (tagProperties map[string]map[string]string, contextProperties map[string]map[string]string, err error)
	SetTaskPropertyProperty(propertyType, propertyName, key, value string) error
}
