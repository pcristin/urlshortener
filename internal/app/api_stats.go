package app

import (
	"net/http"

	"github.com/mailru/easyjson"
)

// StatsHandler returns the stats of the URL shortener service (/api/internal/stats)
func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.Header.Get("X-Real-IP")

	// Get stats using service (includes trusted subnet check)
	stats, err := h.service.GetStats(r.Context(), clientIP)
	if err != nil {
		// If error contains "unauthorized", return 403
		if err.Error() == "unauthorized: IP not in trusted subnet" {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	statsBytes, err := easyjson.Marshal(stats)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(statsBytes)
}
