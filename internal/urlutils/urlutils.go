package urlutils

import (
	randMath "math/rand/v2"
	"regexp"

	"github.com/pcristin/urlshortener/internal/storage"
)

func DecodeURL(token string, storage storage.URLStorager) (string, error) {
	return storage.GetURL(token)
}

func generateRandomNumber(a int, b int) int {
	return randMath.IntN(b-a) + a
}

func generateToken(length int) string {
	var letters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	token := make([]byte, length)
	for i := range token {
		token[i] = letters[randMath.IntN(len(letters))]
	}
	return string(token)
}

// Encode URL to a range number from 6 to 9 of random characters
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

// Check the validity of provided URL address
func URLCheck(url string) bool {
	var regExpURLPattern = regexp.MustCompile(`^((http|https):\/\/)?([a-zA-Z0-9.-]+(\.[a-zA-Z]{2,})+)(\/[a-zA-Z0-9-._~:?#@!$&'()*+,;=]*)?$`)
	return regExpURLPattern.MatchString(url)
}

func GenerateToken() string {
	length := generateRandomNumber(6, 10)
	return generateToken(length)
}
