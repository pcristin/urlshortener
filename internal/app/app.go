package app

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mailru/easyjson"
	mod "github.com/pcristin/urlshortener/internal/models"
	"github.com/pcristin/urlshortener/internal/storage"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
	"go.uber.org/zap"
)

type HandlerInterface interface {
	EncodeURLHandler(http.ResponseWriter, *http.Request)
	DecodeURLHandler(http.ResponseWriter, *http.Request)
	APIEncodeHandler(http.ResponseWriter, *http.Request)
	APIEncodeBatchHandler(http.ResponseWriter, *http.Request)
	PingHandler(http.ResponseWriter, *http.Request)
}

type Handler struct {
	storage storage.URLStorager
}

func NewHandler(storage storage.URLStorager) HandlerInterface {
	return &Handler{
		storage: storage,
	}
}

// Handler to encode URL with plain text and without compressing the data
func (h *Handler) EncodeURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	longURL, err := io.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil || len(string(longURL)) == 0 {
		http.Error(res, "bad request: incorrect long URL", http.StatusBadRequest)
		return
	}

	token, err := uu.EncodeURL(string(longURL), h.storage)
	if err != nil {
		http.Error(res, "bad request: unable to shorten provided url", http.StatusBadRequest)
		return
	}

	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(http.StatusCreated)
	resBody := "http://" + req.Host + "/" + token
	res.Write([]byte(resBody))
}

// Handler to decode encoded long URL
func (h *Handler) DecodeURLHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	token := chi.URLParam(req, "id")
	if token == "" {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	defer req.Body.Close()

	longURL, err := uu.DecodeURL(token, h.storage)
	if err != nil || longURL == "" {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	res.Header().Set("Location", longURL)
	res.Header().Del("Date")
	res.Header().Del("Content-Type")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// Handler to encode the url with compressed data
func (h *Handler) APIEncodeHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost || req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	var body mod.Request
	err := easyjson.UnmarshalFromReader(req.Body, &body)
	defer req.Body.Close()

	if err != nil || len(body.URL) == 0 {
		http.Error(res, "bad request: incorrect url", http.StatusBadRequest)
		return
	}

	// Encode the long URL to a short URL
	shortURL, err := uu.EncodeURL(body.URL, h.storage)
	if err != nil {
		http.Error(res, "bad request: unable to shorten provided url", http.StatusBadRequest)
		return
	}

	// Prepare the response payload
	response := mod.Response{
		Result: "http://" + req.Host + "/" + shortURL,
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	responseBytes, err := easyjson.Marshal(response)
	if err != nil {
		http.Error(res, "internal server error: unable to marshal response", http.StatusInternalServerError)
	}
	res.Write(responseBytes)
}

// Handler to check the connectivity to the database
func (h *Handler) PingHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	// Get the database storage (this handler only applicable for DB storage)
	storage, ok := h.storage.(*storage.DatabaseStorage)
	if !ok || storage.GetDBPool() == nil {
		http.Error(res, "database not configured", http.StatusInternalServerError)
		return
	}

	if err := storage.GetDBPool().Ping(req.Context()); err != nil {
		http.Error(res, "internal server error", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}

// APIEncodeBatchHandler encodes a batch of sent urls
func (h *Handler) APIEncodeBatchHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost || req.Header.Get("Content-Type") != "application/json" {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	var batchRequests mod.BatchRequest
	err := easyjson.UnmarshalFromReader(req.Body, &batchRequests)
	defer req.Body.Close()

	if err != nil {
		http.Error(res, "bad request: invalid JSON", http.StatusBadRequest)
		return
	}

	if len(batchRequests) == 0 {
		http.Error(res, "bad request: empty batch", http.StatusBadRequest)
		return
	}

	// Prepare batch of URLs
	urlBatch := make(map[string]string)
	responses := make(mod.BatchResponse, 0, len(batchRequests))

	for _, item := range batchRequests {
		token := uu.GenerateToken()
		urlBatch[token] = item.OriginalURL
		responses = append(responses, mod.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      "http://" + req.Host + "/" + token,
		})
	}

	// Save batch with temp logging
	if err := h.storage.AddURLBatch(urlBatch); err != nil {
		zap.L().Sugar().Errorw("Error in AddURLBatch", "storageType", h.storage.GetStorageType(), "error", err)
		http.Error(res, "internal server error", http.StatusInternalServerError)
		return
	}

	// Send response
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)

	responseBytes, err := easyjson.Marshal(responses)
	if err != nil {
		http.Error(res, "internal server error: unable to marshal response", http.StatusInternalServerError)
		return
	}
	res.Write(responseBytes)
}
