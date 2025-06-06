package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	"github.com/pcristin/urlshortener/internal/mytls"
	"github.com/pcristin/urlshortener/internal/proto"
	"github.com/pcristin/urlshortener/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
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

	grpcURL := config.GetGRPCURL()
	if grpcURL == "" {
		return errors.New("configuration error | gRPC address can not be empty")
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

	// Initialize service layer
	service := app.NewURLService(urlStorage, config.GetBaseURL(), config.GetTrustedSubnet())

	// Initialize HTTP handler
	handler := app.NewHandler(urlStorage, config)

	// Initialize gRPC server
	grpcServer := app.NewGRPCServer(service, config.GetBaseURL())

	// Create a wait group to manage both servers
	var wg sync.WaitGroup
	wg.Add(2)

	// Channel for errors from servers
	errChan := make(chan error, 2)

	// Channel for notify about server shutdown
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start HTTP server
	go func() {
		defer wg.Done()
		if err := runHTTPServer(ctx, handler, config, log); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start gRPC server
	go func() {
		defer wg.Done()
		if err := runGRPCServer(ctx, grpcServer, config, log); err != nil {
			errChan <- fmt.Errorf("gRPC server error: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-sigint:
		log.Infow("Shutdown signal received")
	case err := <-errChan:
		log.Errorw("Server error", "error", err)
	}

	// Cancel context to trigger shutdown
	cancel()

	// Wait for both servers to finish
	wg.Wait()

	log.Infow("All servers stopped")
	return nil
}

func runHTTPServer(ctx context.Context, handler app.HandlerInterface, config *config.Options, log *zap.SugaredLogger) error {
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
	r.Get("/api/internal/stats", logger.WithLogging(gzip.GzipMiddleware(handler.StatsHandler), log))

	serverURL := config.GetServerURL()
	log.Infow("Starting HTTP server", "address", serverURL, "https", config.GetEnableHTTPS())

	// Initialize server
	server := &http.Server{
		Addr:    serverURL,
		Handler: r,
	}

	// Graceful shutdown goroutine
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Errorw("HTTP server shutdown error", "error", err)
		}
	}()

	// Start server
	if config.GetEnableHTTPS() {
		certManager := mytls.GetTLSManager()
		server.TLSConfig = certManager.TLSConfig()
		return server.ListenAndServeTLS("", "")
	}
	return server.ListenAndServe()
}

func runGRPCServer(ctx context.Context, grpcServer *app.GRPCServer, config *config.Options, log *zap.SugaredLogger) error {
	grpcURL := config.GetGRPCURL()

	lis, err := net.Listen("tcp", grpcURL)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create auth interceptor
	authInterceptor := app.NewAuthInterceptor(config.GetSecret())

	// Server options
	var opts []grpc.ServerOption

	// Add interceptors
	opts = append(opts,
		grpc.UnaryInterceptor(authInterceptor.UnaryServerInterceptor()),
		grpc.StreamInterceptor(authInterceptor.StreamServerInterceptor()),
	)

	// Add TLS if enabled
	if config.GetEnableGRPCTLS() {
		certManager := mytls.GetTLSManager()
		creds := credentials.NewTLS(certManager.TLSConfig())
		opts = append(opts, grpc.Creds(creds))
	}

	// Create gRPC server
	s := grpc.NewServer(opts...)

	// Register service
	proto.RegisterURLShortenerServiceServer(s, grpcServer)

	// Register reflection service for debugging
	reflection.Register(s)

	log.Infow("Starting gRPC server", "address", grpcURL, "tls", config.GetEnableGRPCTLS())

	// Graceful shutdown goroutine
	go func() {
		<-ctx.Done()
		s.GracefulStop()
	}()

	// Start server
	return s.Serve(lis)
}
