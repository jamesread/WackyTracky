package dummy

import (
	dbconn "github.com/wacky-tracky/wacky-tracky-server/pkg/db"
)

type Dummy struct {
	dbconn.DB

	tasks []dbconn.DBTask
	tags  []dbconn.DBTag
	lists []dbconn.DBList
}

func (db Dummy) Connect() error {
	db.setup()

	return nil
}

func (db Dummy) setup() {
	db.tasks = []dbconn.DBTask{
		dbconn.DBTask{
			ID:      1,
			Content: "First Task",
		},
		dbconn.DBTask{
			ID:      2,
			Content: "Second Task",
		},
	}

	db.tags = []dbconn.DBTag{
		dbconn.DBTag{
			ID:    1,
			Title: "First tag",
		},
	}

	db.lists = []dbconn.DBList{
		dbconn.DBList{
			ID:         1,
			Title:      "First list",
			CountTasks: 9,
		},
		dbconn.DBList{
			ID:         2,
			Title:      "Second list",
			CountTasks: 4,
		},
	}
}

func (db Dummy) GetTasks(listId uint64) ([]dbconn.DBTask, error) {
	return db.tasks, nil
}

func (db Dummy) GetTags() ([]dbconn.DBTag, error) {
	return db.tags, nil
}

func (db Dummy) GetLists() ([]dbconn.DBList, error) {
	return db.lists, nil
}

func (db Dummy) CreateTask(content string) error {
	db.tasks = append(db.tasks, dbconn.DBTask{
		ID:      uint64(len(db.tasks) + 1),
		Content: content,
	})

	return nil
}
