package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pcristin/urlshortener/internal/app"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/", app.EncodeURLHandler)
	r.Get("/{id}", app.DecodeURLHandler)

	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		log.Fatalf("error in ListenAndServe: %v", err)
	}
}
