package yamlfiles

import (
	db "github.com/wacky-tracky/wacky-tracky-server/pkg/db"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"github.com/google/uuid"
)

type YamlFileDriver struct {}

var (
	tasks []db.DBTask
	tags  []db.DBTag
	lists []db.DBList

	filenameTasks = "tasks.yaml"
	filenameTags = "tags.yaml"
	filenameLists = "lists.yaml"
)

func save(filename string, structure interface{}) {
	log.WithFields(log.Fields{
		"filename": filename,
	}).Infof("YamlFiles save")

	data, err := yaml.Marshal(structure)

	if err != nil {
		log.Fatalf("%v", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)

	if err != nil {
		log.Fatalf("%v", err)
	}
}

func load(filename string, structure interface{}) {
	yfile, err := ioutil.ReadFile(filename)

	if err != nil {
		log.Fatalf("%v", err)
	}

	err = yaml.Unmarshal(yfile, &structure)

	if err != nil {
		log.Fatalf("%v", err)
	}
}

func (drv *YamlFileDriver) Connect() error {
	load(filenameLists, lists)
	load(filenameTasks, tasks)
	load(filenameTags, tags)

	return nil
}

func (drv *YamlFileDriver) Print() {
}

func (drv *YamlFileDriver) CreateTask(content string) (string, error) {
	id := uuid.New().String()

	task := db.DBTask {
		Content: content,
		ID: id,
	}

	tasks = append(tasks, task)

	save(filenameTasks, tasks)

	return id, nil
}

func (drv *YamlFileDriver) GetLists() ([]db.DBList, error) {
	return lists, nil
}

func (drv *YamlFileDriver) GetTags() ([]db.DBTag, error) {
	return tags, nil
}

func (drv *YamlFileDriver) GetTask(taskId string) (*db.DBTask, error) {
	task := &db.DBTask {

	}

	return task, nil
}

func (drv *YamlFileDriver) GetTasks(listId string) ([]db.DBTask, error) {

	return tasks, nil
}

func (drv *YamlFileDriver) CreateList(title string) error {
	list := db.DBList {
		ID:      uuid.New().String(),
		Title: title,
	}

	lists = append(lists, list)

	save(filenameLists, lists)

	return nil
}
