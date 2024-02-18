package neo4j

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	db "github.com/wacky-tracky/wacky-tracky-server/pkg/db"

	"github.com/google/uuid"

	log "github.com/sirupsen/logrus"

	"fmt"
	"strconv"

	. "github.com/wacky-tracky/wacky-tracky-server/pkg/runtimeconfig"
)

var (
	driver  neo4j.Driver
	session neo4j.Session
)

type Neo4jDB struct {
	db.DB
}

func (self Neo4jDB) Connect() error {
	if driver != nil {
		return nil // FIXME driver might be initialized but stale/disconnected
	}

	uri := fmt.Sprintf("bolt://%v:%v", RuntimeConfig.Database.Hostname, RuntimeConfig.Database.Port)

	log.WithFields(log.Fields{
		"uri": uri,
	}).Infof("Connecting to neo4j")

	username := RuntimeConfig.Database.Username
	password := RuntimeConfig.Database.Password

	var err error

	driver, err = neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))

	if err != nil {
		log.Errorf("%v", err)
		return err
	}

	log.Infof("Driver created")

	//defer driver.Close()

	session = driver.NewSession(neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})

	//session.config.ConnectionAcquisutionTimeout = 5 * time.Second;
	//defer session.Close()

	log.Infof("Session created")

	greeting()

	log.Infof("Connected to neo4j")

	return nil
}

func greeting() (string, error) {
	greeting, err := session.WriteTransaction(func(transaction neo4j.Transaction) (any, error) {
		cql := "CREATE (a:Greeting) SET a.message = $message RETURN a.message + ', from node ' + id(a)"
		cql = "MATCH (n:List) RETURN (n.title)"
		result, err := transaction.Run(
			cql,
			map[string]any{"message": "hello, world"},
		)

		if err != nil {
			log.Errorf("err %v", err)
			return nil, err
		} else {
			transaction.Commit()
		}

		if result.Next() {
			ret := result.Record().Values[0]

			result.Consume()

			return ret, nil

		}
		log.Infof("Greet err")

		return nil, result.Err()
	})
	if err != nil {
		log.Errorf("%v", err)
		return "", err
	}

	log.Infof("Greeting: %v", greeting)

	return greeting.(string), nil
}

func readTx(cql string) []*neo4j.Record {
	return readTxParams(cql, map[string]any{})
}

func readTxParams(cql string, params map[string]any) []*neo4j.Record {
	var ret []*neo4j.Record

	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (any, error) {
		res, err := tx.Run(cql, params)

		if err != nil {
			return nil, err
		} else {
			for res.Next() {
				ret = append(ret, res.Record())
			}

			return nil, nil
		}
	})

	if err != nil {
		log.Errorf("readTx %v", err)
	}

	return ret
}

func (self Neo4jDB) GetTags() ([]db.DBTag, error) {
	var ret []db.DBTag

	//cql := "MATCH (t:Tag) RETURN t"

	//readTx(cql);
	/*
		for _, tag := range readTx(cql) {
			log.Infof("tag props %+v", tag.Props)

			dbtag := db.DBTag{
				ID:    int32(tag.Id),
				Title: tag.Props["title"].(string),
			}

			ret = append(ret, dbtag)
		}
	*/

	log.Infof("tags: %+v", ret)

	return ret, nil
}

func (db Neo4jDB) Print() {}

func (self Neo4jDB) GetLists() ([]db.DBList, error) {
	log.Infof("getting lists")

	var ret []db.DBList

	cql := "MATCH (l:List) OPTIONAL MATCH (l)-->(i:Item)  RETURN l, count(i) as countItems"

	for _, row := range readTx(cql) {
		lst := row.Values[0].(dbtype.Node)

		ret = append(ret, db.DBList{
			ID:         fmt.Sprintf("%v", lst.Id),
			Title:      lst.Props["title"].(string),
			CountTasks: int32(row.Values[1].(int64)),
		})
	}

	log.Infof("got lists: %+v", ret)

	return ret, nil
}

func GetSubItems(itemId int64) []db.DBTask {
	cql := "MATCH (p:Item)-->(i:Item) WHERE id(p) = $parentItemId RETURN i"

	params := map[string]any{
		"parentItemId": itemId,
	}

	var ret []db.DBTask

	for _, row := range readTxParams(cql, params) {
		item := row.Values[0].(dbtype.Node)

		ret = append(ret, db.DBTask{
			ID: fmt.Sprintf("%v", item.Id),
		})
	}

	return ret
}

func (api Neo4jDB) GetTask(listId string) (*db.DBTask, error) {
	return nil, nil
}

func (api Neo4jDB) GetTasks(listId string) ([]db.DBTask, error) {
	var ret []db.DBTask

	cql := "MATCH (l:List)-[]->(i:Item) OPTIONAL MATCH (i)-->(tv:TagValue) OPTIONAL MATCH (i)-->(subItem:Item) OPTIONAL MATCH (externalItem:ExternalItem) WHERE i = externalItem WITH l, i, count(tv) AS countTagValues, count(subItem) AS countItems, externalItem WHERE id(l) = $listId WITH i, l, countTagValues, countItems, externalItem RETURN i, l, countTagValues, countItems, externalItem ORDER BY id(i)"

	listIdNum, _ := strconv.Atoi(listId)

	params := map[string]any{
		"listId": listIdNum,
	}

	for _, row := range readTxParams(cql, params) {
		task := row.Values[0].(dbtype.Node)
		list := row.Values[1].(dbtype.Node)
		countSubitems := int32(row.Values[3].(int64))

		//GetSubItems(x.Id)
		ret = append(ret, db.DBTask{
			ID:            fmt.Sprintf("%v", task.Id),
			Content:       fmt.Sprintf("%v", task.Props["content"]),
			ParentId:      fmt.Sprintf("%v", list.Id),
			ParentType:    "list",
			CountSubitems: countSubitems,
		})
	}

	log.WithFields(log.Fields{
		"id":  listId,
		"len": len(ret),
	}).Infof("Got items from list")

	return ret, nil
}

func (api Neo4jDB) CreateTask(title string) (string, error) {
	id := uuid.New().String()

	return id, nil
}
