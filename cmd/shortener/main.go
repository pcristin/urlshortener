package main

import (
	"log"
	"net/http"
	"urlshortener/internal/app"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.EncodeURLHandler)
	mux.HandleFunc("/{id}", app.DecodeURLHandler)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		log.Fatalf("Something went wrong! Error: %v", err)
	}
}
