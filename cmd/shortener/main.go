package main

import (
	"net/http"

	"github.com/pcristin/urlshortener/internal/app"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.EncodeURLHandler)
	mux.HandleFunc("/{id}", app.DecodeURLHandler)

	err := http.ListenAndServe("localhost:8080", mux)
	if err != nil {
		panic("Something went wrong!")
	}
}
