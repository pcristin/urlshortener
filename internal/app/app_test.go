package app

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	myGzip "github.com/pcristin/urlshortener/internal/gzip"
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
