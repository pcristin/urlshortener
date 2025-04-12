package app

import (
	"net/http"

	"github.com/pcristin/urlshortener/internal/storage"
)

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
