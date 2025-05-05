package storage

import (
	"github.com/pcristin/urlshortener/internal/models"
)

// BaseStorage holds common in-memory cache functionality
type BaseStorage struct {
	cache    map[string]models.URLStorageNode
	urlIndex map[string]string // Maps original URLs to tokens for faster lookups
}

// NewBaseStorage initializes the base storage
func NewBaseStorage() BaseStorage {
	return BaseStorage{
		cache:    make(map[string]models.URLStorageNode),
		urlIndex: make(map[string]string),
	}
}

// Get retrieves a cached node
func (bs *BaseStorage) Get(token string) (models.URLStorageNode, bool) {
	node, ok := bs.cache[token]
	return node, ok
}

// Set caches a node
func (bs *BaseStorage) Set(token string, node models.URLStorageNode) {
	bs.cache[token] = node
	bs.urlIndex[node.OriginalURL] = token
}

// GetTokenByURL returns a token for a given URL using the index
func (bs *BaseStorage) GetTokenByURL(url string) (string, bool) {
	token, ok := bs.urlIndex[url]
	return token, ok
}
