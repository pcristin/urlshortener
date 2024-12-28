package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pcristin/urlshortener/internal/app"
	"github.com/pcristin/urlshortener/internal/config"
)

func main() {
	config.FlagParse()
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/", app.EncodeURLHandler)
	r.Get("/{id}", app.DecodeURLHandler)

	if err := http.ListenAndServe(config.OptionsFlag.ServerURL, r); err != nil {
		log.Fatalf("error in ListenAndServe: %v", err)
	}
}
