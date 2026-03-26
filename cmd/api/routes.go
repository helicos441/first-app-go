package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"first-app-go/internal/data"
	"first-app-go/internal/request"
	"log"
	"net/http"
	"strconv"
)

type bookResponse struct {
	Books []data.Book `json:"books"`
}

func (app *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", app.healthcheckHandler)
	mux.HandleFunc("GET /books", app.listBooksHandler)
	mux.HandleFunc("GET /books/{id}", app.showBookHandler)
	mux.HandleFunc("POST /books", app.createBookHandler)
	mux.HandleFunc("PUT /books/{id}", app.putBookHandler)
	mux.HandleFunc("PATCH /books/{id}", app.patchBookHandler)
	return mux
}

func (app *App) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	response := healthResponse{
		Status:  "ok",
		Version: version,
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *App) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	books, err := app.Stores.Books.GetAll()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := bookResponse{Books: books}

	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *App) showBookHandler(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	id, err := strconv.ParseInt(idString, 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	book, err := app.Stores.Books.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r) // 404
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, book); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *App) createBookHandler(w http.ResponseWriter, r *http.Request) {
	var br request.FullBookRequest

	if err := json.NewDecoder(r.Body).Decode(&br); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	validationErrors := request.ValidateFullBookRequest(&br)
	if len(validationErrors) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": validationErrors})
		return
	}

	book := &data.Book{
		Title:  br.Title,
		Author: br.Author,
		Year:   br.Year,
	}

	book, err := app.Stores.Books.Insert(book)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if err := writeJSON(w, http.StatusCreated, book); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *App) putBookHandler(w http.ResponseWriter, r *http.Request) {
	idPath := r.PathValue("id")
	id, err := strconv.ParseInt(idPath, 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	var br request.FullBookRequest
	if err := json.NewDecoder(r.Body).Decode(&br); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	validationErrors := request.ValidateFullBookRequest(&br)
	if len(validationErrors) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": validationErrors})
		return
	}

	book, err := app.Stores.Books.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	book.Title = br.Title
	book.Author = br.Author
	book.Year = br.Year

	// Step 6: Save the updated book to the DB
	updatedBook, err := app.Stores.Books.Update(book)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			http.NotFound(w, r)
		default:
			log.Printf("failed to update book: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// Step 7: Return the updated book as JSON with a 200 OK status.
	if err := writeJSON(w, http.StatusOK, updatedBook); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *App) patchBookHandler(w http.ResponseWriter, r *http.Request) {
	idPath := r.PathValue("id")
	id, err := strconv.ParseInt(idPath, 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}

	var br request.PartialBookRequest
	if err := json.NewDecoder(r.Body).Decode(&br); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	// Step 3: Validate the input

	// Step 4: Retrieve the existing book

	// Step 5: Replace the provided fields on the book

	// Step 6: Save the updated book to the DB
	// Stub this for the time being

	// Step 7: Return the updated book as JSON with a 200 OK status.
}
