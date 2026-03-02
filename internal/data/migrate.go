package data

import "database/sql"

func Migrate(db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS books (
  id     INTEGER PRIMARY KEY AUTOINCREMENT,
  title  TEXT NOT NULL,
  author TEXT,
  year   INTEGER
);`
	_, err := db.Exec(ddl)
	return err
}
