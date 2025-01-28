package storage

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/pcristin/urlshortener/internal/models"
)

// URLStorager defines the interface for URL storage operations
type URLStorager interface {
	AddURL(token, longURL string) error
	GetURL(token string) (string, error)
	SaveToFile() error
	LoadFromFile(filepath string) error
}

type URLStorage struct {
	Storage  map[string]models.URLStorageNode
	filepath string
}

// InitStorage initializes the URL storage
func NewURLStorage() URLStorager {
	return &URLStorage{
		Storage:  make(map[string]models.URLStorageNode),
		filepath: "", // Will be set via LoadFromFile
	}
}

// AddURL adds a new URL to storage
func (us *URLStorage) AddURL(token, longURL string) error {
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}

	us.Storage[token] = models.URLStorageNode{
		UUID:        uuid.New(),
		ShortURL:    token,
		OriginalURL: longURL,
	}

	// Save immediately after adding
	if err := us.SaveToFile(); err != nil {
		return err
	}

	return nil
}

// GetURL retrieves a URL from storage
func (us *URLStorage) GetURL(token string) (string, error) {
	if nodeStorage, found := us.Storage[token]; found {
		return nodeStorage.OriginalURL, nil
	}
	return "", errors.New("URL not found")
}

// SaveToFile saves URLStorageNode object into a file
func (us *URLStorage) SaveToFile() error {
	if us.filepath == "" {
		return nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(us.filepath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(us.filepath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, node := range us.Storage {
		data, err := easyjson.Marshal(&node)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
		if _, err := writer.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return writer.Flush()
}

// LoadFromFile loads data from json file into URLStorageNode
func (us *URLStorage) LoadFromFile(filepath string) error {
	us.filepath = filepath

	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var node models.URLStorageNode
		if err := easyjson.Unmarshal(scanner.Bytes(), &node); err != nil {
			return err
		}
		us.Storage[node.ShortURL] = node
	}

	return scanner.Err()
}
