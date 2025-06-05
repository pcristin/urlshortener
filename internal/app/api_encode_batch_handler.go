package app

import (
	"net/http"

	"github.com/mailru/easyjson"
	mod "github.com/pcristin/urlshortener/internal/models"
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

	// Convert to service layer format
	items := make([]BatchItem, len(batchRequests))
	for i, req := range batchRequests {
		items[i] = BatchItem{
			CorrelationID: req.CorrelationID,
			OriginalURL:   req.OriginalURL,
		}
	}

	// Process URLs using service
	results, err := h.service.ShortenBatch(req.Context(), items, userID)
	if err != nil {
		http.Error(res, "internal server error", http.StatusInternalServerError)
		return
	}

	// Convert results to response format
	responses := make(mod.BatchResponse, len(results))
	for i, result := range results {
		responses[i] = mod.BatchResponseItem{
			CorrelationID: result.CorrelationID,
			ShortURL:      h.constructURL(result.ShortURL, req),
		}
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
