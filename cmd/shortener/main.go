package main

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pcristin/urlshortener/internal/app"
	"github.com/pcristin/urlshortener/internal/config"
	"github.com/pcristin/urlshortener/internal/logger"
	"github.com/pcristin/urlshortener/internal/storage"
)

func main() {
	// Initialize logger
	log, err := logger.Initialize()

	if err != nil {
		panic("could not initialize logger")
	}

	defer log.Sync()

	// Initialize configuration and get server address from config
	config := config.NewOptions()
	config.ParseFlags()
	serverURL := config.GetServerURL()
	if serverURL == "" {
		//log.Fatalf("server address can not be empty!")
		log.Fatal("server address can not be empty!")
	}

	// Initialize storage
	urlStorage := storage.NewURLStorage()

	// Initialize handler with storage
	handler := app.NewHandler(urlStorage)

	r := chi.NewRouter()

	// Set up the middlewares: logger and 60s timeout
	// r.Use(middleware.Logger)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("/", logger.WithLogging(handler.EncodeURLHandler, log))
	r.Get("/{id}", logger.WithLogging(handler.DecodeURLHandler, log))

	log.Infow(
		"Running server on",
		"address", serverURL,
	)

	if err := http.ListenAndServe(serverURL, r); err != nil {
		log.Fatalf("error in ListenAndServe %v", err)
	}
}
