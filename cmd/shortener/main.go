package main

import (
	"net/http"
	"time"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pcristin/urlshortener/internal/app"
	"github.com/pcristin/urlshortener/internal/config"
	"github.com/pcristin/urlshortener/internal/database"
	"github.com/pcristin/urlshortener/internal/gzip"
	"github.com/pcristin/urlshortener/internal/logger"
	"github.com/pcristin/urlshortener/internal/storage"
	"go.uber.org/zap"
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
		log.Fatal("server address can not be empty!")
	}

	// Determine storage type based on config
	var storageType storage.StorageType
	var dbPool *pgxpool.Pool
	var filePath string

	if databaseDSN := config.GetDatabaseDSN(); databaseDSN != "" {
		zap.L().Sugar().Infow("Database config", "databaseDSN", databaseDSN)
		dbManager, err := database.NewDatabaseManager(databaseDSN)
		if err != nil {
			log.Warnf("Failed to connect to database: %v", err)
		} else {
			storageType = storage.DatabaseStorageType
			dbPool = dbManager.GetPool()
		}
	} else if filePath = config.GetPathToSavedData(); filePath != "" {
		storageType = storage.FileStorageType
	} else {
		storageType = storage.MemoryStorageType
	}

	// Initialize storage with determined type
	urlStorage := storage.NewURLStorage(storageType, filePath, dbPool)

	// Initialize handler with storage and config
	handler := app.NewHandler(urlStorage, config)

	r := chi.NewRouter()

	// Set up the middlewares: 60s timeout
	r.Use(middleware.Timeout(60 * time.Second))

	r.Post("/", logger.WithLogging(gzip.GzipMiddleware(handler.AuthMiddleware(handler.EncodeURLHandler)), log))
	r.Get("/{id}", logger.WithLogging(gzip.GzipMiddleware(handler.DecodeURLHandler), log))
	r.Post("/api/shorten", logger.WithLogging(gzip.GzipMiddleware(handler.AuthMiddleware(handler.APIEncodeHandler)), log))
	r.Post("/api/shorten/batch", logger.WithLogging(gzip.GzipMiddleware(handler.AuthMiddleware(handler.APIEncodeBatchHandler)), log))
	r.Get("/ping", logger.WithLogging(handler.PingHandler, log))
	r.Get("/api/user/urls", logger.WithLogging(gzip.GzipMiddleware(handler.AuthMiddleware(handler.GetUserURLsHandler)), log))
	r.Delete("/api/user/urls", logger.WithLogging(gzip.GzipMiddleware(handler.AuthMiddleware(handler.DeleteUserURLsHandler)), log))

	log.Infow(
		"Running server on",
		"address", serverURL,
	)

	if err := http.ListenAndServe(serverURL, r); err != nil {
		log.Fatalf("error in ListenAndServe %v", err)
	}
}
