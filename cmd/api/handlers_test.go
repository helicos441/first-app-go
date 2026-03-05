package main

import (
	"database/sql"
	"first-app-go/internal/data"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestApp(t *testing.T) *App {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	if err := data.Migrate(db); err != nil {
		t.Fatal(err)
	}

	if err := data.SeedIfEmpty(db); err != nil {
		t.Fatal(err)
	}

	return &App{DB: db}
}
