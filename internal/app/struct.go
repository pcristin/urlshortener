package app

import (
	"net/http"

	"github.com/pcristin/urlshortener/internal/config"
	"github.com/pcristin/urlshortener/internal/storage"
	"go.uber.org/zap"
)

type Handler struct {
	storage storage.URLStorager
	secret  string
	baseURL string
	logger  *zap.Logger
}

func NewHandler(storage storage.URLStorager, config *config.Options) HandlerInterface {
	secret := config.GetSecret()
	if secret == "" {
		secret = "your-secret-key" // fallback for tests and development
	}

	return &Handler{
		storage: storage,
		secret:  secret,
		baseURL: config.GetBaseURL(),
		logger:  zap.L(),
	}
}

// constructURL builds the full URL for a shortened link
func (h *Handler) constructURL(token string, r *http.Request) string {
	if h.baseURL != "" {
		return h.baseURL + "/" + token
	}
	return "http://" + r.Host + "/" + token
}
