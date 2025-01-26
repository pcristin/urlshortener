package gzip

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	w     http.ResponseWriter
	gzipW *gzip.Writer
}

type GzipWriterInterface interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	Close() error
}

func NewGzipWriter(w http.ResponseWriter) GzipWriterInterface {
	return &gzipWriter{
		w:     w,
		gzipW: gzip.NewWriter(w),
	}
}

func (gw *gzipWriter) Header() http.Header {
	return gw.w.Header()
}

func (gw *gzipWriter) Write(data []byte) (int, error) {
	return gw.gzipW.Write(data)
}

func (gw *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		gw.w.Header().Set("Content-Encoding", "gzip")
	}

	gw.w.WriteHeader(statusCode)
}

func (gw *gzipWriter) Close() error {
	return gw.gzipW.Close()
}

type gzipReader struct {
	r     io.ReadCloser
	gzipR *gzip.Reader
}

type GzipReaderInterface interface {
	Read([]byte) (int, error)
	Close() error
}

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

func (gr *gzipReader) Read(data []byte) (int, error) {
	return gr.gzipR.Read(data)
}

func (gr *gzipReader) Close() error {
	if err := gr.r.Close(); err != nil {
		return err
	}
	return gr.gzipR.Close()
}

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
