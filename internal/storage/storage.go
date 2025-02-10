package storage

import "github.com/jackc/pgx/v5/pgxpool"

// NewURLStorage creates a new storage instance based on type
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
