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

type StorageType int

const (
	MemoryStorage StorageType = iota
	FileStorage
	DatabaseStorage
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

type URLStorage struct {
	storageType StorageType
	memStorage  map[string]models.URLStorageNode
	filePath    string
	dbPool      *pgxpool.Pool
}

// InitStorage initializes the URL storage
func NewURLStorage(storageType StorageType, filePath string, dbPool *pgxpool.Pool) URLStorager {
	return &URLStorage{
		storageType: storageType,
		memStorage:  make(map[string]models.URLStorageNode),
		filePath:    filePath,
		dbPool:      dbPool,
	}
}

// AddURL adds a new URL to storage
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
	case DatabaseStorage:
		return us.addToDB(token, longURL)
	case FileStorage:
		if err := us.addToFile(node); err != nil {
			return err
		}
	case MemoryStorage:
		us.memStorage[token] = node
	}

	return nil
}

// GetURL retrieves a URL from storage
func (us *URLStorage) GetURL(token string) (string, error) {
	switch us.storageType {
	case DatabaseStorage:
		return us.getFromDB(token)
	case FileStorage:
		// Try memory first, then file if not found
		if url, err := us.getFromMemory(token); err == nil {
			return url, nil
		}
		return us.getFromFile(token)
	case MemoryStorage:
		return us.getFromMemory(token)
	}
	return "", errors.New("invalid storage type")
}

// Memory storage methods
func (us *URLStorage) getFromMemory(token string) (string, error) {
	if node, ok := us.memStorage[token]; ok {
		return node.OriginalURL, nil
	}
	return "", errors.New("URL not found in memory")
}

// File storage methods
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

// Database storage methods
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

// New method to set database connection
func (us *URLStorage) SetDBPool(pool *pgxpool.Pool) {
	us.dbPool = pool
}

func (us *URLStorage) GetStorageType() StorageType {
	return us.storageType
}

func (us *URLStorage) GetDBPool() *pgxpool.Pool {
	return us.dbPool
}
