package main

import (
	"encoding/json"
	"first-app-go/internal/data"
	"log"
	"net/http"
)

const version = "1.0.0"

type App struct {
	Stores data.Stores
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

	app := &App{Stores: data.NewStores(db)}

	log.Println("starting server on :8080")
	if err := http.ListenAndServe(":8080", app.routes()); err != nil {
		log.Fatal(err)
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
