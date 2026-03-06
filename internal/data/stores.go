package data

import "database/sql"

type Stores struct {
	Books BookStore
}

func NewStores(db *sql.DB) Stores {
	return Stores{
		Books: BookStore{DB: db},
	}
}
