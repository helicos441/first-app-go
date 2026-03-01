package main

import (
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", healthcheck)

	http.ListenAndServe(":8080", mux)
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello, world"))
}
