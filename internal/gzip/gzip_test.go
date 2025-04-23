package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipMiddleware(t *testing.T) {
	// Test handler that returns a simple response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap the handler with the gzip middleware
	wrappedHandler := GzipMiddleware(handler)

	tests := []struct {
		name            string
		acceptEncoding  string
		contentEncoding string
		compressed      bool
	}{
		{
			name:            "with gzip accept-encoding",
			acceptEncoding:  "gzip",
			contentEncoding: "gzip",
			compressed:      true,
		},
		{
			name:            "without gzip accept-encoding",
			acceptEncoding:  "",
			contentEncoding: "",
			compressed:      false,
		},
		{
			name:            "with multiple encodings including gzip",
			acceptEncoding:  "deflate, gzip, br",
			contentEncoding: "gzip",
			compressed:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.contentEncoding, resp.Header.Get("Content-Encoding"))

			var body []byte
			var err error

			if tt.compressed {
				// If response should be compressed, decompress it
				reader, err := gzip.NewReader(resp.Body)
				require.NoError(t, err)
				defer reader.Close()
				body, err = io.ReadAll(reader)
				require.NoError(t, err)
			} else {
				// Otherwise read the body directly
				body, err = io.ReadAll(resp.Body)
			}

			require.NoError(t, err)
			assert.Equal(t, "test response", string(body))
		})
	}
}

func TestNewGzipWriter(t *testing.T) {
	// Create a response recorder
	rec := httptest.NewRecorder()

	// Create a gzip writer
	gw := NewGzipWriter(rec)

	// Write something to the gzip writer
	data := "test data for compression"
	n, err := gw.Write([]byte(data))

	// Check the result
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)

	// Close the writer to flush the data
	gw.Close()

	// Verify the response is gzipped
	assert.NotEqual(t, data, rec.Body.String())

	// Decompress and verify the content
	reader, err := gzip.NewReader(strings.NewReader(rec.Body.String()))
	require.NoError(t, err)
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, data, string(decompressed))
}
