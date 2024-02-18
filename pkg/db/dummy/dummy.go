package dummy

import (
	log "github.com/sirupsen/logrus"
	dbconn "github.com/wacky-tracky/wacky-tracky-server/pkg/db"
	"github.com/google/uuid"
)

type Dummy struct {
	dbconn.DB

	tasks []dbconn.DBTask
	tags  []dbconn.DBTag
	lists []dbconn.DBList
}

func (db *Dummy) Connect() error {
	db.setup()

	return nil
}

func (db *Dummy) setup() {
	db.tasks = []dbconn.DBTask{
		dbconn.DBTask{
			ID:      "1",
			Content: "First List Task One",
			ParentType: "list",
			ParentId: "1",
		},
		dbconn.DBTask{
			ID:      "2",
			Content: "Second List Task One",
			ParentType: "list",
			ParentId: "2", 
		},
		dbconn.DBTask{
			ID: "3",
			Content: "Second List Task Two",
			ParentType: "list",
			ParentId: "2",
		},
	}

	db.tags = []dbconn.DBTag{
		dbconn.DBTag{
			ID:    "1",
			Title: "First tag",
		},
	}

	db.lists = []dbconn.DBList{
		dbconn.DBList{
			ID:         "1",
			Title:      "First list",
			CountTasks: 1,
		},
		dbconn.DBList{
			ID:         "2",
			Title:      "Second list",
			CountTasks: 2,
		},
	}
}

func (db *Dummy) Print() {
	log.Infof("foo %+v", db)
}

func (db *Dummy) GetTask(taskId string) (*dbconn.DBTask, error) {
	for _, task := range db.tasks {
		if task.ID == taskId {
			return &task, nil
		}
	}

	return nil, nil
}

func (db *Dummy) GetTasks(listId string) ([]dbconn.DBTask, error) {
	return db.tasks, nil
}

func (db *Dummy) GetTags() ([]dbconn.DBTag, error) {
	return db.tags, nil
}

func (db *Dummy) GetLists() ([]dbconn.DBList, error) {
	return db.lists, nil
}

func (db *Dummy) CreateList(content string) error {
	db.lists = append(db.lists, dbconn.DBList{
		ID: uuid.New().String(),
		Title: content,
	})

	return nil
}

func (db *Dummy) CreateTask(content string) (string, error) {
	id := uuid.New().String()

	db.tasks = append(db.tasks, dbconn.DBTask{
		ID:      id,
		Content: content,
	})

	return id, nil
}
