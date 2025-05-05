package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	// responseData holds HTTP response information for logging
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter wraps http.ResponseWriter to capture response data for logging
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write captures the size of the response data being written
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader captures the status code of the response
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	if r.responseData.status != 0 {
		return
	}
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// WithLogging is middleware that logs details about HTTP requests and responses
func WithLogging(h http.HandlerFunc, log *zap.SugaredLogger) http.HandlerFunc {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		log.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return logFn
}

// Initialize sets up and returns a configured zap logger for the application
func Initialize() (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()

	config.Level = zap.NewAtomicLevel()

	prodLogger, err := config.Build()

	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(prodLogger)
	return prodLogger.Sugar(), nil
}
