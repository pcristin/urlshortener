package app

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	myGzip "github.com/pcristin/urlshortener/internal/gzip"
	"github.com/pcristin/urlshortener/internal/logger"
	mod "github.com/pcristin/urlshortener/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage is test storage
type MockStorage struct {
	urls     map[string]string
	filepath string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		urls: make(map[string]string),
	}
}

func (m *MockStorage) AddURL(token, longURL string) error {
	if token == "" || longURL == "" {
		return errors.New("token and URL cannot be empty")
	}
	m.urls[token] = longURL
	return nil
}

func (m *MockStorage) GetURL(token string) (string, error) {
	if url, ok := m.urls[token]; ok {
		return url, nil
	}
	return "", errors.New("URL not found")
}

func (m *MockStorage) SaveToFile() error {
	if m.filepath == "" {
		return nil
	}

	file, err := os.OpenFile(m.filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for shortURL, longURL := range m.urls {
		node := mod.URLStorageNode{
			UUID:        uuid.New(),
			ShortURL:    shortURL,
			OriginalURL: longURL,
		}
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

func (m *MockStorage) LoadFromFile(filepath string) error {
	m.filepath = filepath

	file, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var node mod.URLStorageNode
		if err := easyjson.Unmarshal(scanner.Bytes(), &node); err != nil {
			return err
		}
		m.urls[node.ShortURL] = node.OriginalURL
	}

	return scanner.Err()
}

type MockDatabaseManager struct {
	shouldError bool
}

func NewMockDatabaseManager() *MockDatabaseManager {
	return &MockDatabaseManager{
		shouldError: false,
	}
}

func (m *MockDatabaseManager) Ping(ctx context.Context) error {
	if m.shouldError {
		return fmt.Errorf("mock database error")
	}
	return nil
}

func (m *MockDatabaseManager) Close() {}

func TestEncodeURLHandler(t *testing.T) {
	// Initialize logger
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	tests := []struct {
		name        string
		method      string
		url         string
		body        string
		contentType string
		wantStatus  int
	}{
		{
			name:        "valid url",
			method:      http.MethodPost,
			url:         "/",
			body:        "https://google.com",
			contentType: "text/plain; charset=utf-8",
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "empty url",
			method:      http.MethodPost,
			url:         "/",
			body:        "",
			contentType: "text/plain; charset=utf-8",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "invalid url",
			method:      http.MethodPost,
			url:         "/",
			body:        "not-a-url",
			contentType: "text/plain; charset=utf-8",
			wantStatus:  http.StatusCreated,
		},
		{
			name:        "wrong method",
			method:      http.MethodGet,
			url:         "/",
			body:        "https://google.com",
			contentType: "text/plain; charset=utf-8",
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockStorage()
			dbManager := NewMockDatabaseManager()
			ctx := context.Background()

			handler := NewHandler(storage, dbManager, ctx)
			// Wrap the handler with logging
			loggedHandler := logger.WithLogging(handler.EncodeURLHandler, log)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			loggedHandler(w, req)

			// Check response
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusCreated {
				assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestDecodeURLHandler(t *testing.T) {
	// Initialize logger
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	tests := []struct {
		name       string
		method     string
		token      string
		storedURL  string
		wantStatus int
	}{
		{
			name:       "existing url",
			method:     http.MethodGet,
			token:      "abc123",
			storedURL:  "https://google.com",
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:       "non existing url",
			method:     http.MethodGet,
			token:      "nonexistent",
			storedURL:  "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "testing protocol scheme",
			method:     http.MethodGet,
			token:      "A8gtZk8",
			storedURL:  "www.dzen.ru",
			wantStatus: http.StatusTemporaryRedirect,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			token:      "abc123",
			storedURL:  "https://google.com",
			wantStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize storage and populate it
			storage := NewMockStorage()
			dbManager := NewMockDatabaseManager()
			ctx := context.Background()

			// Pre-populate storage with test data if storedURL is not empty
			if tt.storedURL != "" {
				err := storage.AddURL(tt.token, tt.storedURL)
				require.NoError(t, err, "Failed to populate storage")
			}

			handler := NewHandler(storage, dbManager, ctx)

			// Wrap the handler with logging
			loggedHandler := logger.WithLogging(handler.DecodeURLHandler, log)

			// Create chi router for URL parameter handling
			r := chi.NewRouter()
			r.Get("/{id}", loggedHandler)

			// Create request
			req := httptest.NewRequest(tt.method, "/"+tt.token, nil)
			w := httptest.NewRecorder()

			// Serve the request
			r.ServeHTTP(w, req)

			// Check response
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode,
				"Expected status %d but got %d for test %s",
				tt.wantStatus, resp.StatusCode, tt.name)

			if tt.wantStatus == http.StatusTemporaryRedirect {
				location := resp.Header.Get("Location")
				assert.Equal(t, tt.storedURL, location,
					"Expected location %s but got %s for test %s",
					tt.storedURL, location, tt.name)
			}
		})
	}
}

func TestApiEncodeHandler(t *testing.T) {
	// Initialize logger
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	tests := []struct {
		name               string
		url                string
		method             string
		headersContentType string
		longURL            string
		wantStatus         int
	}{
		{
			name:               "good url",
			url:                "/api/shorten",
			method:             http.MethodPost,
			headersContentType: "application/json",
			longURL:            "https://google.com",
			wantStatus:         http.StatusCreated,
		},
		{
			name:               "empty long url",
			url:                "/api/shorten",
			method:             http.MethodPost,
			headersContentType: "application/json",
			longURL:            "",
			wantStatus:         http.StatusBadRequest,
		},
		{
			name:               "incorrect request headers",
			url:                "/api/shorten",
			method:             http.MethodPost,
			headersContentType: "text/plain",
			longURL:            "www.google.com",
			wantStatus:         http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMockStorage()
			dbManager := NewMockDatabaseManager()
			ctx := context.Background()

			handler := NewHandler(storage, dbManager, ctx)
			// Wrap the handler with logging
			loggedHandler := logger.WithLogging(handler.APIEncodeHandler, log)

			// Create request body in JSON format
			requestBody := mod.Request{
				URL: tt.longURL,
			}

			// Marshal the request body to JSON
			bodyBytes, err := easyjson.Marshal(requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(tt.method, tt.url, bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", tt.headersContentType)

			// Create response recorder
			w := httptest.NewRecorder()

			// Call handler
			loggedHandler(w, req)

			// Check response
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusCreated {
				assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestCompressionMiddleware(t *testing.T) {
	// Initialize logger
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	// Sample handler to test middleware
	sampleHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Hello, World!"}`))
	}

	// Wrap the sample handler with GzipMiddleware
	handler := myGzip.GzipMiddleware(sampleHandler)

	tests := []struct {
		name               string
		acceptEncoding     string
		contentEncoding    string
		contentType        string
		requestBody        string
		expectedHeader     string
		expectedBody       string
		expectedStatusCode int
	}{
		{
			name:               "Client supports gzip, response should be compressed",
			acceptEncoding:     "gzip",
			contentEncoding:    "",
			contentType:        "application/json",
			requestBody:        "",
			expectedHeader:     "gzip",
			expectedBody:       `{"message":"Hello, World!"}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Client does not support gzip, response should not be compressed",
			acceptEncoding:     "",
			contentEncoding:    "",
			contentType:        "application/json",
			requestBody:        "",
			expectedHeader:     "",
			expectedBody:       `{"message":"Hello, World!"}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Client sends gzip compressed request, server should decompress",
			acceptEncoding:     "",
			contentEncoding:    "gzip",
			contentType:        "application/json",
			requestBody:        `{"message":"Hello, Server!"}`,
			expectedHeader:     "",
			expectedBody:       `{"message":"Hello, World!"}`,
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Unsupported Content-Type, should not compress",
			acceptEncoding:     "gzip",
			contentEncoding:    "",
			contentType:        "application/xml",
			requestBody:        "",
			expectedHeader:     "",
			expectedBody:       `{"message":"Hello, World!"}`,
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var reqBody io.Reader
			if tt.contentEncoding == "gzip" {
				var buf bytes.Buffer
				gzipWriter := gzip.NewWriter(&buf)
				_, err := gzipWriter.Write([]byte(tt.requestBody))
				require.NoError(t, err)
				gzipWriter.Close()
				reqBody = &buf
			} else {
				reqBody = bytes.NewBufferString(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodGet, "/", reqBody)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}
			req.Header.Set("Content-Type", tt.contentType)

			// Create response recorder
			rr := httptest.NewRecorder()

			// Serve the request
			handler(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			// Check Content-Encoding header
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, rr.Header().Get("Content-Encoding"))
			} else {
				assert.Empty(t, rr.Header().Get("Content-Encoding"))
			}

			// Check response body
			if tt.expectedBody != "" {
				var responseBody string
				if tt.expectedHeader == "gzip" {
					// Decompress response body
					gzipReader, err := gzip.NewReader(rr.Body)
					require.NoError(t, err)
					decompressedBody, err := io.ReadAll(gzipReader)
					require.NoError(t, err)
					gzipReader.Close()
					responseBody = string(decompressedBody)
				} else {
					responseBody = rr.Body.String()
				}
				assert.Equal(t, tt.expectedBody, responseBody)
			}
		})
	}
}

func TestFileStorage(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "urlshortener_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up after test

	testFile := filepath.Join(tmpDir, "test_urls.json")

	tests := []struct {
		name    string
		urls    map[string]string // map[shortURL]longURL
		wantErr bool
	}{
		{
			name: "basic save and load",
			urls: map[string]string{
				"abc123": "https://google.com",
				"def456": "https://github.com",
			},
			wantErr: false,
		},
		{
			name:    "empty storage",
			urls:    map[string]string{},
			wantErr: false,
		},
		{
			name: "multiple urls",
			urls: map[string]string{
				"abc123": "https://google.com",
				"def456": "https://github.com",
				"ghi789": "https://example.com",
				"jkl012": "https://test.com",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize storage
			storage := NewMockStorage()

			// Add URLs to storage
			for shortURL, longURL := range tt.urls {
				err := storage.AddURL(shortURL, longURL)
				require.NoError(t, err)
			}

			// Save to file
			storage.filepath = testFile // Set filepath before saving
			err = storage.SaveToFile()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Create new storage instance
			newStorage := NewMockStorage()

			// Load from file
			err = newStorage.LoadFromFile(testFile)
			require.NoError(t, err)

			// Verify all URLs were loaded correctly
			for shortURL, longURL := range tt.urls {
				loaded, err := newStorage.GetURL(shortURL)
				require.NoError(t, err)
				assert.Equal(t, longURL, loaded)
			}
		})
	}
}

func TestPingHandler(t *testing.T) {
	log, err := logger.Initialize()
	require.NoError(t, err)
	defer log.Sync()

	tests := []struct {
		name       string
		method     string
		dbError    bool
		wantStatus int
	}{
		{
			name:       "successful ping",
			method:     http.MethodGet,
			dbError:    false,
			wantStatus: http.StatusOK,
		},
		{
			name:       "database error",
			method:     http.MethodGet,
			dbError:    true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "wrong method",
			method:     http.MethodPost,
			dbError:    false,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mocks and handler
			storage := NewMockStorage()
			dbManager := NewMockDatabaseManager()
			dbManager.shouldError = tt.dbError
			ctx := context.Background()

			handler := NewHandler(storage, dbManager, ctx)

			// Wrap the handler with logging
			loggedHandler := logger.WithLogging(handler.PingHandler, log)

			// Create request
			req := httptest.NewRequest(tt.method, "/ping", nil)
			w := httptest.NewRecorder()

			// Call handler
			loggedHandler(w, req)

			// Check response
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
