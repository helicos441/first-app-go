package data

import (
	"context"
	"database/sql"
	"time"

	_ "modernc.org/sqlite" // registers the sqlite driver
)

const dsn = "file:books.db?_pragma=busy_timeout(5000)"

func OpenSQLite() (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		// If ping fails, close the pool before returning the error.
		_ = db.Close()
		return nil, err
	}

	return db, nil
}
