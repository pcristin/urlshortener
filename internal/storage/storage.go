package storage

import "errors"

// URLStorager defines the interface for URL storage operations
type URLStorager interface {
	AddURL(token, longURL string) error
	GetURL(token string) (string, error)
}

type URLStorage struct {
	Storage map[string]string
}

// InitStorage initializes the URL storage
func NewURLStorage() URLStorager {
	return &URLStorage{
		Storage: make(map[string]string),
	}
}

// AddURL adds a new URL to storage
func (us *URLStorage) AddURL(token, longURL string) error {
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}
	us.Storage[token] = longURL
	return nil
}

// GetURL retrieves a URL from storage
func (us *URLStorage) GetURL(token string) (string, error) {
	if url, found := us.Storage[token]; found {
		return url, nil
	}
	return "", errors.New("URL not found")
}
