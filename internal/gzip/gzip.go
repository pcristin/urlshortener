package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipWriter wraps an http.ResponseWriter and provides gzip compression
type gzipWriter struct {
	w     http.ResponseWriter
	gzipW *gzip.Writer
}

// GzipWriterInterface defines methods for a response writer with gzip compression
type GzipWriterInterface interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	Close() error
}

// NewGzipWriter creates a new gzip writer that implements GzipWriterInterface
func NewGzipWriter(w http.ResponseWriter) GzipWriterInterface {
	return &gzipWriter{
		w:     w,
		gzipW: gzip.NewWriter(w),
	}
}

// Header returns the header map of the underlying ResponseWriter
func (gw *gzipWriter) Header() http.Header {
	return gw.w.Header()
}

// Write compresses the data and writes it to the underlying ResponseWriter
func (gw *gzipWriter) Write(data []byte) (int, error) {
	return gw.gzipW.Write(data)
}

// WriteHeader sets the status code and adds gzip content encoding header
func (gw *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		gw.w.Header().Set("Content-Encoding", "gzip")
	}

	gw.w.WriteHeader(statusCode)
}

// Close closes the gzip writer to flush any remaining data
func (gw *gzipWriter) Close() error {
	return gw.gzipW.Close()
}

// gzipReader wraps an io.ReadCloser and provides gzip decompression
type gzipReader struct {
	r     io.ReadCloser
	gzipR *gzip.Reader
}

// GzipReaderInterface defines methods for a reader with gzip decompression
type GzipReaderInterface interface {
	Read([]byte) (int, error)
	Close() error
}

// NewGzipReader creates a new gzip reader that implements GzipReaderInterface
func NewGzipReader(r io.ReadCloser) (GzipReaderInterface, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:     r,
		gzipR: zr,
	}, nil
}

// Read decompresses data from the underlying reader
func (gr *gzipReader) Read(data []byte) (int, error) {
	return gr.gzipR.Read(data)
}

// Close closes both the gzip reader and the underlying reader
func (gr *gzipReader) Close() error {
	if err := gr.r.Close(); err != nil {
		return err
	}
	return gr.gzipR.Close()
}

// GzipMiddleware provides HTTP middleware that handles gzip compression and decompression
// It automatically compresses responses and decompresses requests with gzip Content-Encoding
func GzipMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		originalWriter := w

		acceptEncoding := req.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")

		contentType := req.Header.Get("Content-Type")
		supportsContentTypes := strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")

		if supportsGzip && supportsContentTypes {
			gzipW := NewGzipWriter(w)
			originalWriter = gzipW
			defer gzipW.Close()
		}

		contentEncoding := req.Header.Get("Content-Encoding")
		sentGzip := strings.Contains(contentEncoding, "gzip")

		if sentGzip {
			gzipReader, err := NewGzipReader(req.Body)
			if err != nil {
				http.Error(w, "unsupported content encoding", http.StatusInternalServerError)
				return
			}

			req.Body = gzipReader
			defer gzipReader.Close()
		}

		h.ServeHTTP(originalWriter, req)
	}
}
