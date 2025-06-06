package app

import (
	"net/http"
)

// Handler to check the connectivity to the database
func (h *Handler) PingHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "bad request", http.StatusBadRequest)
		return
	}

	// Check database connectivity using service
	dbOk, err := h.service.Ping(req.Context())
	if err != nil || !dbOk {
		http.Error(res, "internal server error", http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
