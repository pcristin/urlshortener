package storage

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Common errors
var (
	ErrURLExists = errors.New("url already exists")
)

type StorageType int

const (
	MemoryStorageType StorageType = iota
	FileStorageType
	DatabaseStorageType
)

// URLStorager defines the interface for URL storage operations
type URLStorager interface {
	AddURL(token, longURL string) error
	GetURL(token string) (string, error)
	SaveToFile() error
	LoadFromFile(filepath string) error
	SetDBPool(*pgxpool.Pool)
	GetStorageType() StorageType
	GetDBPool() *pgxpool.Pool
	AddURLBatch(urls map[string]string) error
	GetTokenByURL(longURL string) (string, error)
}
