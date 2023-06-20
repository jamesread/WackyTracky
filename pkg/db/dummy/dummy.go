package dummy


import (
	dbconn "github.com/wacky-tracky/wacky-tracky-server/pkg/db"
)

type Dummy struct {
	dbconn.DB
}

func (db Dummy) GetItems(listId uint64) ([]dbconn.DBItem, error) {
	var ret []dbconn.DBItem

	ret = append(ret, dbconn.DBItem{
		ID: 1,
		Content: "one",
	})

	ret = append(ret, dbconn.DBItem{
		ID: 2,
		Content: "two",
	})

	return ret, nil
}

func (db Dummy) GetTags() ([]dbconn.DBTag, error) {
	var ret []dbconn.DBTag


	ret = append(ret, dbconn.DBTag {
		ID: 1,
		Title: "One",
	})

	return ret, nil
}

func (db Dummy) GetLists() ([]dbconn.DBList, error) {
	var ret []dbconn.DBList

	ret = append(ret, dbconn.DBList{
		ID: 1,
	})

	ret = append(ret, dbconn.DBList{
		ID: 2,
	})



	return ret, nil
}
