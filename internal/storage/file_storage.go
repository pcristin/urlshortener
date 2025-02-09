package storage

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailru/easyjson"
	"github.com/pcristin/urlshortener/internal/models"
)

type FileStorage struct {
	urls     map[string]models.URLStorageNode
	filePath string
}

func NewFileStorage(filePath string) *FileStorage {
	fs := &FileStorage{
		urls:     make(map[string]models.URLStorageNode),
		filePath: filePath,
	}
	if filePath != "" {
		fs.LoadFromFile(filePath)
	}
	return fs
}

func (fs *FileStorage) AddURL(token, longURL string) error {
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}

	node := models.URLStorageNode{
		UUID:        uuid.New(),
		ShortURL:    token,
		OriginalURL: longURL,
	}

	fs.urls[token] = node

	if fs.filePath != "" {
		return fs.appendToFile(node)
	}
	return nil
}

func (fs *FileStorage) appendToFile(node models.URLStorageNode) error {
	dir := filepath.Dir(fs.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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

func (fs *FileStorage) GetURL(token string) (string, error) {
	if node, ok := fs.urls[token]; ok {
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
		return err
	}

	file, err := os.OpenFile(fs.filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, node := range fs.urls {
		data, err := easyjson.Marshal(&node)
		if err != nil {
			return err
		}
		if _, err := writer.Write(append(data, '\n')); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (fs *FileStorage) LoadFromFile(filepath string) error {
	fs.filePath = filepath

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
		fs.urls[node.ShortURL] = node
	}
	return scanner.Err()
}

func (fs *FileStorage) SetDBPool(*pgxpool.Pool) {}

func (fs *FileStorage) GetStorageType() StorageType {
	return FileStorageType
}

func (fs *FileStorage) GetDBPool() *pgxpool.Pool {
	return nil
}
