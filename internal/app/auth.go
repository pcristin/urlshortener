package app

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
)

const (
	userIDCookieName    = "user_id"
	signatureCookieName = "signature"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// getUserIDFromContext retrieves user ID from context
func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value(userIDContextKey).(string); ok {
		return userID
	}
	return ""
}

// setUserIDToContext adds user ID to context
func setUserIDToContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

// generateSignature creates an HMAC signature for the given user ID
func generateSignature(userID string, secret []byte) string {
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(userID))
	return hex.EncodeToString(h.Sum(nil))
}

// validateSignature validates the HMAC signature for the given user ID
func validateSignature(userID, signature string, secret []byte) bool {
	expectedSignature := generateSignature(userID, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// AuthMiddleware handles user authentication via cookies
func (h *Handler) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try to get existing cookies
		userIDCookie, err := r.Cookie(userIDCookieName)
		signatureCookie, signErr := r.Cookie(signatureCookieName)

		// If either cookie is missing or invalid, create new ones
		if err != nil || signErr != nil || !validateSignature(userIDCookie.Value, signatureCookie.Value, []byte(h.secret)) {
			// Generate new user ID and signature
			userID := uuid.New().String()
			signature := generateSignature(userID, []byte(h.secret))

			// Set new cookies
			http.SetCookie(w, &http.Cookie{
				Name:  userIDCookieName,
				Value: userID,
				Path:  "/",
			})
			http.SetCookie(w, &http.Cookie{
				Name:  signatureCookieName,
				Value: signature,
				Path:  "/",
			})

			// Update request with new user ID
			r = r.WithContext(setUserIDToContext(r.Context(), userID))
		} else {
			// Update request with existing user ID
			r = r.WithContext(setUserIDToContext(r.Context(), userIDCookie.Value))
		}

		next.ServeHTTP(w, r)
	}
}
