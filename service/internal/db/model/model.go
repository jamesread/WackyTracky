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
}

type DB interface {
	Connect() error
	Print()
	GetTags() ([]DBTag, error)
	GetTask(id string) (*DBTask, error)
	GetTasks(listId string) ([]DBTask, error)
	GetSubtasks(itemId string) ([]DBTask, error)
	GetLists() ([]DBList, error)
	CreateTask(content string) (string, error)
	CreateList(content string) error
}
