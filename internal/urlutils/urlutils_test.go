package urlutils

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pcristin/urlshortener/internal/models"
	"github.com/pcristin/urlshortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorager is a mock implementation of storage.URLStorager
type MockStorager struct {
	mock.Mock
}

func (m *MockStorager) AddURL(token, longURL string, userID string) error {
	args := m.Called(token, longURL, userID)
	return args.Error(0)
}

func (m *MockStorager) GetURL(token string) (string, error) {
	args := m.Called(token)
	return args.String(0), args.Error(1)
}

func (m *MockStorager) SaveToFile() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockStorager) LoadFromFile(filepath string) error {
	args := m.Called(filepath)
	return args.Error(0)
}

func (m *MockStorager) SetDBPool(pool *pgxpool.Pool) {
	m.Called(pool)
}

func (m *MockStorager) GetStorageType() storage.StorageType {
	args := m.Called()
	return args.Get(0).(storage.StorageType)
}

func (m *MockStorager) GetDBPool() *pgxpool.Pool {
	args := m.Called()
	return args.Get(0).(*pgxpool.Pool)
}

func (m *MockStorager) AddURLBatch(urls map[string]string) error {
	args := m.Called(urls)
	return args.Error(0)
}

func (m *MockStorager) GetTokenByURL(longURL string) (string, error) {
	args := m.Called(longURL)
	return args.String(0), args.Error(1)
}

func (m *MockStorager) GetUserURLs(userID string) ([]models.URLStorageNode, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.URLStorageNode), args.Error(1)
}

func (m *MockStorager) DeleteURLs(userID string, tokens []string) error {
	args := m.Called(userID, tokens)
	return args.Error(0)
}

func TestEncodeURL(t *testing.T) {
	// Create a mock storage
	mockStorage := new(MockStorager)
	userID := "test-user"
	longURL := "https://github.com/pcristin/urlshortener"

	// Test when URL doesn't exist
	mockStorage.On("GetTokenByURL", longURL).Return("", errors.New("url not found"))
	mockStorage.On("AddURL", mock.Anything, longURL, userID).Return(nil)

	token, err := EncodeURL(longURL, mockStorage, userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockStorage.AssertExpectations(t)
}

func TestDecodeURL(t *testing.T) {
	// Create a mock storage
	mockStorage := new(MockStorager)
	token := "abc123"
	longURL := "https://github.com/pcristin/urlshortener"

	// Test when token exists
	mockStorage.On("GetURL", token).Return(longURL, nil)

	result, err := DecodeURL(token, mockStorage)
	assert.NoError(t, err)
	assert.Equal(t, longURL, result)
	mockStorage.AssertExpectations(t)

	// Test when token doesn't exist
	mockStorage = new(MockStorager)
	mockStorage.On("GetURL", "nonexistent").Return("", errors.New("url not found"))

	result, err = DecodeURL("nonexistent", mockStorage)
	assert.Error(t, err)
	assert.Empty(t, result)
	mockStorage.AssertExpectations(t)
}
