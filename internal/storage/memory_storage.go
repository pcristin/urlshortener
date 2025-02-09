package storage

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pcristin/urlshortener/internal/models"
)

type MemoryStorage struct {
	urls map[string]models.URLStorageNode
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		urls: make(map[string]models.URLStorageNode),
	}
}

func (ms *MemoryStorage) AddURL(token, longURL string) error {
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}
	ms.urls[token] = models.URLStorageNode{
		UUID:        uuid.New(),
		ShortURL:    token,
		OriginalURL: longURL,
	}
	return nil
}

func (ms *MemoryStorage) GetURL(token string) (string, error) {
	if node, ok := ms.urls[token]; ok {
		return node.OriginalURL, nil
	}
	return "", errors.New("URL not found")
}

func (ms *MemoryStorage) SaveToFile() error {
	return nil
}

func (ms *MemoryStorage) LoadFromFile(filepath string) error {
	return nil
}

func (ms *MemoryStorage) SetDBPool(*pgxpool.Pool) {}

func (ms *MemoryStorage) GetStorageType() StorageType {
	return MemoryStorageType
}

func (ms *MemoryStorage) GetDBPool() *pgxpool.Pool {
	return nil
}
