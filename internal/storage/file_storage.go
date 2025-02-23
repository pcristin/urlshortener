package storage

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailru/easyjson"
	"github.com/pcristin/urlshortener/internal/models"
)

type FileStorage struct {
	*MemoryStorage
	filePath string
}

func NewFileStorage(filePath string) *FileStorage {
	fs := &FileStorage{
		MemoryStorage: NewMemoryStorage(),
		filePath:      filePath,
	}
	if filePath != "" {
		fs.LoadFromFile(filePath)
	}
	return fs
}

func (fs *FileStorage) AddURL(token, longURL string, userID string) error {
	err := fs.MemoryStorage.AddURL(token, longURL, userID)
	if err != nil {
		return err
	}
	node, _ := fs.Get(token)
	// Ignore file operation errors
	_ = fs.appendToFile(node)
	return nil
}

func (fs *FileStorage) appendToFile(node models.URLStorageNode) error {
	if fs.filePath == "" {
		return nil
	}

	dir := filepath.Dir(fs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// If we can't create directory, just log and continue
		return nil
	}

	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		// If we can't open file, just log and continue
		return nil
	}
	defer file.Close()

	data, err := easyjson.Marshal(&node)
	if err != nil {
		return err
	}

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		// If we can't write to file, just log and continue
		return nil
	}
	return nil
}

func (fs *FileStorage) GetURL(token string) (string, error) {
	if node, ok := fs.Get(token); ok {
		if node.IsDeleted {
			return "", ErrURLDeleted
		}
		return node.OriginalURL, nil
	}
	return "", errors.New("URL not found")
}

func (fs *FileStorage) SaveToFile() error {
	if fs.filePath == "" {
		return nil
	}

	dir := filepath.Dir(fs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// If we can't create directory, just log and continue
		return nil
	}

	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		// If we can't open file, just log and continue
		return nil
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, node := range fs.cache {
		data, err := easyjson.Marshal(&node)
		if err != nil {
			return err
		}
		if _, err := writer.Write(append(data, '\n')); err != nil {
			// If we can't write to file, just log and continue
			return nil
		}
	}
	if err := writer.Flush(); err != nil {
		// If we can't flush to file, just log and continue
		return nil
	}
	return nil
}

func (fs *FileStorage) LoadFromFile(filepath string) error {
	fs.filePath = filepath

	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		// If we can't open file, just continue with empty storage
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var node models.URLStorageNode
		if err := easyjson.Unmarshal(scanner.Bytes(), &node); err != nil {
			// If we can't unmarshal a line, skip it and continue
			continue
		}
		fs.Set(node.ShortURL, node)
	}
	// Ignore scanner errors
	return nil
}

func (fs *FileStorage) SetDBPool(*pgxpool.Pool) {}

func (fs *FileStorage) GetStorageType() StorageType {
	return FileStorageType
}

func (fs *FileStorage) GetDBPool() *pgxpool.Pool {
	return nil
}

func (fs *FileStorage) AddURLBatch(urls map[string]string) error {
	// First add to memory
	err := fs.MemoryStorage.AddURLBatch(urls)
	if err != nil {
		return err
	}

	// Then append each URL to file, ignoring file operation errors
	for token := range urls {
		node, _ := fs.Get(token)
		_ = fs.appendToFile(node)
	}
	return nil
}

// Gets a token by original URL from file storage
func (fs *FileStorage) GetTokenByURL(longURL string) (string, error) {
	for _, node := range fs.cache {
		if node.OriginalURL == longURL {
			return node.ShortURL, nil
		}
	}
	return "", errors.New("url not found")
}

// DeleteURLs marks multiple URLs as deleted for a specific user
func (fs *FileStorage) DeleteURLs(userID string, tokens []string) error {
	err := fs.MemoryStorage.DeleteURLs(userID, tokens)
	if err != nil {
		return err
	}
	return fs.SaveToFile()
}
