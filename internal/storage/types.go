package storage

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pcristin/urlshortener/internal/models"
)

// Common errors
var (
	// ErrURLExists is returned when attempting to add a URL that already exists in storage
	ErrURLExists = errors.New("url already exists")
	// ErrURLDeleted is returned when attempting to access a URL that has been marked as deleted
	ErrURLDeleted = errors.New("url was deleted")
)

// StorageType defines the type of storage mechanism used for URL data
type StorageType int

// StorageType constants
const (
	// MemoryStorageType represents in-memory storage (volatile)
	MemoryStorageType StorageType = iota
	// FileStorageType represents file-based persistent storage
	FileStorageType
	// DatabaseStorageType represents database-backed persistent storage
	DatabaseStorageType
)

// URLStorager defines the interface for URL storage operations
type URLStorager interface {
	// AddURL adds a new URL to storage with an associated token and user ID
	AddURL(token, longURL string, userID string) error

	// GetURL retrieves the original URL associated with a token
	GetURL(token string) (string, error)

	// SaveToFile persists the current state to a file (for file-based storage)
	SaveToFile() error

	// LoadFromFile loads the state from a file (for file-based storage)
	LoadFromFile(filepath string) error

	// SetDBPool sets the database connection pool (for database storage)
	SetDBPool(*pgxpool.Pool)

	// GetStorageType returns the type of storage being used
	GetStorageType() StorageType

	// GetDBPool returns the database connection pool (for database storage)
	GetDBPool() *pgxpool.Pool

	// AddURLBatch adds multiple URLs to storage in a single operation
	AddURLBatch(urls map[string]string) error

	// GetTokenByURL retrieves the token associated with a long URL
	GetTokenByURL(longURL string) (string, error)

	// GetUserURLs retrieves all URLs associated with a specific user
	GetUserURLs(userID string) ([]models.URLStorageNode, error)

	// DeleteURLs marks the specified URLs as deleted for a given user
	DeleteURLs(userID string, tokens []string) error
}

// NewURLStorage creates a new storage instance based on type
// It returns a specific implementation of URLStorager depending on the provided type:
// - DatabaseStorageType: A database-backed storage using the provided connection pool
// - FileStorageType: A file-based storage using the provided file path
// - MemoryStorageType (default): An in-memory storage
func NewURLStorage(storageType StorageType, filePath string, dbPool *pgxpool.Pool) URLStorager {
	switch storageType {
	case DatabaseStorageType:
		return NewDatabaseStorage(dbPool)
	case FileStorageType:
		return NewFileStorage(filePath)
	default:
		return NewMemoryStorage()
	}
}
