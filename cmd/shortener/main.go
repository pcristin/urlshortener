package main

import (
	"errors"
	"fmt"
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

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	if err := run(); err != nil {
		// Log the error before exiting
		if logger, err := logger.Initialize(); err == nil {
			logger.Errorw("application error", "error", err)
			logger.Sync()
		} else {
			fmt.Printf("logger error: %v, original error: %v\n", err, err)
		}
	}
}

// run encapsulates the main application logic and returns an error instead of exiting directly
func run() error {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	// Initialize logger
	log, err := logger.Initialize()
	if err != nil {
		return fmt.Errorf("logger error | failed to initialize logger: %w", err)
	}

	// Flush logs
	defer log.Sync()

	// Initialize configuration and get server address from config
	config := config.NewOptions()
	config.ParseFlags()

	serverURL := config.GetServerURL()
	if serverURL == "" {
		return errors.New("configuration error | server address can not be empty")
	}

	// Determine storage type based on config
	var storageType storage.StorageType
	var dbPool *pgxpool.Pool
	var filePath string

	if databaseDSN := config.GetDatabaseDSN(); databaseDSN != "" {
		zap.L().Sugar().Infow("Database config", "databaseDSN", databaseDSN)
		dbManager, err := database.NewDatabaseManager(databaseDSN)
		if err != nil {
			log.Warnf("database error | failed to connect to database: %v", err)
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
		return fmt.Errorf("server error | failed to listen and serve: %w", err)
	}

	return nil
}
