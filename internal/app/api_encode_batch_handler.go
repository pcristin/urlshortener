package app

import (
	"errors"
	"net/http"

	"github.com/mailru/easyjson"
	mod "github.com/pcristin/urlshortener/internal/models"
	"github.com/pcristin/urlshortener/internal/storage"
	uu "github.com/pcristin/urlshortener/internal/urlutils"
	"go.uber.org/zap"
)

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

	// Get user ID from context
	userID := getUserIDFromContext(req.Context())

	// Process URLs and collect responses
	responses := make(mod.BatchResponse, 0, len(batchRequests))

	for _, item := range batchRequests {
		token, err := uu.EncodeURL(item.OriginalURL, h.storage, userID)
		if err != nil && !errors.Is(err, storage.ErrURLExists) {
			zap.L().Sugar().Errorw("Error encoding URL", "error", err, "url", item.OriginalURL)
			http.Error(res, "internal server error", http.StatusInternalServerError)
			return
		}
		responses = append(responses, mod.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      h.constructURL(token, req),
		})
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
