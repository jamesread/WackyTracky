package db

type DBTag struct {
	ID    int32
	Title string
}

type DBList struct {
	ID         int32
	Title      string
	CountTasks int32
}

type DBTask struct {
	ID            int32
	Content       string
	ParentId      int32
	ParentType    string
	CountSubitems int32
}

type DB interface {
	Connect() error
	Print()
	GetTags() ([]DBTag, error)
	GetTasks(listId int32) ([]DBTask, error)
	GetLists() ([]DBList, error)
	CreateTask(content string) error
}
