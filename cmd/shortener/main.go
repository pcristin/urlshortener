package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pcristin/urlshortener/internal/app"
	"github.com/pcristin/urlshortener/internal/config"
	"github.com/pcristin/urlshortener/internal/storage"
)

func main() {
	// Initialize configuration and get server address from config
	config := config.NewOptions()
	config.ParseFlags()
	serverURL := config.GetServerURL()
	if serverURL == "" {
		log.Fatalf("server address can not be empty!")
	}

	// Initialize storage
	urlStorage := storage.NewURLStorage()

	// Initialize handler with storage
	handler := app.NewHandler(urlStorage)

	r := chi.NewRouter()

	// Set up the middlewares: logger and 60s timeout
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("/", handler.EncodeURLHandler)
	r.Get("/{id}", handler.DecodeURLHandler)

	log.Printf("Running server on %s\n", serverURL)

	if err := http.ListenAndServe(serverURL, r); err != nil {
		log.Fatalf("error in ListenAndServe %v", err)
	}
}
