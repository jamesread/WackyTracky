package db

type DBTag struct {
	ID uint64
	Title string
}

type DBList struct {
	ID uint64
	Title string
}

type DBItem struct {
	ID uint64
	Content string
}


type DB interface {
	GetTags() ([]DBTag, error)
	GetItems(listId uint64) ([]DBItem, error)
	GetLists() ([]DBList, error)
}

