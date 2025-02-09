package storage

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailru/easyjson"
	"github.com/pcristin/urlshortener/internal/models"
)

type URLStorage struct {
	storageType StorageType
	memStorage  map[string]models.URLStorageNode
	filePath    string
	dbPool      *pgxpool.Pool
}

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

// AddURL adds a new URL to storage for each of 3 storage types: memory,
// file or database
func (us *URLStorage) AddURL(token, longURL string) error {
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}

	node := models.URLStorageNode{
		UUID:        uuid.New(),
		ShortURL:    token,
		OriginalURL: longURL,
	}

	switch us.storageType {
	case DatabaseStorageType:
		return us.addToDB(token, longURL)
	case FileStorageType:
		if err := us.addToFile(node); err != nil {
			return err
		}
	case MemoryStorageType:
		us.memStorage[token] = node
	}

	return nil
}

// GetURL retrieves a URL from storage any type: memory,
// file or database
func (us *URLStorage) GetURL(token string) (string, error) {
	switch us.storageType {
	case DatabaseStorageType:
		return us.getFromDB(token)
	case FileStorageType:
		// Try memory first, then file if not found
		if url, err := us.getFromMemory(token); err == nil {
			return url, nil
		}
		return us.getFromFile(token)
	case MemoryStorageType:
		return us.getFromMemory(token)
	}
	return "", errors.New("invalid storage type")
}

// Memory storage get method
func (us *URLStorage) getFromMemory(token string) (string, error) {
	if node, ok := us.memStorage[token]; ok {
		return node.OriginalURL, nil
	}
	return "", errors.New("URL not found in memory")
}

// File storage methods add method
func (us *URLStorage) addToFile(node models.URLStorageNode) error {
	if us.filePath == "" {
		return errors.New("file path not set")
	}

	dir := filepath.Dir(us.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(us.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := easyjson.Marshal(&node)
	if err != nil {
		return err
	}

	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}

	return nil
}

// File storage get method
func (us *URLStorage) getFromFile(token string) (string, error) {
	file, err := os.OpenFile(us.filePath, os.O_RDONLY, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var node models.URLStorageNode
		if err := easyjson.Unmarshal(scanner.Bytes(), &node); err != nil {
			continue
		}
		if node.ShortURL == token {
			us.memStorage[token] = node // Cache in memory
			return node.OriginalURL, nil
		}
	}
	return "", errors.New("URL not found in file")
}

// Database storage add method
func (us *URLStorage) addToDB(token, longURL string) error {
	if us.dbPool == nil {
		return errors.New("database not initialized")
	}
	ctx := context.Background()
	_, err := us.dbPool.Exec(ctx,
		"INSERT INTO urls (token, original_url) VALUES ($1, $2)",
		token, longURL)
	return err
}

// Database storage get method
func (us *URLStorage) getFromDB(token string) (string, error) {
	if us.dbPool == nil {
		return "", errors.New("database not initialized")
	}
	ctx := context.Background()
	var longURL string
	err := us.dbPool.QueryRow(ctx,
		"SELECT original_url FROM urls WHERE token = $1",
		token).Scan(&longURL)
	return longURL, err
}

// SaveToFile saves URLStorageNode object into a file
func (us *URLStorage) SaveToFile() error {
	if us.filePath == "" {
		return nil
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(us.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(us.filePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, node := range us.memStorage {
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
	us.filePath = filepath

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
		us.memStorage[node.ShortURL] = node
	}

	return scanner.Err()
}

// Method to set database connection
func (us *URLStorage) SetDBPool(pool *pgxpool.Pool) {
	us.dbPool = pool
}

// Storage method to get the storage type:
// 0 - Memory Storage;
// 1 - File Storage;
// 2 - DB Storage
func (us *URLStorage) GetStorageType() StorageType {
	return us.storageType
}

// Database storage method to get the pool (connection) object
func (us *URLStorage) GetDBPool() *pgxpool.Pool {
	return us.dbPool
}
