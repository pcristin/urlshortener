package models

//go:generate easyjson -all request_models.go

//easyjson:json
type Request struct {
	URL string `json:"url"`
}

//easyjson:json
type Response struct {
	Result string `json:"result"`
}

//easyjson:json
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

//easyjson:json
type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

//easyjson:json
type BatchRequest []BatchRequestItem

//easyjson:json
type BatchResponse []BatchResponseItem
