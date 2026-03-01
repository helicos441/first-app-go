package main

import (
	"encoding/json"
	"net/http"
)

const version = "1.0.0"

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthcheck)

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
