package models

//go:generate easyjson -all request_models.go

// Request represents a request to shorten a single URL
//
//easyjson:json
type Request struct {
	URL string `json:"url"`
}

// Response contains the result of a successful URL shortening operation
//
//easyjson:json
type Response struct {
	Result string `json:"result"`
}

// BatchRequestItem represents a single URL in a batch shortening request
//
//easyjson:json
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchResponseItem represents a single result in a batch shortening response
//
//easyjson:json
type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// BatchRequest is a collection of URLs to be shortened in a single request
//
//easyjson:json
type BatchRequest []BatchRequestItem

// BatchResponse is a collection of shortened URLs returned in response to a batch request
//
//easyjson:json
type BatchResponse []BatchResponseItem
