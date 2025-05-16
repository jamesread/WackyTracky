package dummy

import (
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	dbmdl "github.com/wacky-tracky/wacky-tracky-server/internal/db/model"
)

type Dummy struct {
	dbmdl.DB

	tasks []dbmdl.DBTask
	tags  []dbmdl.DBTag
	lists []dbmdl.DBList
}

func (db *Dummy) Connect() error {
	db.setup()

	return nil
}

func (db *Dummy) setup() {
	db.tasks = []dbmdl.DBTask{
		dbmdl.DBTask{
			ID:         "1",
			Content:    "First List Task One",
			ParentType: "list",
			ParentId:   "1",
		},
		dbmdl.DBTask{
			ID:         "2",
			Content:    "Second List Task One",
			ParentType: "list",
			ParentId:   "2",
		},
		dbmdl.DBTask{
			ID:         "3",
			Content:    "Second List Task Two",
			ParentType: "list",
			ParentId:   "2",
		},
	}

	db.tags = []dbmdl.DBTag{
		dbmdl.DBTag{
			ID:    "1",
			Title: "First tag",
		},
	}

	db.lists = []dbmdl.DBList{
		dbmdl.DBList{
			ID:         "1",
			Title:      "First list",
			CountTasks: 1,
		},
		dbmdl.DBList{
			ID:         "2",
			Title:      "Second list",
			CountTasks: 2,
		},
	}
}

func (db *Dummy) Print() {
	log.Infof("dbdummy %+v", db)
}

func (db *Dummy) GetTask(taskId string) (*dbmdl.DBTask, error) {
	for _, task := range db.tasks {
		if task.ID == taskId {
			return &task, nil
		}
	}

	return nil, nil
}

func (db *Dummy) GetTasks(listId string) ([]dbmdl.DBTask, error) {
	return db.tasks, nil
}

func (db *Dummy) GetTags() ([]dbmdl.DBTag, error) {
	return db.tags, nil
}

func (db *Dummy) GetLists() ([]dbmdl.DBList, error) {
	return db.lists, nil
}

func (db *Dummy) CreateList(content string) error {
	db.lists = append(db.lists, dbmdl.DBList{
		ID:    uuid.New().String(),
		Title: content,
	})

	return nil
}

func (db *Dummy) CreateTask(content string) (string, error) {
	id := uuid.New().String()

	db.tasks = append(db.tasks, dbmdl.DBTask{
		ID:      id,
		Content: content,
	})

	return id, nil
}
