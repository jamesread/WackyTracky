package neo4j

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
	db "github.com/wacky-tracky/wacky-tracky-server/pkg/db"

	log "github.com/sirupsen/logrus"
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

	log.Infof("Connecting to neo4j")

	uri := "bolt://neo4j:7687"
	username := "neo4j"
	password := "password"

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

	log.Infof("Connected")

	return nil
}

func greeting() (string, error) {
	greeting, err := session.WriteTransaction(func(transaction neo4j.Transaction) (any, error) {
		log.Infof("Write TX")
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
			log.Infof("Wrote TX")
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

func readTx(cql string) []dbtype.Node {
	return readTxParams(cql, map[string]any{})
}

func readTxParams(cql string, params map[string]any) []dbtype.Node {
	var ret []dbtype.Node


	_, err := session.ReadTransaction(func(tx neo4j.Transaction) (any, error) {
		res, err := tx.Run(cql, params)

		if err != nil {
			return nil, err
		} else {
			for res.Next() {
				ret = append(ret, res.Record().Values[0].(dbtype.Node))
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
			ID:    uint64(tag.Id),
			Title: tag.Props["title"].(string),
		}

		ret = append(ret, dbtag)
	}
	*/

	log.Infof("tags: %+v", ret)

	return ret, nil
}

func (self Neo4jDB) GetLists() ([]db.DBList, error) {
	log.Infof("getting lists")

	var ret []db.DBList

	cql := "MATCH (n:List) RETURN n"

	for _, lst := range readTx(cql) {
		ret = append(ret, db.DBList{
			ID:    uint64(lst.Id),
			Title: lst.Props["title"].(string),
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

	for _, subitem := range readTxParams(cql, params) {
		ret = append(ret, db.DBTask{
			ID: uint64(subitem.Id),
		})
	}

	return ret
}

func (db Neo4jDB) GetTasks(listId uint64) ([]db.DBTask, error) {
	cql := "MATCH (l:List)-[]->(i:Item) OPTIONAL MATCH (i)-->(tv:TagValue) OPTIONAL MATCH (i)-->(subItem:Item) OPTIONAL MATCH (externalItem:ExternalItem) WHERE i = externalItem WITH l, i, count(tv) AS countTagValues, count(subItem) AS countItems, externalItem WHERE id(l) = $listId WITH i, countTagValues, countItems, externalItem RETURN i, countTagValues, countItems, externalItem ORDER BY id(i)"

	log.Infof("getting items from list %v", listId)

	ret, err := session.ReadTransaction(func(tx neo4j.Transaction) (any, error) {
		res, err := tx.Run(cql, map[string]any{
			"listId": listId,
		})

		var items []interface{}

		if err != nil {
			log.Errorf("%v", err)
			return nil, err
		} else {
			for res.Next() {
				x := res.Record().Values[0].(dbtype.Node)
				log.Infof("Items %+v", x)
				GetSubItems(x.Id)
				items = append(items, res.Record().Values[0])
			}
		}

		return items, nil
	})

	log.Infof("%v %v", ret, err)

	return nil, nil
}
