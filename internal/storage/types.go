package storage

import "github.com/jackc/pgx/v5/pgxpool"

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
}
