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

func SeedIfEmpty(db *sql.DB) error {
	var count int

	err := db.QueryRow(`SELECT COUNT(*) FROM books`).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	_, err = db.Exec(`
INSERT INTO books (title, author, year) VALUES
  ('The Go Programming Language', 'Alan Donovan', 2015),
  ('Designing Data-Intensive Applications', 'Martin Kleppmann', 2017)`)

	return err
}
