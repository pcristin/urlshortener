package urlutils

import (
	randMath "math/rand/v2"
	"regexp"
	"sync"

	"github.com/pcristin/urlshortener/internal/storage"
)

// Pre-compile the regexp pattern once for better performance
var (
	regExpURLPattern = regexp.MustCompile(`^((http|https):\/\/)?([a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})+)(\/[a-zA-Z0-9-._~:?#@!$&'()*+,;=]*)?$`)
	lettersMu        sync.Mutex
	letters          = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
)

// DecodeURL retrieves the original URL from storage using the provided token
func DecodeURL(token string, storage storage.URLStorager) (string, error) {
	return storage.GetURL(token)
}

func generateRandomNumber(a int, b int) int {
	return randMath.IntN(b-a) + a
}

func generateToken(length int) string {
	token := make([]byte, length)
	lettersMu.Lock()
	for i := range token {
		token[i] = letters[randMath.IntN(len(letters))]
	}
	lettersMu.Unlock()
	return string(token)
}

// EncodeURL shortens a URL to a token with 6-9 random characters
// It first checks if the URL already exists in storage and returns the existing token if found.
// Otherwise, it generates a new token and adds the URL to storage.
func EncodeURL(url string, s storage.URLStorager, userID string) (string, error) {
	// First, check if the URL already exists
	if token, err := s.GetTokenByURL(url); err == nil {
		return token, storage.ErrURLExists
	}

	// If URL doesn't exist, generate a new token and add it
	length := generateRandomNumber(6, 10)
	token := generateToken(length)

	err := s.AddURL(token, url, userID)
	if err != nil {
		// Handle any other errors
		return "", err
	}
	return token, nil
}

// URLCheck validates a URL using a regular expression pattern
// Returns true if the URL is valid, false otherwise
func URLCheck(url string) bool {
	return regExpURLPattern.MatchString(url)
}

// GenerateToken creates a random token with length between 6 and 10 characters
func GenerateToken() string {
	length := generateRandomNumber(6, 10)
	return generateToken(length)
}
