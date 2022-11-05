package neo4j

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	log "github.com/sirupsen/logrus"
)

var (
	driver neo4j.Driver
	session neo4j.Session
)

func connect() (string, error) {
	log.Infof("Connecting to neo4j")

	uri := "bolt://neo4j:7687"
	username := "neo4j"
	password := "password"

	var err error;

	driver, err = neo4j.NewDriver(uri, neo4j.BasicAuth(username, password, ""))

	if err != nil {
		log.Errorf("%v", err)
		return "", err
	}

	log.Infof("Driver created")

	//defer driver.Close()

	session = driver.NewSession(neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})

	//session.config.ConnectionAcquisutionTimeout = 5 * time.Second;
	//defer session.Close()

	log.Infof("Session created")

	greeting()
	greeting()

	log.Infof("Connected");

	return "", nil
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

func GetLists() ([]string, error) {
	connect()

	cql := "MATCH (n:List) RETURN n.title, id(n)"

	log.Infof("getting lists")
	ret, err := session.ReadTransaction(func(tx neo4j.Transaction) (any, error) {
		res, err := tx.Run(cql, nil)

		var list []string

		if err != nil {
			log.Errorf("%v", err)
			return nil, err
		} else {
			for res.Next() {
				list = append(list, res.Record().Values[0].(string))

				log.Infof("%v %v", res.Record().Values[0], res.Record().Values[1])
			}
		}

		return list, nil

	})

	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}
	log.Infof("got lists: %+v", ret.([]string))

	return ret.([]string), nil
}

func GetItems(listId int) {
	connect()

	cql := "MATCH (l:List)-[]->(i:Item) OPTIONAL MATCH (i)-->(tv:TagValue) OPTIONAL MATCH (i)-->(subItem:Item) OPTIONAL MATCH (externalItem:ExternalItem) WHERE i = externalItem WITH l, i, count(tv) AS countTagValues, count(subItem) AS countItems, externalItem WHERE id(l) = $listId WITH i, countTagValues, countItems, externalItem RETURN i, countTagValues, countItems, externalItem ORDER BY id(i)"

	log.Infof("getting items from list")
	ret, err := session.ReadTransaction(func(tx neo4j.Transaction) (any, error) {
		res, err := tx.Run(cql, map[string]any {
			"listId": listId,
		})

		var list []string

		if err != nil {
			log.Errorf("%v", err)
			return nil, err
		} else {
			for res.Next() {
				list = append(list, res.Record().Values[0].(string))
			}
		}

		return list, nil
	})

	log.Infof("%v %v", ret, err)
}
