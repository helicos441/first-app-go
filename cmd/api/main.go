package main

import (
	"database/sql"
	"encoding/json"
	"first-app-go/internal/data"
	"log"
	"net/http"
)

const version = "1.0.0"

type App struct {
	DB *sql.DB
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func main() {
	db, err := data.OpenSQLite()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = data.Migrate(db)
	if err != nil {
		log.Fatal(err)
	}

	if err := data.SeedIfEmpty(db); err != nil {
		log.Fatal(err)
	}

	app := &App{DB: db}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthcheck)
	mux.HandleFunc("GET /books", app.listBooksHandler)

	println("Ready on localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	response := healthResponse{
		Status:  "OK",
		Version: version,
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (app *App) listBooksHandler(w http.ResponseWriter, r *http.Request) {
	books := []data.Book{
		{ID: 1, Title: "The Go Programming Language", Author: "Alan Donovan", Year: 2015},
		{ID: 2, Title: "Designing Data-Intensive Applications", Author: "Martin Kleppmann", Year: 2017},
	}

	if err := writeJSON(w, http.StatusOK, books); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(b)
	return err
}
