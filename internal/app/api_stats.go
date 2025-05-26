package app

import (
	"net"
	"net/http"

	"github.com/mailru/easyjson"
)

// StatsHandler returns the stats of the URL shortener service (/api/internal/stats)
func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	clientIP := r.Header.Get("X-Real-IP")
	trusted := h.trustedSubnet

	if trusted != "" {
		_, trustedIP, err := net.ParseCIDR(trusted)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		if !trustedIP.Contains(net.ParseIP(clientIP)) {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	stats, err := h.storage.GetStats()
	if err != nil {
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
