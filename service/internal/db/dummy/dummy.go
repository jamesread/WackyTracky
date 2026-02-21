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
	var out []dbmdl.DBTask
	for _, t := range db.tasks {
		if t.ParentType == "list" && t.ParentId == listId {
			out = append(out, t)
		}
	}
	return out, nil
}

func (db *Dummy) GetSubtasks(itemId string) ([]dbmdl.DBTask, error) {
	var out []dbmdl.DBTask
	for _, t := range db.tasks {
		if t.ParentType == "task" && t.ParentId == itemId {
			out = append(out, t)
		}
	}
	return out, nil
}

func (db *Dummy) GetTags() ([]dbmdl.DBTag, error) {
	return db.tags, nil
}

func (db *Dummy) GetLists() ([]dbmdl.DBList, error) {
	return db.lists, nil
}

func (db *Dummy) CreateList(title string) error {
	db.lists = append(db.lists, dbmdl.DBList{
		ID:    uuid.New().String(),
		Title: title,
	})
	return nil
}

func (db *Dummy) CreateTask(content string, listId string, parentTaskId string) (string, error) {
	id := uuid.New().String()
	parentType := "list"
	parentId := listId
	if listId == "" {
		parentId = "1"
	}
	if parentTaskId != "" {
		parentType = "task"
		parentId = parentTaskId
	}
	db.tasks = append(db.tasks, dbmdl.DBTask{
		ID:         id,
		Content:    content,
		ParentId:   parentId,
		ParentType: parentType,
	})
	return id, nil
}

func (db *Dummy) UpdateList(id string, title string) error {
	for i := range db.lists {
		if db.lists[i].ID == id {
			db.lists[i].Title = title
			return nil
		}
	}
	return nil
}

func (db *Dummy) DeleteList(id string) error {
	for i, l := range db.lists {
		if l.ID == id {
			db.lists = append(db.lists[:i], db.lists[i+1:]...)
			return nil
		}
	}
	return nil
}
