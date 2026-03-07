package main

import (
	"first-app-go/internal/data"
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
	// Get the value of id
	idString := r.PathValue("id")
	// Convert to an int for the db lookup
	id, err := strconv.ParseInt(idString, 10, 64)
	// Validate the id
	if err != nil || id < 1 {
		// Return not found if can't be validated
		http.NotFound(w, r)
		return
	}

	// For now, we return a hard-coded book.
	// Later we’ll replace this with a real database lookup.
	book := data.Book{
		ID:     id,
		Title:  "Stub",
		Author: "N/A",
		Year:   0,
	}

	// Write the json response
	if err := writeJSON(w, http.StatusOK, book); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
