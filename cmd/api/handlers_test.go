package main

import (
	"database/sql"
	"encoding/json"
	"first-app-go/internal/data"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestShowBookHandler(t *testing.T) {
	app := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/books/1", http.NoBody)

	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK { // 200
		t.Errorf("want status code %d; got %d", http.StatusOK, rr.Code)
	}

	var book data.Book

	if err := json.NewDecoder(rr.Body).Decode(&book); err != nil {
		t.Fatal(err)
	}

	expected := data.Book{
		ID:     1,
		Title:  "The Go Programming Language",
		Author: "Alan Donovan",
		Year:   2015,
	}

	if book != expected {
		t.Errorf("want %#v; got %#v", expected, book)
	}
}

func TestCreateBookHandler_ValidInput(t *testing.T) {
	app := setupTestApp(t)

	body := strings.NewReader(`{
		"title":"Testing Go",
		"author":"Gary Clarke",
		"year":2030
	}`)

	req := httptest.NewRequest(http.MethodPost, "/books", body)

	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("want status code %d; got %d", http.StatusCreated, rr.Code)
	}

	var book data.Book

	if err := json.NewDecoder(rr.Body).Decode(&book); err != nil {
		t.Fatal(err)
	}

	if book.ID < 1 {
		t.Errorf("expected book to have a positive value ID")
	}
	if book.Title != "Testing Go" {
		t.Errorf("expected title to be 'Testing Go'; got %q", book.Title)
	}
	if book.Author != "Gary Clarke" {
		t.Errorf("expected author to be 'Gary Clarke'; got %q", book.Author)
	}
	if book.Year != 2030 {
		t.Errorf("expected year to be 2030; got %d", book.Year)
	}

	stored, err := app.Stores.Books.Get(book.ID)
	if err != nil {
		t.Fatalf("failed to fetch book from DB: %v", err)
	}

	if *stored != book {
		t.Errorf("book in DB does not match response. got: %#v", stored)
	}
}

func TestCreateBookHandler_InvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		wantCode int
		wantKeys []string // expected keys in the "errors" object of the response
	}{
		{
			name:     "missing all fields",
			payload:  `{}`,
			wantCode: http.StatusUnprocessableEntity,
			wantKeys: []string{"title", "author", "year"},
		},
		{
			name:     "missing title",
			payload:  `{"author": "Gary", "year": 2023}`,
			wantCode: http.StatusUnprocessableEntity,
			wantKeys: []string{"title"},
		},
		{
			name:     "missing author",
			payload:  `{"title": "Testing Go", "year": 2023}`,
			wantCode: http.StatusUnprocessableEntity,
			wantKeys: []string{"author"},
		},
		{
			name:     "invalid year (zero)",
			payload:  `{"title": "Testing Go", "author": "Gary", "year": 0}`,
			wantCode: http.StatusUnprocessableEntity,
			wantKeys: []string{"year"},
		},
		{
			name:     "invalid JSON format",
			payload:  `{`,
			wantCode: http.StatusBadRequest,
			wantKeys: nil, // No "errors" object expected — it's a decoding error
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := setupTestApp(t)

			body := strings.NewReader(tc.payload)

			req := httptest.NewRequest(http.MethodPost, "/books", body)

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			app.routes().ServeHTTP(rr, req)

			if rr.Code != tc.wantCode {
				t.Errorf("want status code %d; got %d", tc.wantCode, rr.Code)
			}

			if tc.wantKeys != nil {
				var resp map[string]any
				err := json.NewDecoder(rr.Body).Decode(&resp)
				if err != nil {
					t.Fatal(err)
				}

				errorsMap, ok := resp["errors"].(map[string]any)
				if !ok {
					t.Fatalf("expected 'errors' field in response, got: %#v", resp)
				}

				for _, key := range tc.wantKeys {
					if _, ok := errorsMap[key]; !ok {
						t.Errorf("expected error for key %q in response", key)
					}
				}
			}
		})
	}
}
