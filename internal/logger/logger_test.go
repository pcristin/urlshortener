package logger

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	// Test that logger initialization works
	logger, err := Initialize()
	require.NoError(t, err)
	assert.NotNil(t, logger)
}

func TestWithLogging(t *testing.T) {
	// Initialize the logger
	logger, err := Initialize()
	require.NoError(t, err)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap the handler with logging
	loggedHandler := WithLogging(testHandler, logger)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	// Call the handler
	loggedHandler(rec, req)

	// Verify the response
	resp := rec.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestLoggingResponseWriter(t *testing.T) {
	// Create a base response writer
	rec := httptest.NewRecorder()

	// Create a logging response writer
	responseData := &responseData{
		status: 0,
		size:   0,
	}
	lw := loggingResponseWriter{
		ResponseWriter: rec,
		responseData:   responseData,
	}

	// Write a response
	lw.WriteHeader(http.StatusOK)
	n, err := lw.Write([]byte("test response"))

	// Verify the results
	assert.NoError(t, err)
	assert.Equal(t, 13, n)
	assert.Equal(t, http.StatusOK, responseData.status)
	assert.Equal(t, 13, responseData.size)

	// Test that writing the header twice doesn't change the status
	lw.WriteHeader(http.StatusBadRequest)
	assert.Equal(t, http.StatusOK, responseData.status)
}
