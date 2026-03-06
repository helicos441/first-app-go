package main

import (
	"database/sql"
	"encoding/json"
	"first-app-go/internal/data"
	"net/http"
	"net/http/httptest"
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

	return &App{Stores: data.NewStores(db)}
}

func TestListBooksHandler(t *testing.T) {
	app := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/books", nil)

	rr := httptest.NewRecorder()

	app.listBooksHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("want status code %d; got %d", http.StatusOK, rr.Code)
	}

	var resp bookResponse

	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	booksCount := len(resp.Books)
	if booksCount != 2 {
		t.Errorf("want books count of 2; got %d", booksCount)
	}
}
