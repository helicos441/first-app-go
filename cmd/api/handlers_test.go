package main

import (
	"database/sql"
	"encoding/json"
	"first-app-go/internal/data"
	"fmt"
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

func TestPutBookHandler_ValidInput(t *testing.T) {
	// Set up a fresh in-memory app
	app := setupTestApp(t)

	// First, create a book to update
	createBody := strings.NewReader(`{
		"title": "Original Title",
		"author": "Original Author",
		"year": 2020
	}`)
	reqCreate := httptest.NewRequest(http.MethodPost, "/books", createBody)
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	app.routes().ServeHTTP(rrCreate, reqCreate)

	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("failed to create book: got status %d", rrCreate.Code)
	}

	var created data.Book
	if err := json.NewDecoder(rrCreate.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	// Now update the book
	updateBody := strings.NewReader(fmt.Sprintf(`{
		"title": "Updated Title",
		"author": "Updated Author",
		"year": 2021
	}`))
	reqUpdate := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/books/%d", created.ID), updateBody)
	reqUpdate.Header.Set("Content-Type", "application/json")
	rrUpdate := httptest.NewRecorder()

	app.routes().ServeHTTP(rrUpdate, reqUpdate)

	if rrUpdate.Code != http.StatusOK {
		t.Errorf("want status code %d; got %d", http.StatusOK, rrUpdate.Code)
	}

	var updated data.Book
	if err := json.NewDecoder(rrUpdate.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}

	// Assert the updated fields
	if updated.ID != created.ID {
		t.Errorf("expected ID %d; got %d", created.ID, updated.ID)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title'; got %q", updated.Title)
	}
	if updated.Author != "Updated Author" {
		t.Errorf("expected author 'Updated Author'; got %q", updated.Author)
	}
	if updated.Year != 2021 {
		t.Errorf("expected year 2021; got %d", updated.Year)
	}

	// Confirm it's also updated in the DB
	stored, err := app.Stores.Books.Get(updated.ID)
	if err != nil {
		t.Fatalf("failed to fetch book from DB: %v", err)
	}
	if *stored != updated {
		t.Errorf("book in DB does not match response. got: %#v", stored)
	}
}

func TestPutBookHandler_InvalidInput(t *testing.T) {
	app := setupTestApp(t)

	// Create a book first (we need an existing one to update)
	createBody := strings.NewReader(`{
		"title": "Original Title",
		"author": "Original Author",
		"year": 2020
	}`)
	reqCreate := httptest.NewRequest(http.MethodPost, "/books", createBody)
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	app.routes().ServeHTTP(rrCreate, reqCreate)

	var created data.Book
	if err := json.NewDecoder(rrCreate.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		payload  string
		wantCode int
		wantKeys []string
	}{
		{
			name:     "missing all fields",
			payload:  `{}`,
			wantCode: http.StatusUnprocessableEntity,
			wantKeys: []string{"title", "author", "year"},
		},
		{
			name:     "invalid year",
			payload:  `{"title":"T","author":"A","year":0}`,
			wantCode: http.StatusUnprocessableEntity,
			wantKeys: []string{"year"},
		},
		{
			name:     "invalid JSON",
			payload:  `{`,
			wantCode: http.StatusBadRequest,
			wantKeys: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			updateBody := strings.NewReader(tc.payload)
			reqUpdate := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/books/%d", created.ID), updateBody)
			reqUpdate.Header.Set("Content-Type", "application/json")
			rrUpdate := httptest.NewRecorder()

			app.routes().ServeHTTP(rrUpdate, reqUpdate)

			if rrUpdate.Code != tc.wantCode {
				t.Errorf("want status code %d; got %d", tc.wantCode, rrUpdate.Code)
			}

			if tc.wantKeys != nil {
				var resp map[string]any
				if err := json.NewDecoder(rrUpdate.Body).Decode(&resp); err != nil {
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

func TestPatchBookHandler_ValidInput(t *testing.T) {
	// Step 1: Set up a fresh in-memory app
	// Same as in our other handler tests — this gives us a clean DB for each run.
	app := setupTestApp(t)

	// Step 2: First, create a book to update.
	// We POST a book just like before so that we have an existing record to PATCH.
	createBody := strings.NewReader(`{
		"title": "Original Title",
		"author": "Original Author",
		"year": 2020
	}`)
	// Create a book using the post endpoint
	// - create post request
	// - create recorder
	// - serveHTTP
	reqCreate := httptest.NewRequest(http.MethodPost, "/books", createBody)
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	app.routes().ServeHTTP(rrCreate, reqCreate)

	// Check created status code
	// Use t.Fatalf if code is incorrect...means book ws not created..can't proceed
	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("failed to create book: got status %d", rrCreate.Code)
	}

	// decode response body into a data.Book
	var created data.Book
	if err := json.NewDecoder(rrCreate.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	// Step 3: Send a PATCH request to update just one field.
	// 🔑 DIFFERENCE: With PUT we sent all fields, here we send only the field we want to change.
	// - create patchBody
	patchBody := strings.NewReader(`{
		"year": 2025
	}`)
	// - create patch request
	// - create recorder
	// - serveHTTP
	reqPatch := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/books/%d", created.ID), patchBody)
	reqPatch.Header.Set("Content-Type", "application/json")
	rrPatch := httptest.NewRecorder()

	app.routes().ServeHTTP(rrPatch, reqPatch)

	// Step 4: Assert that the response status code is 200 OK.
	// Same as PUT — success should still be 200.
	if rrPatch.Code != http.StatusOK {
		t.Errorf("want status code %d; got %d", http.StatusOK, rrPatch.Code)
	}

	// Step 5: Decode the response JSON into a Book struct.
	var updated data.Book
	if err := json.NewDecoder(rrPatch.Body).Decode(&updated); err != nil {
		t.Fatal(err)
	}

	// Step 6: Assert the updated field has changed.
	// 🔑 DIFFERENCE: With PUT we asserted all fields; here we only assert the field we patched.
	if updated.Year != 2025 {
		t.Errorf("expected year to be 2025; got %d", updated.Year)
	}

	// Step 7: Assert the unchanged fields remain the same.
	// 🔑 DIFFERENCE: This is new for PATCH — we’re explicitly checking that untouched fields stayed intact.
	if updated.Title != created.Title {
		t.Errorf("expected title to remain %q; got %q", created.Title, updated.Title)
	}
	if updated.Author != created.Author {
		t.Errorf("expected author to remain %q; got %q", created.Author, updated.Author)
	}

	// Step 8: Confirm the updated record in the DB.
	// Same as PUT — fetch from DB and compare with our updated response.
	stored, err := app.Stores.Books.Get(updated.ID)
	if err != nil {
		t.Fatalf("failed to fetch book from DB: %v", err)
	}
	if *stored != updated {
		t.Errorf("book in DB does not match response. got: %#v", stored)
	}
}

func TestPatchBookHandler_InvalidInput(t *testing.T) {
	app := setupTestApp(t)

	// Create a book first
	createBody := strings.NewReader(`{
		"title": "Original Title",
		"author": "Original Author",
		"year": 2020
	}`)
	reqCreate := httptest.NewRequest(http.MethodPost, "/books", createBody)
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	app.routes().ServeHTTP(rrCreate, reqCreate)

	var created data.Book
	if err := json.NewDecoder(rrCreate.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		payload  string
		wantCode int
		wantKeys []string
	}{
		{
			name:     "invalid year (zero)",
			payload:  `{"year": 0}`,
			wantCode: http.StatusUnprocessableEntity,
			wantKeys: []string{"year"},
		},
		{
			name:     "empty title",
			payload:  `{"title": ""}`,
			wantCode: http.StatusUnprocessableEntity,
			wantKeys: []string{"title"},
		},
		{
			name:     "invalid JSON",
			payload:  `{`,
			wantCode: http.StatusBadRequest,
			wantKeys: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			patchBody := strings.NewReader(tc.payload)
			reqPatch := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/books/%d", created.ID), patchBody)
			reqPatch.Header.Set("Content-Type", "application/json")
			rrPatch := httptest.NewRecorder()

			app.routes().ServeHTTP(rrPatch, reqPatch)

			if rrPatch.Code != tc.wantCode {
				t.Errorf("want status code %d; got %d", tc.wantCode, rrPatch.Code)
			}

			if tc.wantKeys != nil {
				var resp map[string]any
				if err := json.NewDecoder(rrPatch.Body).Decode(&resp); err != nil {
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

func TestDeleteBookHandler_RecordFound(t *testing.T) {
	// Step 1: Set up a fresh in-memory app with seeded data
	app := setupTestApp(t)

	// Step 2: Create a new book to delete (we use POST to insert a real record)
	createBody := strings.NewReader(`{
		"title": "To Delete",
		"author": "Some Author",
		"year": 2000
	}`)
	reqCreate := httptest.NewRequest(http.MethodPost, "/books", createBody)
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	app.routes().ServeHTTP(rrCreate, reqCreate)

	// Step 3: Ensure it was created successfully
	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("failed to create book for deletion: got status %d", rrCreate.Code)
	}

	// Step 4: Decode the response so we know the book ID
	var created data.Book
	if err := json.NewDecoder(rrCreate.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}

	// Step 5: Make DELETE request to remove the book by ID
	reqDelete := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/books/%d", created.ID), nil)
	rrDelete := httptest.NewRecorder()
	app.routes().ServeHTTP(rrDelete, reqDelete)

	// Step 6: Assert response is 204 No Content
	if rrDelete.Code != http.StatusNoContent {
		t.Errorf("expected status 204; got %d", rrDelete.Code)
	}

	// Step 7: Try to retrieve the book from the DB to confirm it's deleted
	_, err := app.Stores.Books.Get(created.ID)
	if err == nil {
		t.Errorf("expected book to be deleted, but it still exists")
	}
}

func TestDeleteBookHandler_RecordNotFound(t *testing.T) {
	// Step 1: Set up a clean test app
	app := setupTestApp(t)

	// Step 2: Send DELETE request for a book that doesn’t exist
	req := httptest.NewRequest(http.MethodDelete, "/books/999", nil)
	rr := httptest.NewRecorder()
	app.routes().ServeHTTP(rr, req)

	// Step 3: Assert response is 404 Not Found
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404; got %d", rr.Code)
	}
}
