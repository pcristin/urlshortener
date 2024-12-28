package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pcristin/urlshortener/internal/app"
	"github.com/pcristin/urlshortener/internal/config"
)

func main() {
	// Print initial state
	fmt.Printf("Initial ServerURL: %q\n", config.OptionsFlag.ServerURL)

	config.FlagParse()

	// Print after parsing
	fmt.Printf("After parsing ServerURL: %q\n", config.OptionsFlag.ServerURL)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/", app.EncodeURLHandler)
	r.Get("/{id}", app.DecodeURLHandler)

	fmt.Printf("Running server on ServerURL=%q\n", config.OptionsFlag.ServerURL)

	if err := http.ListenAndServe(config.OptionsFlag.ServerURL, r); err != nil {
		log.Fatalf("error in ListenAndServe: %v", err)
	}
}
