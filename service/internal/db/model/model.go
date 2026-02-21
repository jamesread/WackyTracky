package db

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
	CountSubitems int32
	Tags          []string
	Contexts      []string
	WaitUntil     string
	Priority      string
	DueDate       string
}

// SavedSearch is used by backends that support saved searches (e.g. todotxt).
type SavedSearch struct {
	ID    string
	Title string
	Query string
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
