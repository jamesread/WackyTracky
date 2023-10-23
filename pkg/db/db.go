package db

type DBTag struct {
	ID    uint64
	Title string
}

type DBList struct {
	ID         uint64
	Title      string
	CountTasks uint64
}

type DBTask struct {
	ID         uint64
	Content    string
	ParentId   uint64
	ParentType string
}

type DB interface {
	GetTags() ([]DBTag, error)
	GetTasks(listId uint64) ([]DBTask, error)
	GetLists() ([]DBList, error)
	CreateTask(content string) error
}
