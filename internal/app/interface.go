package app

import "net/http"

type HandlerInterface interface {
	EncodeURLHandler(http.ResponseWriter, *http.Request)
	DecodeURLHandler(http.ResponseWriter, *http.Request)
	APIEncodeHandler(http.ResponseWriter, *http.Request)
	APIEncodeBatchHandler(http.ResponseWriter, *http.Request)
	PingHandler(http.ResponseWriter, *http.Request)
	GetUserURLsHandler(http.ResponseWriter, *http.Request)
	AuthMiddleware(http.HandlerFunc) http.HandlerFunc
}
