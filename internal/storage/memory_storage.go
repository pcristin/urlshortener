package storage

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pcristin/urlshortener/internal/models"
)

// MemoryStorage implements URLStorager interface with in-memory storage
type MemoryStorage struct {
	BaseStorage
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{BaseStorage: NewBaseStorage()}
}

// AddURL adds a new URL to the in-memory storage
func (ms *MemoryStorage) AddURL(token, longURL string, userID string) error {
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}

	// Instead of always checking all existing URLs, first check if the token exists
	// which is more efficient as a map lookup
	if _, exists := ms.cache[token]; exists {
		return errors.New("token already exists")
	}

	// Only check for duplicate URLs if we need to enforce uniqueness
	// This is a costly operation, so we might want to make it configurable
	// or skip it if performance is critical in the future
	for _, node := range ms.cache {
		if node.OriginalURL == longURL && !node.IsDeleted {
			return ErrURLExists
		}
	}

	node := models.URLStorageNode{
		UUID:        uuid.New(),
		ShortURL:    token,
		OriginalURL: longURL,
		UserID:      userID,
	}
	ms.Set(token, node)
	return nil
}

// GetURL retrieves a URL by its token from in-memory storage
func (ms *MemoryStorage) GetURL(token string) (string, error) {
	if node, ok := ms.Get(token); ok {
		if node.IsDeleted {
			return "", ErrURLDeleted
		}
		return node.OriginalURL, nil
	}
	return "", errors.New("URL not found")
}

// SaveToFile is a no-op for memory storage
func (ms *MemoryStorage) SaveToFile() error {
	return nil
}

// LoadFromFile is a no-op for memory storage
func (ms *MemoryStorage) LoadFromFile(filepath string) error {
	return nil
}

// SetDBPool is a no-op for memory storage
func (ms *MemoryStorage) SetDBPool(*pgxpool.Pool) {}

// GetStorageType returns the type of storage (memory)
func (ms *MemoryStorage) GetStorageType() StorageType {
	return MemoryStorageType
}

// GetDBPool returns nil for memory storage
func (ms *MemoryStorage) GetDBPool() *pgxpool.Pool {
	return nil
}

// AddURLBatch adds multiple URLs to storage in a single operation
func (ms *MemoryStorage) AddURLBatch(urls map[string]string) error {
	for token, longURL := range urls {
		node := models.URLStorageNode{
			UUID:        uuid.New(),
			ShortURL:    token,
			OriginalURL: longURL,
		}
		ms.Set(token, node)
	}
	return nil
}

// GetTokenByURL retrieves a token associated with a long URL
func (ms *MemoryStorage) GetTokenByURL(longURL string) (string, error) {
	if token, ok := ms.BaseStorage.GetTokenByURL(longURL); ok {
		return token, nil
	}
	return "", errors.New("url not found")
}

// GetUserURLs returns all URLs shortened by a specific user
func (ms *MemoryStorage) GetUserURLs(userID string) ([]models.URLStorageNode, error) {
	// Preallocate with a reasonable capacity to reduce allocations
	userURLs := make([]models.URLStorageNode, 0, len(ms.cache)/4)

	for _, node := range ms.cache {
		if node.UserID == userID {
			userURLs = append(userURLs, node)
		}
	}
	return userURLs, nil
}

// DeleteURLs marks multiple URLs as deleted for a specific user
func (ms *MemoryStorage) DeleteURLs(userID string, tokens []string) error {
	if len(tokens) == 0 {
		return nil
	}

	for _, token := range tokens {
		if node, ok := ms.Get(token); ok && node.UserID == userID {
			node.IsDeleted = true
			ms.Set(token, node)
		}
	}
	return nil
}
