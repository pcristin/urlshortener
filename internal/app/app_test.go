package app

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	"github.com/pcristin/urlshortener/internal/logger"
	mod "github.com/pcristin/urlshortener/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockStorage is test storage
type MockStorage struct {
	urls map[string]string
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		urls: make(map[string]string),
	}
}

func (m *MockStorage) AddURL(token, longURL string) error {
	m.urls[token] = longURL
	return nil
}

func (m *MockStorage) GetURL(token string) (string, error) {
	if url, ok := m.urls[token]; ok {
		return url, nil
	}
	return "", nil
}

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
			// Initialize mock storage and handler
			storage := NewMockStorage()
			handler := NewHandler(storage)

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
			name:       "testing ptorocol scheme",
			method:     http.MethodGet,
			token:      "A8gtZk8",
			storedURL:  "www.dzen.ru",
			wantStatus: http.StatusTemporaryRedirect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize mock storage and handler
			storage := NewMockStorage()
			if tt.storedURL != "" {
				err := storage.AddURL(tt.token, tt.storedURL)
				require.NoError(t, err)
			}

			handler := NewHandler(storage)

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

			assert.Equal(t, tt.wantStatus, resp.StatusCode)

			if tt.wantStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.storedURL, resp.Header.Get("Location"))
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
			// Initialize mock storage and handler
			storage := NewMockStorage()
			handler := NewHandler(storage)

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
